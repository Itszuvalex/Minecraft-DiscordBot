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

/*
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session, ", err)
		return
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection, ", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}
}

*/
