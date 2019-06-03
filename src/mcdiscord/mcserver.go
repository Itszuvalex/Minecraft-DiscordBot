package mcdiscord

import (
	"fmt"

	"golang.org/x/net/websocket"
)

type Status int

const (
	Disconnected Status = iota
	Stopped
	Starting
	Running
)

type McServerData struct {
	Memory     int             `json:"memory"`
	MemoryMax  int             `json:"memorymax"`
	Storage    int             `json:"storage"`
	StorageMax int             `json:"storagemax"`
	Players    []string        `json:"players"`
	PlayerMax  int             `json:"playermax"`
	Tps        map[int]float32 `json:"tps"`
	Name       string          `json:"name"`
	Status     Status          `json:"status"`
	ActiveTime int             `json:"activetime"`
}

type NetLocation struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

type McServerNet struct {
	Location    NetLocation
	Origin      string
	Conn        *websocket.Conn
	JsonHandler *JsonHandler
}

type McServer struct {
	net  McServerNet
	data McServerData
}

func NewMcServer(location NetLocation, origin string) *McServer {
	return &McServer{
		McServerNet{
			Location:    location,
			Origin:      origin,
			Conn:        nil,
			JsonHandler: NewJsonHandler(),
		},
		McServerData{},
	}
}

func (server *McServerNet) Connect() error {
	var err error
	server.Conn, err = websocket.Dial(fmt.Sprintf("ws://%s:%i", server.Location.Address, server.Location.Port), "", server.Origin)
	return err
}
