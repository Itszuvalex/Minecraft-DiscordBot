package mcdiscord

import "golang.org/x/net/websocket"

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

type McServerNet struct {
	Address string
	Port    int
	Conn    websocket.Conn
}

type McServer struct {
	net  McServerNet
	data McServerData
}
