package bot

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/snyderks/xp-bot/chart"
	"github.com/snyderks/xp-bot/db"

	"github.com/bwmarrin/discordgo"
	"github.com/snyderks/xp-bot/logger"
	"github.com/snyderks/xp-bot/util"
)

var (
	// DBURI is the location of the database.
	DBURI = os.Getenv("XP_BOT_DB_URI")
)

var (
	prefix             = "g!"
	topSwitch          = []string{"top", "t"}
	usersSwitch        = []string{"users", "u", "user", "us"}
	lastSwitch         = []string{"last", "l", "days", "d"}
	pup                = "king"
	usage              = "`g! top [number] last [days]`"
	tatsu              = "Tatsumaki"
	leaderboardTrigger = "Guild Score Leaderboards"
)

// Args represents the returned arguments for the command.
type Args struct {
	Usernames []string
	Top       int
	Days      int
	King      bool
}

// MessageCreate will be called every time a new message is created.
// Designed to be the way to handle all interaction.
func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages by the bot
	if m.Author.ID == s.State.User.ID {
		return
	}

	if /* m.Author.Username == "Crouton" && */ strings.HasPrefix(m.Content, prefix) {
		// It's a request for the bot to serve a graph up.
		logger.Log.Info("Handler matched on graph request. ",
			"channel: ", m.ChannelID, "sent by: ", m.Author, "content: ", m.Content)
		args, err := parseGraphArgs(m.Content)
		if args.Usernames != nil {
			// Not supported yet.
			s.ChannelMessageSend(m.ChannelID, usage)
			return
		}
		if err != nil {
			s.ChannelMessageSend(
				m.ChannelID,
				fmt.Sprintf("%s %s", "Your command wasn't recognized. The correct usage is:", usage))
			return
		}
		img, path, err := makeGraph(&args)
		if err != nil {
			if strings.Contains(err.Error(), NeedMoreRecordsError) {
				s.ChannelMessageSend(m.ChannelID, NeedMoreRecordsError)
				return
			}
		}
		logger.Log.Info("Sending image")
		_, err = s.ChannelFileSend(m.ChannelID, path, bytes.NewReader(img))
		if err != nil {
			logger.Log.Error("Failed to send image. ", err.Error(), "imageLen", len(img), "path", path)
		}
	} else if /*m.Author.Username == "Crouton"*/ m.Author.Username == tatsu && m.Author.Bot && strings.Contains(m.Content, leaderboardTrigger) {
		// It's a Tatsumaki leaderboard to parse.
		result, err := Parse(m.Content)
		if err != nil {
			logger.Log.Error("Couldn't parse Tatsumaki leaderboard: ", err.Error())
			return
		}

		c, err := db.CreateDB(DBURI)
		if err != nil {
			logger.Log.Error("Couldn't connect to database: ", err.Error())
			return
		}

		err = c.AddDay(result)
		if err != nil {
			logger.Log.Error("Couldn't log the day: ", err.Error())
		}
		err = c.AddPeople(result)
		if err != nil {
			logger.Log.Error("Couldn't log people: ", err.Error())
		}
	}
}

// parseGraphArgs attempts to extract the graph arguments from a string.
// Returns only user-friendly errors.
func parseGraphArgs(s string) (Args, error) {
	splitted := strings.Split(s, " ")
	if len(splitted) == 1 {
		// No args. Return the default.
		return Args{nil, 10, 0, false}, nil
	}
	args := Args{}

	splitted = splitted[1:]

	if splitted[len(splitted)-1] == pup {
		args.King = true
		splitted = splitted[1:]
	}
	if len(splitted) == 0 {
		// All done. Return default again!
		args.Top = 10
		return args, nil
	}

	for _, x := range topSwitch {
		if len(splitted) >= 2 && splitted[0] == x {
			top, err := strconv.Atoi(splitted[1])
			if err != nil {
				return Args{},
					errors.New("please pass a number as the argument to top")
			}

			// Check for last x days
			// TODO: move this into its own function for use with usernames.
			if len(splitted) >= 4 {
				for _, x := range lastSwitch {
					if splitted[2] == x {
						last, err := strconv.Atoi(splitted[3])
						if err != nil {
							return Args{},
								errors.New("please pass a number as the argument to last")
						}
						if last >= chart.GlobalChartConfig.DaysLimit {
							return Args{},
								errors.New("too many days requested. Please request fewer days")
						}
						if last <= 0 {
							return Args{},
								errors.New("nice try")
						}
						args.Top = top
						args.Days = last
						return args, nil
					}
				}
			}

			args.Top = top
			return args, nil
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
	return Args{}, errors.New("your command wasn't recognized")
}

func makeGraph(args *Args) (img []byte, path string, err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Error("Recovered in makeGraph", r)
			err = errors.New("something went wrong when generating the graph")
		}
	}()

	c, err := db.CreateDB(DBURI)
	if err != nil {
		logger.Log.Error("Couldn't connect to database:", err.Error())
		return nil, "", err
	}

	src, _, err := RankLineChart(&c, args)
	if err != nil {
		logger.Log.Error("Failed to construct chart:", err.Error())
		return nil, "", err
	}

	img, path = src.Make()

	return img, path, nil
}
