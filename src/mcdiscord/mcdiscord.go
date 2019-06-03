package mcdiscord

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type CommandHandler func(string, *discordgo.MessageCreate) error

type McDiscord struct {
	session         *discordgo.Session
	commandHandlers map[string]CommandHandler
	servers         map[NetLocation]*McServer
}

func New(token string) (*McDiscord, error) {
	discord := new(McDiscord)
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session, ", err)
		return nil, err
	}
	discord.session = session

	// Add handlers
	discord.AddHandler(discord.messageCreate)

	// Add command handlers
	discord.commandHandlers = make(map[string]CommandHandler)
	discord.AddCommandHandler("json", discord.handleJsonTest)

	discord.servers = make(map[NetLocation]*McServer)

	return discord, nil
}

func (discord *McDiscord) AddHandler(handler interface{}) func() {
	return discord.session.AddHandler(handler)
}

func (discord *McDiscord) Open() error {
	return discord.session.Open()
}

func (discord *McDiscord) Close() error {
	return discord.session.Close()
}

func (discord *McDiscord) AddCommandHandler(command string, handler CommandHandler) error {
	if _, ok := discord.commandHandlers[command]; ok {
		fmt.Println("Command handler already registered for command:", command)
		return errors.New("Command handler already registered for command:" + command)
	}
	discord.commandHandlers[command] = handler
	return nil
}

func (discord *McDiscord) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

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

		handler(data, m)
	}
}

func (discord *McDiscord) handleJsonTest(data string, m *discordgo.MessageCreate) error {
	handler := NewJsonHandler()

	handler.RegisterHandler(MessageType, func(message interface{}) error {
		msg, ok := message.(*Message)
		if !ok {
			return errors.New("Received incorrect message type.")
		}

		discord.session.ChannelMessageSend(m.ChannelID, "Received message: "+msg.Message+", at timestamp:"+msg.Timestamp)

		return nil

	})
	handler.HandleJson([]byte(data))

	return nil
}

func (discord *McDiscord) AddServer(address NetLocation) error {
	server := NewMcServer(address.Address, address.Port, "192.168.0.1")
	discord.servers[address] = server
	return nil
}
