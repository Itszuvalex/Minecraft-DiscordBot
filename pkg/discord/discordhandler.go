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

// CommandHandler Type of function that receives new message callbacks from discord
type commandHandler func(string, *discordgo.MessageCreate) error

// DiscordHandler Struct that contains all Discord-related information and handles messages to/from Discord
type DiscordHandler struct {
	session         *discordgo.Session
	commandHandlers map[string]commandHandler
	config          DiscordHandlerConfig
	Input, Output   chan api.MessageWithSender
	stopchan        chan bool
	masterconfig    api.IConfig
	serverhandler   api.IServerHandler
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

type DiscordHandlerConfig struct {
	ChannelId   string `json:"channelId"`
	ControlChar string `json:"controlChar"`
}

// NewDiscordHandler Creates a new DiscordHandler given a bot Token
func NewDiscordHandler(token string, masterconfig api.IConfig) (*DiscordHandler, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session, ", err)
		return nil, err
	}
	handler := &DiscordHandler{
		session:         session,
		commandHandlers: make(map[string]commandHandler),
		config: DiscordHandlerConfig{
			ChannelId:   "",
			ControlChar: "!",
		},
		Input:        make(chan api.MessageWithSender, BufferSize),
		Output:       make(chan api.MessageWithSender, BufferSize),
		stopchan:     make(chan bool, 2),
		masterconfig: masterconfig,
	}

	// Add handlers
	handler.AddHandler(handler.messageCreate)
	//handler.AddHandler(handler.messageReactionAdd)

	// Add command handlers
	handler.commandHandlers = make(map[string]commandHandler)
	handler.AddCommandHandler("json", handler.handleJsonTest)
	handler.AddCommandHandler("setchannel", handler.handleSetChannel)
	handler.AddCommandHandler("ls", handler.handleListServers)
	handler.AddCommandHandler("as", handler.handleAddServer)
	handler.AddCommandHandler("rm", handler.handleRemoveServer)

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

func (discord *DiscordHandler) AddCommandHandler(command string, handler commandHandler) error {
	if _, ok := discord.commandHandlers[command]; ok {
		fmt.Println("Command handler already registered for command:", command)
		return errors.New("Command handler already registered for command:" + command)
	}
	discord.commandHandlers[command] = handler
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
		if discord.isCommandMessage(m) {
			err := discord.handleCommandMessage(s, m)
			if err != nil {

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
		return err
	}
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
	return nil
}

func (discord *DiscordHandler) handleConfigWrite() (json.RawMessage, error) {
	return json.Marshal(&discord.config)
}

func (discord *DiscordHandler) isCommandMessage(m *discordgo.MessageCreate) bool {
	return strings.HasPrefix(m.Content, discord.config.ControlChar)
}

func (discord *DiscordHandler) parseCommandMessage(m *discordgo.MessageCreate) (string, string) {
	var command, data string
	ispace := strings.Index(m.Content, " ")
	if ispace > 0 {
		command = m.Content[1:ispace]
		data = m.Content[ispace+1:]
	} else {
		command = m.Content[1:]
		data = ""
	}
	return command, data
}

func (discord *DiscordHandler) handleCommandMessage(s *discordgo.Session, m *discordgo.MessageCreate) error {
	command, data := discord.parseCommandMessage(m)

	println("Received command: ", command, ", from user: ", m.Author.Username, ", with data: ", data)

	handler, ok := discord.commandHandlers[command]
	if !ok {
		println("No handler registered for command:", command)
		return fmt.Errorf("Unknown command: %s", command)
	}

	err := handler(data, m)
	if err != nil {
		err := s.MessageReactionAdd(m.Message.ChannelID, m.Message.ID, Emoji_X)
		if err != nil {
			fmt.Println("Error adding reaction, ", err)
		}
	} else {
		err := s.MessageReactionAdd(m.Message.ChannelID, m.Message.ID, Emoji_Check)
		if err != nil {
			fmt.Println("Error adding reaction, ", err)
		}
	}
	return nil
}
