package mcdiscord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type McDiscord struct {
	session *discordgo.Session
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
	discord.AddHandler(messageCreate)

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

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	println("Received message: ", m.Content, ", from user: ", m.Author.Username)

	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, m.Author.Username+": Pong!")
	}
}
