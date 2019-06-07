package mcdiscord

import (
	"fmt"
)

type McDiscord struct {
	Discord *DiscordHandler
	Servers *ServerHandler
	Config  *Config
}

func New(token string, configFile string) (*McDiscord, error) {
	discord := new(McDiscord)
	discord.Config = NewConfig(configFile)
	discordhandler, err := NewDiscordHandler(token, discord)
	if err != nil {
		fmt.Println("Error creating Discord Handler session, ", err)
		return nil, err
	}
	discord.Discord = discordhandler
	discord.Servers = NewServerHandler(discord.Config, discord)

	err = discord.Config.Read()
	if err != nil {
		fmt.Println("Error reading Config,", err)
		return nil, err
	}

	return discord, nil
}

func (discord *McDiscord) Open() error {
	return discord.Discord.Open()
}

func (discord *McDiscord) Close() []error {
	var errors []error
	err := discord.Discord.Close()
	if err != nil {
		errors = append(errors, err)
	}
	servErrors := discord.Servers.Close()
	if servErrors != nil {
		errors = append(errors, servErrors...)
	}
	return errors
}
