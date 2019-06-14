package mcdiscord

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

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

type McServerData struct {
	Memory      int             `json:"memory"`
	MemoryMax   int             `json:"memorymax"`
	Storage     uint64          `json:"storage"`
	StorageMax  uint64          `json:"storagemax"`
	Players     []string        `json:"players"`
	PlayerCount int             `json:"playercount"`
	PlayerMax   int             `json:"playermax"`
	Tps         map[int]float32 `json:"tps"`
	Name        string          `json:"name"`
	Status      string          `json:"status"`
	ActiveTime  int             `json:"activetime"`
}

type NetLocation struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

const (
	ConsecutiveErrorMax = 5
)

type McServerNet struct {
	Location    NetLocation
	Origin      string
	Conn        *websocket.Conn
	JsonHandler *JsonHandler
	JsonChan    chan Header
	stopchan    chan bool
	Status      ConnectionStatus
	errcount    int
	mutex       sync.Mutex
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
			Status:      Disconnected,
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

		fmt.Println(message.Timestamp, "  :", message.Message)

		msgchan <- MessageWithSender{Sender: "", Message: message.Message}

		return nil
	})
	server.net.JsonHandler.RegisterHandler(StatusType, func(obj interface{}) error {
		message, ok := obj.(*McServerData)
		if !ok {
			fmt.Println("MessageHandler passed non *McServerData obj")
			return errors.New("MessageHandler passed non *McServerData obj")
		}

		fmt.Println(message)
		return nil
	})

	return server
}

func (server *McServerNet) StartConnectLoop() error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.Status != Disconnected {
		return nil
	}

	fmt.Println(fmt.Sprintf("Starting to connect to server %s:%d", server.Location.Address, server.Location.Port))
	server.Status = Connecting

	go func() {
		for {
			server.mutex.Lock()
			status := server.Status
			server.mutex.Unlock()

			if status != Connecting {
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

func (server *McServerNet) HandleError(err error) error {

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

func (server *McServerNet) Connect() error {
	var err error
	server.Conn, err = websocket.Dial(fmt.Sprintf("ws://%s:%d", server.Location.Address, server.Location.Port), "", fmt.Sprintf("http://%s", server.Origin))
	if err != nil {
		fmt.Println("Error connecting to server, ", err)
		return err
	}

	fmt.Println("Successfully connected to server")
	server.mutex.Lock()
	server.Status = Connected
	server.mutex.Unlock()

	go server.handleMessages()
	go server.handleInput()

	t := time.Now()
	message := Message{Timestamp: t.Format(time.Stamp), Message: "Discord Bot: Successfully connected to server."}
	var header Header
	MarshallMessageToHeader(&message, &header)
	server.HandleError(websocket.JSON.Send(server.Conn, &header))

	fmt.Println("Successfully sent bytes to server")

	return nil
}

func (server *McServerNet) Close() error {
	server.Status = Disconnected
	server.errcount = 0
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
			server.HandleError(websocket.JSON.Receive(server.Conn, &header))
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
			server.HandleError(websocket.JSON.Send(server.Conn, &header))
		}
	}
}
