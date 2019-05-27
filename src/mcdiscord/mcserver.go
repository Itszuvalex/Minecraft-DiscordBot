package mcdiscord

import "golang.org/x/net/websocket"

type Status int

const (
	Disconnected Status = iota
	Stopped
	Starting
	Running
)

type TpsDim struct {
	Dim int
	Tps float32
}

type McServerRuntime struct {
	Memory     int
	MemoryMax  int
	Storage    int
	StorageMax int
	Players    []string
	PlayerMax  int
	Tps        []TpsDim
}

type McServerData struct {
	Name       string
	Status     Status
	ActiveTime int
	Runtime    McServerRuntime
}

type McServerNet struct {
	Address string
	Port    int
	Conn    websocket.Conn
}

type McServer struct {
	net  McServerNet
	data McServerData
}
