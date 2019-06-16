package api // "github.com/itszuvalex/mcdiscord/pkg/api"

type MessageWithSender struct {
	Message string
	Sender  string
}

// State exists because Go doesn't have enums for some reason.
type State int

const (
	// NotRunning indicates something...
	NotRunning State = 0
	// Starting indicates the server is running but not ready for players yet.
	Starting State = 1
	// Running indicates the server is ready for players to connect to.
	Running State = 2
)

type ConnectionStatus int

const (
	Disconnected ConnectionStatus = 0
	Connecting   ConnectionStatus = 1
	Connected    ConnectionStatus = 2
)

type NetLocation struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

type IServerHandler interface {
	AddServer(address NetLocation, name string) error
	RemoveServer(address NetLocation) error
	RemoveServerByName(name string) error
	SendPacketToAllServers(header Header)
	SendPacketToServer(header Header, address NetLocation) error
	SendPacketToServerByName(header Header, name string) error
	Servers() map[NetLocation]IServer
	Close() []error
}

type IServer interface {
	Location() NetLocation
	Name() string
	StartConnectLoop() error
	Close() error
	JsonChan() chan Header
}
