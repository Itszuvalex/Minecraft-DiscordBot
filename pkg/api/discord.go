package api // "github.com/itszuvalex/mcdiscord/pkg/api"

type IDiscordHandler interface {
	ChatInput() chan MessageWithSender
	ChatOutput() chan MessageWithSender
	SetServerHandler(handler IServerHandler)
	Open() error
	Close() error
}
