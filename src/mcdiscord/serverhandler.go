package mcdiscord

import "fmt"

type ServerHandler struct {
	Servers    map[NetLocation]*McServer
	mainconfig *Config
	mcdiscord  *McDiscord
}

func NewServerHandler(config *Config, mcdiscord *McDiscord) *ServerHandler {
	return &ServerHandler{
		Servers:    make(map[NetLocation]*McServer),
		mainconfig: config,
		mcdiscord:  mcdiscord,
	}
}

func (discord *ServerHandler) AddServer(address NetLocation, name string) error {
	server := NewMcServer(address, "192.168.0.1", name, discord.mcdiscord.Discord.Input)
	err := server.net.Connect()
	if err != nil {
		return err
	}
	discord.Servers[address] = server
	fmt.Println("Added server at: ", address.Address, ":", address.Port)
	return discord.mainconfig.Write()
}

func (discord *ServerHandler) Close() []error {
	var errors []error
	for _, server := range discord.Servers {
		err := server.net.Close()
		if err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func (discord *ServerHandler) RemoveServer(address NetLocation) error {
	server, ok := discord.Servers[address]
	if !ok {
		return fmt.Errorf("Could not find server of address %s:%d", address.Address, address.Port)
	}
	server.net.Close()
	delete(discord.Servers, address)
	return nil
}

func (discord *ServerHandler) RemoveServerByName(name string) error {
	for loc, server := range discord.Servers {
		if server.Name == name {
			server.net.Close()
			delete(discord.Servers, loc)
			return nil
		}
	}
	return fmt.Errorf("Could not find a server of name %s", name)
}

func (handler *ServerHandler) SendPacketToAllServers(header Header) {
	for loc, server := range handler.Servers {
		fmt.Println("Broadcasting message of type:", header.Type, " to server:", loc.Address)
		server.net.JsonChan <- header
	}
}
