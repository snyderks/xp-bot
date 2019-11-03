package bot

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/snyderks/xp-bot/db"

	"github.com/bwmarrin/discordgo"
	"github.com/snyderks/xp-bot/util"
)

var (
	DBURI = os.Getenv("XP_BOT_DB_URI")
)

var (
	prefix             = "g!"
	topSwitch          = []string{"top", "t"}
	usersSwitch        = []string{"users", "u", "user", "us"}
	usage              = "`g! <command> [<args>]\ng! top [number]\ng! users <usernames>`"
	tatsu              = "Tatsumaki"
	leaderboardTrigger = "Guild Score Leaderboards"
)

// Args represents the returned arguments for the command.
type Args struct {
	Usernames []string
	Top       int
}

// MessageCreate will be called every time a new message is created.
// Designed to be the way to handle all interaction.
func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Println("Handler called!")
	// Ignore messages by the bot
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.HasPrefix(m.Content, prefix) {
		// It's a request for the bot to serve a graph up.
		args, err := parseGraphArgs(m.Content)
		if err != nil {
			s.ChannelMessageSend(
				m.ChannelID,
				fmt.Sprintf("%s %s\n%s", "Looks like your request was misformatted: ", err.Error(), usage))
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprint(args))
		}
	} else if m.Author.Username == "Crouton" /*m.Author.Username == tatsu && m.Author.Bot*/ && strings.Contains(m.Content, leaderboardTrigger) {
		// It's a Tatsumaki leaderboard to parse.
		result, err := Parse(m.Content)
		if err != nil {
			log.Print("Couldn't parse Tatsumaki leaderboard:", err.Error())
			return
		}

		c, err := db.CreateDB(DBURI)
		if err != nil {
			log.Print("Couldn't connect to database:", err.Error())
			return
		}

		err = c.AddDay(result)
		if err != nil {
			log.Print("Couldn't log the day:", err.Error())
		}
		err = c.AddPeople(result)
		if err != nil {
			log.Print("Couldn't log people:", err.Error())
		}
	}
}

// parseGraphArgs attempts to extract the graph arguments from a string.
// Returns only user-friendly errors.
func parseGraphArgs(s string) (Args, error) {
	splitted := strings.Split(s, " ")
	if len(splitted) == 1 {
		// No args. Return the default.
		return Args{nil, 10}, nil
	} else {
		splitted = splitted[1:]
		for _, x := range topSwitch {
			if splitted[0] == x {
				i, err := strconv.Atoi(splitted[1])
				if err != nil {
					return Args{}, errors.New("please pass a number as the argument to top")
				}
				return Args{nil, i}, nil
			}
		}
		for _, x := range usersSwitch {
			if splitted[0] == x {
				if len(splitted) == 1 {
					return Args{},
						errors.New("please pass at least one username as the argument to users")
				}
				names := util.StripUsernames(splitted[1:])
				return Args{Usernames: names}, nil
			}
		}
	}
	return Args{}, errors.New("your command wasn't recognized")
}
