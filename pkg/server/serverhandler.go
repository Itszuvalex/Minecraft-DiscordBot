package server // "github.com/itszuvalex/mcdiscord/pkg/server"

import (
	"fmt"
	"net"

	"github.com/itszuvalex/mcdiscord/pkg/api"
)

type ServerHandler struct {
	ServerMap      map[api.NetLocation]api.IServer
	mainconfig     api.IConfig
	discordhandler api.IDiscordHandler
}

func NewServerHandler(config api.IConfig, discordhandler api.IDiscordHandler) api.IServerHandler {
	return &ServerHandler{
		ServerMap:      make(map[api.NetLocation]api.IServer),
		mainconfig:     config,
		discordhandler: discordhandler,
	}
}

func (discord *ServerHandler) Servers() map[api.NetLocation]api.IServer {
	return discord.ServerMap
}

func (discord *ServerHandler) AddServer(address api.NetLocation, name string) error {
	server := NewMcServer(address, GetLocalIP(), name, discord.discordhandler.ChatInput())
	err := server.StartConnectLoop()
	if err != nil {
		return err
	}
	discord.ServerMap[address] = server
	fmt.Println("Added server at: ", address.Address, ":", address.Port)
	return discord.mainconfig.Write()
}

// GetLocalIP returns the non loopback local IP of the host
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func (discord *ServerHandler) Close() []error {
	var errors []error
	for _, server := range discord.ServerMap {
		err := server.Close()
		if err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func (discord *ServerHandler) RemoveServer(address api.NetLocation) error {
	server, ok := discord.ServerMap[address]
	if !ok {
		return fmt.Errorf("Could not find server of address %s:%d", address.Address, address.Port)
	}
	server.Close()
	delete(discord.ServerMap, address)
	return nil
}

func (discord *ServerHandler) RemoveServerByName(name string) error {
	for loc, server := range discord.ServerMap {
		if server.Name() == name {
			server.Close()
			delete(discord.ServerMap, loc)
			return nil
		}
	}
	return fmt.Errorf("Could not find a server of name %s", name)
}

func (handler *ServerHandler) SendPacketToAllServers(header api.Header) {
	for loc, server := range handler.ServerMap {
		fmt.Println("Broadcasting message of type:", header.Type, " to server:", loc.Address)
		server.JsonChan() <- header
	}
}

func (handler *ServerHandler) SendPacketToServer(header api.Header, address api.NetLocation) error {
	server, ok := handler.ServerMap[address]
	if !ok {
		return fmt.Errorf("Could not find server of address %s:%d", address.Address, address.Port)
	}
	fmt.Println("Sent message, ", header, "to server: ", address)
	server.JsonChan() <- header
	return nil
}

func (handler *ServerHandler) SendPacketToServerByName(header api.Header, name string) error {
	for _, server := range handler.ServerMap {
		if server.Name() == name {
			fmt.Println("Sent message, ", header, "to server: ", name)
			server.JsonChan() <- header
			return nil
		}
	}
	fmt.Println("Could not find a server of name: ", name)
	return fmt.Errorf("Could not find a server of name %s", name)
}
