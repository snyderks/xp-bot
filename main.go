package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/snyderks/xp-bot/bot"
	"github.com/snyderks/xp-bot/logger"
)

// Variables used for command line parameters
var (
	Token string
)

func init() {
	Token = os.Getenv("XP_BOT_TOKEN")
}

func main() {
	// Make sure the logger has written its buffer.
	logger.Log.Sync()

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		logger.Log.Fatal("Failed to initialize Discord object", err.Error())
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(bot.MessageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		logger.Log.Fatal("Error opening Discord connection,", err.Error())
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")

	u, err := dg.User("@me")
	if err != nil {
		logger.Log.Fatal("Something went wrong when connecting; couldn't get username.",
			err.Error())
		return
	}
	dg.UpdateStatus(1, "Type g! for a graph")
	logger.Log.Info("Connected to Discord as ", u.Username)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}
