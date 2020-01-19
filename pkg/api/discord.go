package api // "github.com/itszuvalex/mcdiscord/pkg/api"

import (
	"github.com/bwmarrin/discordgo"
)

type IDiscordHandler interface {
	ChatInput() chan MessageWithSender
	ChatOutput() chan MessageWithSender
	SetServerHandler(handler IServerHandler)
	Open() error
	Close() error
	Session() *discordgo.Session
}
