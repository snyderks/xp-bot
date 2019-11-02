package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/snyderks/xp-bot/bot"
)

// Variables used for command line parameters
var (
	Token   string
	Message string
)

func init() {
	flag.StringVar(&Token, "t", "", "Token")
	flag.StringVar(&Message, "m", "", "Message")
	flag.Parse()
}

func botSetup() {
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("GraphBot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func main() {
	result, err := bot.Parse(flag.Arg(0))
	if err != nil {
		print("Bad.")
		return
	}

	counts := make([]string, 0)
	for _, r := range result {
		counts = append(counts, strconv.Itoa(r.XP))
	}

	print("\n" + strings.Join(counts, "\t"))
}
