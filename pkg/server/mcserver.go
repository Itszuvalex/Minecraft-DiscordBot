package server // "github.com/itszuvalex/mcdiscord/pkg/server"

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/itszuvalex/mcdiscord/pkg/api"
	"golang.org/x/net/websocket"
)

const (
	ConsecutiveErrorMax = 5
)

type mcServerNet struct {
	Location    api.NetLocation
	Origin      string
	Conn        *websocket.Conn
	JsonHandler *api.JsonHandler
	JsonChan    chan api.Header
	stopchan    chan bool
	Status      api.ConnectionStatus
	errcount    int
	mutex       sync.Mutex
}

type mcServer struct {
	net  mcServerNet
	data api.McServerData
	name string
}

type McServerIdentifier struct {
	Location api.NetLocation `json:"loc"`
	Name     string          `json:"name"`
}

func (mcs *mcServer) Location() api.NetLocation {
	return mcs.net.Location
}

func (mcs *mcServer) Name() string {
	return mcs.name
}

func (mcs *mcServer) StartConnectLoop() error {
	return mcs.net.StartConnectLoop()
}

func (mcs *mcServer) Close() error {
	return mcs.net.Close()
}

func (mcs *mcServer) JsonChan() chan api.Header {
	return mcs.net.JsonChan
}

func NewMcServer(location api.NetLocation, origin string, name string, msgchan chan api.MessageWithSender) api.IServer {
	server := &mcServer{
		mcServerNet{
			Location:    location,
			Origin:      origin,
			Conn:        nil,
			JsonHandler: api.NewJsonHandler(),
			JsonChan:    make(chan api.Header, 40),
			stopchan:    make(chan bool, 2),
			Status:      api.Disconnected,
		},
		api.McServerData{Name: name},
		name,
	}
	server.net.JsonHandler.RegisterHandler(api.MessageType, func(obj interface{}) error {
		message, ok := obj.(*api.Message)
		if !ok {
			fmt.Println("MessageHandler passed non *Message obj")
			return errors.New("MessageHandler passed non *Message obj")
		}

		fmt.Println(message.Timestamp, "  :", message.Message)

		msgchan <- api.MessageWithSender{Sender: "", Message: message.Message}

		return nil
	})
	server.net.JsonHandler.RegisterHandler(api.StatusType, func(obj interface{}) error {
		message, ok := obj.(*api.McServerData)
		if !ok {
			fmt.Println("MessageHandler passed non *McServerData obj")
			return errors.New("MessageHandler passed non *McServerData obj")
		}

		fmt.Println(message)
		return nil
	})

	return server
}

func (server *mcServerNet) StartConnectLoop() error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.Status != api.Disconnected {
		return nil
	}

	fmt.Println(fmt.Sprintf("Starting to connect to server %s:%d", server.Location.Address, server.Location.Port))
	server.Status = api.Connecting

	go func() {
		for {
			server.mutex.Lock()
			status := server.Status
			server.mutex.Unlock()

			if status != api.Connecting {
				break
			}

			err := server.Connect()
			if err == nil {
				break
			}

			timer := time.NewTimer(15 * time.Second)
			<-timer.C
		}
	}()

	return nil
}

func (server *mcServerNet) HandleError(err error) error {

	if err != nil {
		fmt.Println(fmt.Sprintf("Encountered error on server %s:%d, ", server.Location.Address, server.Location.Port), err)
		server.mutex.Lock()
		server.errcount++
		errCount := server.errcount
		server.mutex.Unlock()
		if errCount > ConsecutiveErrorMax {
			fmt.Println(fmt.Sprintf("Too many errors encountered, closing and restarting connection to server %s:%d", server.Location.Address, server.Location.Port))
			server.Close()
			server.StartConnectLoop()
		}
	} else {
		server.mutex.Lock()
		server.errcount = 0
		server.mutex.Unlock()
	}
	return err
}

func (server *mcServerNet) Connect() error {
	var err error
	server.Conn, err = websocket.Dial(fmt.Sprintf("ws://%s:%d", server.Location.Address, server.Location.Port), "", fmt.Sprintf("http://%s", server.Origin))
	if err != nil {
		fmt.Println("Error connecting to server, ", err)
		return err
	}

	fmt.Println("Successfully connected to server")
	server.mutex.Lock()
	server.Status = api.Connected
	server.mutex.Unlock()

	go server.handleMessages()
	go server.handleInput()

	t := time.Now()
	message := api.Message{Timestamp: t.Format(time.Stamp), Message: "Discord Bot: Successfully connected to server."}
	var header api.Header
	api.MarshallMessageToHeader(&message, &header)
	server.HandleError(websocket.JSON.Send(server.Conn, &header))

	fmt.Println("Successfully sent bytes to server")

	return nil
}

func (server *mcServerNet) Close() error {
	server.Status = api.Disconnected
	server.errcount = 0
	server.stopchan <- true
	server.stopchan <- true
	if server.Conn != nil {
		return server.Conn.Close()
	}
	return nil
}

func (server *mcServerNet) handleMessages() {
	for {
		select {
		case <-server.stopchan:
			return
		default:
			var header api.Header
			server.HandleError(websocket.JSON.Receive(server.Conn, &header))
			server.JsonHandler.HandleJson(header)
		}
	}
}

func (server *mcServerNet) handleInput() {
	for {
		select {
		case <-server.stopchan:
			return
		case header := <-server.JsonChan:
			server.HandleError(websocket.JSON.Send(server.Conn, &header))
		}
	}
}
