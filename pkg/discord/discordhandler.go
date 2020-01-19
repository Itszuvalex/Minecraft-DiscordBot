package discord // "github.com/itszuvalex/mcdiscord/pkg/discord"

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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

type ServerCommand struct {
	Name        string
	Usage       string
	Description string
	Handler     commandHandler
}

// CommandHandler Type of function that receives new message callbacks from discord
type commandHandler func(string, *discordgo.Session, *discordgo.MessageCreate) error

// DiscordHandler Struct that contains all Discord-related information and handles messages to/from Discord
type DiscordHandler struct {
	session        *discordgo.Session
	Commands       map[string]Command
	ServerCommands map[string]ServerCommand
	config         DiscordHandlerConfig
	Input, Output  chan api.MessageWithSender
	stopchan       chan bool
	masterconfig   api.IConfig
	serverhandler  api.IServerHandler
	permhandler    api.IPermHandler
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
	ChannelId       string          `json:"channelId"`
	StatusChannelId string          `json:"statuschannelId"`
	ControlChar     string          `json:"controlChar"`
	PermData        json.RawMessage `json:"perms"`
}

// NewDiscordHandler Creates a new DiscordHandler given a bot Token
func NewDiscordHandler(token string, masterconfig api.IConfig) (*DiscordHandler, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session, ", err)
		return nil, err
	}
	handler := &DiscordHandler{
		session:        session,
		Commands:       make(map[string]Command),
		ServerCommands: make(map[string]ServerCommand),
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
	handler.AddCommandHandler(api.Allow, Command{"listcommands", "listcommands", "Lists all commands", handler.handleListCommands})
	handler.AddCommandHandler(api.Allow, Command{"json", "json <json>", "Test json input", handler.handleJsonTest})
	handler.AddCommandHandler(api.Allow, Command{"setchannel", "setchannel", "Sets the server interaction channel here.", handler.handleSetChannel})
	handler.AddCommandHandler(api.Allow, Command{"setstatuschannel", "setstatuschannel", "Sets the server status channel here.", handler.handleSetStatusChannel})
	handler.AddCommandHandler(api.Allow, Command{"ls", "ls", "List servers", handler.handleListServers})
	handler.AddCommandHandler(api.Allow, Command{"as", "as <address:port> <name>", "Add servers", handler.handleAddServer})
	handler.AddCommandHandler(api.Allow, Command{"rm", "rm (<address:port>|<name>)", "Remove server", handler.handleRemoveServer})

	handler.AddCommandHandler(api.Allow, Command{"lp", "lp (filterstring)", "List Permission Nodes with optional filter", handler.handleListPermissions})
	handler.AddCommandHandler(api.Allow, Command{"lpn", "lpn <perm.node>", "List Permissions for Node", handler.handleListPermissionsForNode})
	handler.AddCommandHandler(api.Block, Command{"apr", "apr <perm.node> <mentionrole> <yes|no>", "Add perm for mentioned role with value.", handler.handleAddPermRole})
	handler.AddCommandHandler(api.Block, Command{"apu", "apu <perm.node> <mentionuser> <yes|no>", "Add perm for mentioned user with value.", handler.handleAddPermUser})
	handler.AddCommandHandler(api.Block, Command{"rpr", "rpr <perm.node> <rolename>", "Remove perm for mentioned role", handler.handleRemovePermRole})
	handler.AddCommandHandler(api.Block, Command{"rpu", "rpu <perm.node> <username>", "Remove perm for mentioned user", handler.handleRemovePermUser})

	handler.AddCommandHandler(api.Allow, Command{"lsc", "lsc", "Lists all server commands", handler.handleListServerCommands})

	handler.AddServerCommandHandler(ServerCommand{"start", "start", "Starts the server if it's not running", handler.handleSendServerCommandMessage})
	handler.AddServerCommandHandler(ServerCommand{"stop", "stop", "Stop the server", handler.handleSendServerCommandMessage})
	handler.AddServerCommandHandler(ServerCommand{"kill", "kill", "Force kill the server", handler.handleSendServerCommandMessage})
	handler.AddServerCommandHandler(ServerCommand{"reboot", "reboot", "Stops and reboots the server", handler.handleSendServerCommandMessage})
	handler.AddServerCommandHandler(ServerCommand{"forcereboot", "forcereboot", "Force kills the server and reboots.", handler.handleSendServerCommandMessage})
	handler.AddServerCommandHandler(ServerCommand{"save", "save", "Saves the server.", handler.handleSendServerCommandMessage})
	handler.ServerCommandRoot().GetOrAddPermNode("unknown").SetPermDefault(api.Block)

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

func (discord *DiscordHandler) Session() *discordgo.Session {
	return discord.session
}

func (discord *DiscordHandler) AddCommandHandler(perm api.PermDefault, command Command) error {
	if _, ok := discord.Commands[command.Name]; ok {
		fmt.Println("Command handler already registered for command:", command.Name)
		return errors.New("Command handler already registered for command:" + command.Name)
	}
	node := discord.CommandRoot().GetOrAddPermNode(command.Name)
	node.SetPermDefault(perm)
	discord.Commands[command.Name] = command
	return nil
}

func (discord *DiscordHandler) AddServerCommandHandler(command ServerCommand) error {
	if _, ok := discord.ServerCommands[command.Name]; ok {
		fmt.Println("ServerCommand handler already registered for command:", command.Name)
		return errors.New("ServerCommand handler already registered for command:" + command.Name)
	}
	node := discord.ServerCommandRoot().GetOrAddPermNode(command.Name)
	node.SetPermDefault(api.Block)
	discord.ServerCommands[command.Name] = command
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
		println("Received message: ", m.Content, ", from user: ", m.Author.Username, ", id:", m.Author.ID, ", in channel:", m.ChannelID, ", in guild: ", m.GuildID)
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
		data = message[ispace+1:]
	} else {
		command = strings.TrimSpace(message)
		data = ""
	}
	return strings.TrimSpace(command), strings.TrimSpace(data)
}

