package discord // "github.com/itszuvalex/mcdiscord/pkg/discord"

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/itszuvalex/mcdiscord/pkg/api"
)

const (
	Emoji_Check string = "✅"
	Emoji_X     string = "❌"
	ConfigKey          = "discord"
	BufferSize         = 100
)

type Command struct {
	Name        string
	Usage       string
	Description string
	Handler     commandHandler
}

// CommandHandler Type of function that receives new message callbacks from discord
type commandHandler func(string, *discordgo.MessageCreate) error

// DiscordHandler Struct that contains all Discord-related information and handles messages to/from Discord
type DiscordHandler struct {
	session       *discordgo.Session
	Commands      map[string]Command
	config        DiscordHandlerConfig
	Input, Output chan api.MessageWithSender
	stopchan      chan bool
	masterconfig  api.IConfig
	serverhandler api.IServerHandler
	permhandler   api.IPermHandler
}

func (d *DiscordHandler) ChatInput() chan api.MessageWithSender {
	return d.Input
}

func (d *DiscordHandler) ChatOutput() chan api.MessageWithSender {
	return d.Output
}

func (d *DiscordHandler) SetServerHandler(handler api.IServerHandler) {
	d.serverhandler = handler
}

func (d *DiscordHandler) PermRoot() api.IPermNode {
	return d.permhandler.GetOrAddRoot("mcdiscord")
}

func (d *DiscordHandler) CommandRoot() api.IPermNode {
	return d.PermRoot().GetOrAddPermNode("command")
}

func (d *DiscordHandler) ServerRoot() api.IPermNode {
	return d.PermRoot().GetOrAddPermNode("server")
}

func (d *DiscordHandler) ServerCommandRoot() api.IPermNode {
	return d.ServerRoot().GetOrAddPermNode("command")
}

type DiscordHandlerConfig struct {
	ChannelId   string          `json:"channelId"`
	ControlChar string          `json:"controlChar"`
	PermData    json.RawMessage `json:"perms"`
}

// NewDiscordHandler Creates a new DiscordHandler given a bot Token
func NewDiscordHandler(token string, masterconfig api.IConfig) (*DiscordHandler, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session, ", err)
		return nil, err
	}
	handler := &DiscordHandler{
		session:  session,
		Commands: make(map[string]Command),
		config: DiscordHandlerConfig{
			ChannelId:   "",
			ControlChar: "!",
		},
		Input:        make(chan api.MessageWithSender, BufferSize),
		Output:       make(chan api.MessageWithSender, BufferSize),
		stopchan:     make(chan bool, 2),
		masterconfig: masterconfig,
		permhandler:  NewPermHandler(),
	}

	// Add handlers
	handler.AddHandler(handler.messageCreate)
	//handler.AddHandler(handler.messageReactionAdd)

	// Add command handlers
	handler.Commands = make(map[string]Command)
	handler.AddCommandHandler(Command{"listcommands", "listcommands", "Lists all commands", handler.handleListCommands})
	handler.AddCommandHandler(Command{"json", "json <json>", "Test json input", handler.handleJsonTest})
	handler.AddCommandHandler(Command{"setchannel", "setchannel", "Sets the server interaction channel here.", handler.handleSetChannel})
	handler.AddCommandHandler(Command{"ls", "ls", "List servers", handler.handleListServers})
	handler.AddCommandHandler(Command{"as", "as <address:port> <name>", "Add servers", handler.handleAddServer})
	handler.AddCommandHandler(Command{"rm", "rm (<address:port>|<name>)", "Remove server", handler.handleRemoveServer})

	handler.masterconfig.AddReadHandler(ConfigKey, handler.handleConfigRead)
	handler.masterconfig.AddWriteHandler(ConfigKey, handler.handleConfigWrite)

	return handler, nil
}

func (discord *DiscordHandler) AddHandler(handler interface{}) func() {
	return discord.session.AddHandler(handler)
}

func (discord *DiscordHandler) Open() error {
	err := discord.session.Open()
	if err != nil {
		return err
	}

	go discord.HandleInputChannel()
	go discord.HandleOutputChannel()

	return nil
}

func (discord *DiscordHandler) HandleInputChannel() {
	for {
		select {
		case <-discord.stopchan:
			return
		case i := <-discord.Input:
			if discord.config.ChannelId != "" {
				discord.session.ChannelMessageSend(discord.config.ChannelId, i.Sender+": "+i.Message)
			}
		}
	}
}

