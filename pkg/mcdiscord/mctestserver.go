package mcdiscord // "github.com/itszuvalex/mcdiscord/pkg/mcdiscord"

import (
	"fmt"
	"net/http"

	"golang.org/x/net/websocket"
	"github.com/itszuvalex/mcdiscord/pkg/api"
)

type TestServer struct {
	Port   int
	Server http.Server
}

func NewTestServer(port int) (*TestServer, error) {
	server := new(TestServer)
	server.Port = port
	return server, nil
}

func (server *TestServer) Start() error {
	mux := http.NewServeMux()
	mux.Handle("/", websocket.Handler(server.handle))
	server.Server = http.Server{Addr: fmt.Sprintf(":%d", server.Port), Handler: mux}
	go server.Server.ListenAndServe()

	return nil
}

func (server *TestServer) Close() error {
	return server.Server.Close()
}

func (server *TestServer) handle(ws *websocket.Conn) {
	fmt.Println("Received connection")
	for {
		var data api.Header
		err := websocket.JSON.Receive(ws, &data)
		if err != nil {
			return
		}
		fmt.Printf("Received Header:%s from connection\n", data.Type)
		err = websocket.JSON.Send(ws, &data)
		if err != nil {
			return
		}
	}
}
