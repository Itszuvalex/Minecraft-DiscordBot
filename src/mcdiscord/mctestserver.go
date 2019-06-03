package mcdiscord

import (
	"fmt"
	"net/http"

	"golang.org/x/net/websocket"
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
	server.Server = http.Server{Addr: fmt.Sprintf(":%i", server.Port), Handler: mux}
	return server.Server.ListenAndServe()
}

func (server *TestServer) handle(ws *websocket.Conn) {
	fmt.Println("Received connection")
}
