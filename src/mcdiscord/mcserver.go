package mcdiscord

import (
	"errors"
	"fmt"
	"time"

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
	JsonChan    chan Header
	stopchan    chan bool
}

type McServer struct {
	net  McServerNet
	data McServerData
	Name string
}

func NewMcServer(location NetLocation, origin string, name string, msgchan chan MessageWithSender) *McServer {
	server := &McServer{
		McServerNet{
			Location:    location,
			Origin:      origin,
			Conn:        nil,
			JsonHandler: NewJsonHandler(),
			JsonChan:    make(chan Header, 40),
			stopchan:    make(chan bool, 2),
		},
		McServerData{Name: name},
		name,
	}
	server.net.JsonHandler.RegisterHandler(MessageType, func(obj interface{}) error {
		message, ok := obj.(*Message)
		if !ok {
			fmt.Println("MessageHandler passed non *Message obj")
			return errors.New("MessageHandler passed non *Message obj")
		}

		fmt.Println(message.Timestamp, "  ", message.Sender, ":", message.Message)

		msgchan <- MessageWithSender{Sender: message.Sender, Message: message.Message}

		return nil
	})

	return server
}

func (server *McServerNet) Connect() error {
	var err error
	server.Conn, err = websocket.Dial(fmt.Sprintf("ws://%s:%d", server.Location.Address, server.Location.Port), "", fmt.Sprintf("http://%s", server.Origin))
	if err != nil {
		fmt.Println("Error connectiong to server, ", err)
		return err
	}

	fmt.Println("Successfully connected to server")

	go server.handleMessages()
	go server.handleInput()

	t := time.Now()
	message := Message{Timestamp: t.Format(time.Stamp), Sender: "Discord Bot", Message: "Successfully connected to server."}
	var header Header
	MarshallMessageToHeader(&message, &header)
	err = websocket.JSON.Send(server.Conn, &header)
	if err != nil {
		server.Close()
		return err
	}

	fmt.Println("Successfully sent bytes to server")

	return nil
}

func (server *McServerNet) Close() error {
	server.stopchan <- true
	server.stopchan <- true
	return server.Conn.Close()
}

func (server *McServerNet) handleMessages() {
	for {
		select {
		case <-server.stopchan:
			return
		default:
			var header Header
			websocket.JSON.Receive(server.Conn, &header)
			server.JsonHandler.HandleJson(header)
		}
	}
}

func (server *McServerNet) handleInput() {
	for {
		select {
		case <-server.stopchan:
			return
		case header := <-server.JsonChan:
			websocket.JSON.Send(server.Conn, &header)
		}
	}
}
