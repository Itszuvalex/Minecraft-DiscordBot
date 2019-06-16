package mcdiscord // "github.com/itszuvalex/mcdiscord/pkg/mcdiscord"

import (
	"fmt"

	"github.com/itszuvalex/mcdiscord/pkg/api"
	mydisc "github.com/itszuvalex/mcdiscord/pkg/discord"
	"github.com/itszuvalex/mcdiscord/pkg/server"
)

type McDiscord struct {
	Discord api.IDiscordHandler
	Servers api.IServerHandler
	Config  api.IConfig
}

func New(token string, configFile string) (*McDiscord, error) {
	discord := new(McDiscord)
	discord.Config = NewConfig(configFile)
	discordhandler, err := mydisc.NewDiscordHandler(token, discord.Config)
	if err != nil {
		fmt.Println("Error creating Discord Handler session, ", err)
		return nil, err
	}
	discord.Discord = discordhandler
	discord.Servers = server.NewServerHandler(discord.Config, discord.Discord)
	discord.Discord.SetServerHandler(discord.Servers)

	err = discord.Config.Read()
	if err != nil {
		fmt.Println("Error reading Config,", err)
		return nil, err
	}

	err = discord.Config.Write()
	if err != nil {
		fmt.Println("Error writing Config,", err)
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
