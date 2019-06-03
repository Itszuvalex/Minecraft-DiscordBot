package mcdiscord

type ServerHandler struct {
	servers    map[NetLocation]*McServer
	mainconfig *Config
	discord    *DiscordHandler
}

func NewServerHandler(config *Config, discord *DiscordHandler) *ServerHandler {
	return &ServerHandler{
		servers:    make(map[NetLocation]*McServer),
		mainconfig: config,
		discord:    discord,
	}
}

func (discord *ServerHandler) AddServer(address NetLocation) error {
	server := NewMcServer(address, "192.168.0.1")
	discord.servers[address] = server
	return discord.mainconfig.Write()
}
