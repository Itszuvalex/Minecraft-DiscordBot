package main

import (
	"flag"
	"fmt"
	"mcdiscord"
	"os"
	"os/signal"
	"syscall"
)

var (
	Token string
	Port  int
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.IntVar(&Port, "p", 3553, "Test Port")
	flag.Parse()
}

func main() {
	dg, err := mcdiscord.New(Token)
	if err != nil {
		fmt.Println("error creating McDiscord, ", err)
		return
	}

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection, ", err)
		return
	}
	defer dg.Close()

	testServer, err := mcdiscord.NewTestServer(Port)
	if err != nil {
		fmt.Println("error creating test server, ", err)
		return
	}
	testServer.Start()

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
