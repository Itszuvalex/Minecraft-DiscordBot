package mcdiscord

import "golang.org/x/net/websocket"

type TestServer struct {
	Address string
	Port    int
	Conn    websocket.Conn
}

func NewTestServer(address string, port int) (*TestServer, error) {
	server := new(TestServer)
	server.Address = address
	server.Port = port
	return server, nil
}