func (discord *DiscordHandler) HandleOutputChannel() {
	for {
		select {
		case <-discord.stopchan:
			return
		case o := <-discord.Output:
			command := api.Command{fmt.Sprintf("say %s: %s", o.Sender, o.Message)}
			var header api.Header
			err := api.MarshalCommandToHeader(&command, &header)
			if err != nil {
				fmt.Println("Error marshalling command", err)
			}
			discord.serverhandler.SendPacketToAllServers(header)
		}
	}
}

func (discord *DiscordHandler) Close() error {
	discord.stopchan <- true
	discord.stopchan <- true
	return discord.session.Close()
}

func (discord *DiscordHandler) AddCommandHandler(command Command) error {
	if _, ok := discord.Commands[command.Name]; ok {
		fmt.Println("Command handler already registered for command:", command.Name)
		return errors.New("Command handler already registered for command:" + command.Name)
	}
	node := discord.CommandRoot().GetOrAddPermNode(command.Name)
	node.SetPermDefault(api.Allow)
	discord.Commands[command.Name] = command
	return nil
}

func (discord *DiscordHandler) messageReactionAdd(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	if m.UserID == s.State.User.ID {
		return
	}

	fmt.Println("Seeing reaction added name:", m.Emoji.Name, ", id:", m.Emoji.ID)
}

func (discord *DiscordHandler) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	go func() {
		println("Received message: ", m.Content, ", from user: ", m.Author.Username)
		if discord.isCommandMessage(m.Content) {
			if discord.isCommandMessage(m.Content[1:]) {
				err := discord.handleServerCommandMessage(s, m)
				if err != nil {
					fmt.Println("Error handling server command message", err)
				}
			} else {
				err := discord.handleCommandMessage(s, m)
				if err != nil {
					fmt.Println("Error handling command message", err)
				}
			}
		} else {
			if m.Message.ChannelID == discord.config.ChannelId {
				println("Broadcasting message from user: ", m.Author.Username, ", with message: ", m.Content)
				discord.Output <- api.MessageWithSender{Message: m.Content, Sender: m.Author.Username}
			}
		}
	}()
}

func (discord *DiscordHandler) handleSetChannel(data string, m *discordgo.MessageCreate) error {
	discord.config.ChannelId = m.ChannelID
	return discord.masterconfig.Write()
}

func (discord *DiscordHandler) handleJsonTest(data string, m *discordgo.MessageCreate) error {
	return nil
}

func (discord *DiscordHandler) handleListServers(data string, m *discordgo.MessageCreate) error {
	if discord.config.ChannelId != m.Message.ChannelID {
		fmt.Println("Wrong channel")
		return errors.New("Wrong channel")
	}
	var serverfields []*discordgo.MessageEmbedField
	for _, server := range discord.serverhandler.Servers() {
		serverfields = append(serverfields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("%s", server.Name()),
			Value: fmt.Sprintf("%s:%d", server.Location().Address, server.Location().Port),
		})
	}
	embed := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       0x00ff00,
		Description: "Servers connected to discord.",
		Fields:      serverfields,
		Timestamp:   time.Now().Format(time.RFC3339),
		Title:       "List Servers",
	}
	_, err := discord.session.ChannelMessageSendEmbed(discord.config.ChannelId, embed)
	if err != nil {
		fmt.Println("Error embedding message", err)
		return err
	}
	return nil
}

func (discord *DiscordHandler) handleListCommands(data string, m *discordgo.MessageCreate) error {
	var commandfields []*discordgo.MessageEmbedField
	for _, command := range discord.Commands {
		commandfields = append(commandfields, &discordgo.MessageEmbedField{
			Name:   command.Name,
			Value:  fmt.Sprintf("Usage: %s\n%s", discord.config.ControlChar+command.Usage, command.Description),
			Inline: true,
		})
	}
	embed := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       0x00ff00,
		Description: "Bot discord commands",
		Fields:      commandfields,
		Timestamp:   time.Now().Format(time.RFC3339),
		Title:       "List Commands",
	}
	_, err := discord.session.ChannelMessageSendEmbed(discord.config.ChannelId, embed)
	if err != nil {
		fmt.Println("Error embedding message", err)
		return err
	}
	return nil

	return nil
}