func (discord *DiscordHandler) handleCommandMessage(s *discordgo.Session, m *discordgo.MessageCreate) error {
	command, data := discord.parseCommandMessage(m.Content[1:])

	println("Received command: ", command, ", from user: ", m.Author.Username, ", with data: ", data)

	userid := m.Author.ID

	commandNode, err := discord.CommandRoot().GetPermNode(command)
	if err != nil {
		return err
	}

	guild, err := s.Guild(m.GuildID)
	if err != nil {
		return err
	}

	if guild.OwnerID != userid {
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
	} else {
		fmt.Println("Server owners can do everything.")
	}

	cmd, ok := discord.Commands[command]
	if !ok {
		println("No handler registered for command:", command)
		return fmt.Errorf("Unknown command: %s", command)
	}

	err = cmd.Handler(data, s, m)
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
	command, _ := discord.parseCommandMessage(cmddata)

	println("Received command:", command, " to server: ", server, ", from user: ", m.Author.Username, ", with id: ", m.Author.ID, ", with data: ", cmddata)

	userid := m.Author.ID

	var commandNode api.IPermNode
	var err error
	cmd, knownCommand := discord.ServerCommands[command]
	if !knownCommand {
		fmt.Println("Unknown command:", command)
		commandNode, err = discord.ServerCommandRoot().GetPermNode("unknown")
	} else {
		commandNode, err = discord.ServerCommandRoot().GetPermNode(command)
	}
	if err != nil {
		fmt.Println("Error getting perm node")
		return err
	}

	guild, err := s.Guild(m.GuildID)
	if err != nil {
		return err
	}

	if guild.OwnerID != userid {
		member, err := s.GuildMember(m.GuildID, userid)
		if err != nil {
			return err
		}

		fmt.Println("ServerCommand node path:", commandNode.FullName())

		check, err := discord.permhandler.UserWithRolesAllowed(commandNode.FullName(), m.GuildID, userid, member.Roles)
		if err != nil {
			fmt.Println("Error on detecting permissions:", err)
			return err
		}

		if !check.Allowed {
			fmt.Println("Perms not allowed at path:", check.Path)
			return errors.New("Perms not allowed")
		}
	} else {
		fmt.Println("Server owners can do everything.")
	}

	if !knownCommand {
		err = discord.handleSendServerCommandMessage(cmddata, s, m)
	} else {
		err = cmd.Handler(cmddata, s, m)
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

func (discord *DiscordHandler) handleSendServerCommandMessage(s string, session *discordgo.Session, m *discordgo.MessageCreate) error {
	server, cmddata := discord.parseCommandMessage(m.Content[2:])

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
	return err
}
