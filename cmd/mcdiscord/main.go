package main // github.com/Itszuvalex/Minecraft-DiscordBot

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"mcdiscord"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	ConfigFolder = "config"
)

var (
	Token, TokenFile string
	Port             int
)

func RootPath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

func ConfigPath() string {
	return filepath.Join(RootPath(), ConfigFolder)
}

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&TokenFile, "tf", filepath.Join(ConfigPath(), "Token.txt"), "File containing bot Token")
	flag.IntVar(&Port, "p", 3553, "Test Port")
	flag.Parse()

	if Token == "" && TokenFile == "" {
		fmt.Errorf("Missing token and tokenFile")
	}
}

func main() {
	if Token == "" {
		data, err := ioutil.ReadFile(TokenFile)
		if err != nil {
			fmt.Errorf("Error reading Token File", err)
		}
		Token = strings.TrimSpace(string(data))
	}

	dg, err := mcdiscord.New(Token, filepath.Join(ConfigPath(), "config.json"))
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

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
