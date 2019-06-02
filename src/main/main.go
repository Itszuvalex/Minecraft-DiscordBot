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
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
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

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}
