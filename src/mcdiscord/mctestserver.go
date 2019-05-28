package mcdiscord

import "golang.org/x/net/websocket"

type TestServer struct {
	Address string
	Port    int
	Conn    websocket.Conn
}