func (discord *DiscordHandler) handleAddServer(data string, m *discordgo.MessageCreate) error {
	if discord.config.ChannelId != m.Message.ChannelID {
		return errors.New("Wrong channel")
	}
	args := strings.Split(data, " ")
	if len(args) < 2 {
		fmt.Println("Add server needs server name.")
		return errors.New("Add server needs args {ip:port} {name}")
	}

	location, err := api.ParseNetLocation(args[0])
	if err != nil {
		fmt.Println("Add server could not parse NetLocation:", err)
		return err
	}

	name := strings.Join(args[1:], " ")
	err = discord.serverhandler.AddServer(*location, name)
	if err != nil {
		fmt.Println("Add server could not add server", err)
		return err
	}
	return nil
}
func (discord *DiscordHandler) handleRemoveServer(data string, m *discordgo.MessageCreate) error {
	if discord.config.ChannelId != m.Message.ChannelID {
		return errors.New("Wrong channel")
	}
	if strings.Contains(data, ":") {
		location, err := api.ParseNetLocation(data)
		if err != nil {
			return err
		}
		return discord.serverhandler.RemoveServer(*location)
	} else {
		return discord.serverhandler.RemoveServerByName(data)
	}
}

func (discord *DiscordHandler) handleConfigRead(data json.RawMessage) error {
	err := json.Unmarshal(data, &discord.config)
	if err != nil {
		return err
	}
	return discord.permhandler.ReadJson(discord.config.PermData)
}

func (discord *DiscordHandler) handleConfigWrite() (json.RawMessage, error) {
	permdata, err := discord.permhandler.WriteJson()
	if err != nil {
		return nil, err
	}
	discord.config.PermData = permdata
	return json.Marshal(&discord.config)
}

func (discord *DiscordHandler) isCommandMessage(message string) bool {
	return strings.HasPrefix(message, discord.config.ControlChar)
}

func (discord *DiscordHandler) parseCommandMessage(message string) (string, string) {
	var command, data string
	ispace := strings.Index(strings.TrimSpace(message), " ")
	if ispace > 0 {
		command = message[:ispace]
		data = message[ispace:]
	} else {
		command = strings.TrimSpace(message)
		data = ""
	}
	return strings.TrimSpace(command), strings.TrimSpace(data)
}

func (discord *DiscordHandler) handleCommandMessage(s *discordgo.Session, m *discordgo.MessageCreate) error {
	command, data := discord.parseCommandMessage(m.Content[1:])

	println("Received command: ", command, ", from user: ", m.Author.Username, ", with data: ", data)

	commandNode, err := discord.CommandRoot().GetPermNode(command)
	if err != nil {
		return err
	}
	userid := m.Author.ID
	member, err := s.GuildMember(m.GuildID, userid)
	if err != nil {
		return err
	}

	fmt.Println("Command node path:", commandNode.FullName())

	check, err := discord.permhandler.UserWithRolesAllowed(commandNode.FullName(), m.GuildID, userid, member.Roles)
	if err != nil {
		fmt.Println("Error on detecting permissions:", err)
		return err
	}

	if !check.Allowed {
		fmt.Println("Perms not allowed at path:", check.Path)
		return errors.New("Perms not allowed")
	}

	cmd, ok := discord.Commands[command]
	if !ok {
		println("No handler registered for command:", command)
		return fmt.Errorf("Unknown command: %s", command)
	}

	err = cmd.Handler(data, m)
	if err != nil {
		err := s.MessageReactionAdd(m.Message.ChannelID, m.Message.ID, Emoji_X)
		if err != nil {
			fmt.Println("Error adding reaction, ", err)
			return err
		}
	} else {
		err := s.MessageReactionAdd(m.Message.ChannelID, m.Message.ID, Emoji_Check)
		if err != nil {
			fmt.Println("Error adding reaction, ", err)
			return err
		}
	}
	return nil
}

func (discord *DiscordHandler) handleServerCommandMessage(s *discordgo.Session, m *discordgo.MessageCreate) error {
	server, cmddata := discord.parseCommandMessage(m.Content[2:])

	println("Received command to server: ", server, ", from user: ", m.Author.Username, ", with id: ", m.Author.ID, ", with data: ", cmddata)

	command := api.Command{Command: cmddata}
	var header api.Header
	err := api.MarshalCommandToHeader(&command, &header)
	if err != nil {
		fmt.Println("Error marshalling command", err)
	} else {
		loc, inerr := api.ParseNetLocation(server)
		if inerr == nil {
			err = discord.serverhandler.SendPacketToServer(header, *loc)
		} else {
			err = discord.serverhandler.SendPacketToServerByName(header, server)
		}
	}

	if err != nil {
		err := s.MessageReactionAdd(m.Message.ChannelID, m.Message.ID, Emoji_X)
		if err != nil {
			fmt.Println("Error adding reaction, ", err)
			return err
		}
	} else {
		err := s.MessageReactionAdd(m.Message.ChannelID, m.Message.ID, Emoji_Check)
		if err != nil {
			fmt.Println("Error adding reaction, ", err)
			return err
		}
	}
	return nil
}
