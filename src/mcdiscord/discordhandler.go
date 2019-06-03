package mcdiscord

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const (
	Emoji_Check string = "✅"
	Emoji_X     string = "❌"
	ConfigKey          = "discord"
	BufferSize         = 100
)

// CommandHandler Type of function that receives new message callbacks from discord
type CommandHandler func(string, *discordgo.MessageCreate) error

type MessageWithSender struct {
	Message string
	Sender  string
}

// DiscordHandler Struct that contains all Discord-related information and handles messages to/from Discord
type DiscordHandler struct {
	session         *discordgo.Session
	commandHandlers map[string]CommandHandler
	config          DiscordHandlerConfig
	Input, Output   chan MessageWithSender
	stopchan        chan bool
	mainconfig      *Config
}

type DiscordHandlerConfig struct {
	ChannelId string `json:"channelId"`
}

// NewDiscordHandler Creates a new DiscordHandler given a bot Token
func NewDiscordHandler(token string, config *Config) (*DiscordHandler, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session, ", err)
		return nil, err
	}
	handler := &DiscordHandler{
		session:         session,
		commandHandlers: make(map[string]CommandHandler),
		config: DiscordHandlerConfig{
			ChannelId: "",
		},
		Input:      make(chan MessageWithSender, BufferSize),
		Output:     make(chan MessageWithSender, BufferSize),
		stopchan:   make(chan bool, 1),
		mainconfig: config,
	}

	// Add handlers
	handler.AddHandler(handler.messageCreate)
	//handler.AddHandler(handler.messageReactionAdd)

	// Add command handlers
	handler.commandHandlers = make(map[string]CommandHandler)
	handler.AddCommandHandler("json", handler.handleJsonTest)
	handler.AddCommandHandler("setchannel", handler.handleSetChannel)

	config.AddReadHandler(ConfigKey, handler.handleConfigRead)
	config.AddWriteHandler(ConfigKey, handler.handleConfigWrite)

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

func (discord *DiscordHandler) Close() error {
	discord.stopchan <- true
	return discord.session.Close()
}

func (discord *DiscordHandler) AddCommandHandler(command string, handler CommandHandler) error {
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

		if m.Content == "ping" {
			s.ChannelMessageSend(m.ChannelID, m.Author.Username+": Pong!")
		}

		if strings.HasPrefix(m.Content, "!") {
			ispace := strings.Index(m.Content, " ")
			var command, data string
			if ispace > 0 {
				command = m.Content[1:ispace]
				data = m.Content[ispace+1:]
			} else {
				command = m.Content[1:]
				data = ""
			}

			println("Received command: ", command, ", from user: ", m.Author.Username, ", with data: ", data)

			handler, ok := discord.commandHandlers[command]
			if !ok {
				println("No handler registered for command:", command)
				return
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
		} else {
			discord.Output <- MessageWithSender{Message: m.Content, Sender: m.Author.Username}
		}
	}()
}

func (discord *DiscordHandler) handleSetChannel(data string, m *discordgo.MessageCreate) error {
	discord.config.ChannelId = m.ChannelID
	return discord.mainconfig.Write()
}

func (discord *DiscordHandler) handleJsonTest(data string, m *discordgo.MessageCreate) error {
	handler := NewJsonHandler()

	handler.RegisterHandler(MessageType, func(message interface{}) error {
		msg, ok := message.(*Message)
		if !ok {
			return errors.New("Received incorrect message type.")
		}

		discord.session.ChannelMessageSend(m.ChannelID, "Received message: "+msg.Message+", at timestamp:"+msg.Timestamp)

		return nil

	})
	err := handler.HandleJson([]byte(data))
	if err != nil {
		fmt.Println("Error handling json,", err)
		return err
	}

	return nil
}

func (discord *DiscordHandler) handleConfigRead(data json.RawMessage) error {
	err := json.Unmarshal(data, &discord.config)
	if err != nil {
		return err
	}
	fmt.Println("Loaded channel id:", discord.config.ChannelId)
	return nil
}

func (discord *DiscordHandler) handleConfigWrite() (json.RawMessage, error) {
	return json.Marshal(&discord.config)
}