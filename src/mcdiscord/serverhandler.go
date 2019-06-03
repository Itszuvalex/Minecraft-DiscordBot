package mcdiscord

type ServerHandler struct {
	servers    map[NetLocation]*McServer
	mainconfig *Config
}

func NewServerHandler(config *Config) *ServerHandler {
	return &ServerHandler{
		servers:    make(map[NetLocation]*McServer),
		mainconfig: config,
	}
}

func (discord *ServerHandler) AddServer(address NetLocation) error {
	server := NewMcServer(address.Address, address.Port, "192.168.0.1")
	discord.servers[address] = server
	return discord.mainconfig.Write()
}
