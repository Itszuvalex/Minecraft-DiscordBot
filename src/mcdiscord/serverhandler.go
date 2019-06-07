package mcdiscord

import "fmt"

type ServerHandler struct {
	servers    map[NetLocation]*McServer
	mainconfig *Config
	mcdiscord  *McDiscord
}

func NewServerHandler(config *Config, mcdiscord *McDiscord) *ServerHandler {
	return &ServerHandler{
		servers:    make(map[NetLocation]*McServer),
		mainconfig: config,
		mcdiscord:  mcdiscord,
	}
}

func (discord *ServerHandler) AddServer(address NetLocation) error {
	server := NewMcServer(address, "192.168.0.1", discord.mcdiscord.Discord.Input)
	err := server.net.Connect()
	if err != nil {
		return err
	}
	discord.servers[address] = server
	fmt.Println("Added server at: ", address.Address, ":", address.Port)
	return discord.mainconfig.Write()
}

func (discord *ServerHandler) Close() []error {
	var errors []error
	for _, server := range discord.servers {
		err := server.net.Close()
		if err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func (handler *ServerHandler) SendPacketToAllServers(header Header) {
	for loc, server := range handler.servers {
		fmt.Println("Broadcasting message of type:", header.Type, " to server:", loc.Address)
		server.net.JsonChan <- header
	}
}
