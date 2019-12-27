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

	"github.com/akamensky/argparse"
)

var (
	// DBURI is the location of the database.
	DBURI = os.Getenv("XP_BOT_DB_URI")
)

var (
	prefix             = "g!"
	topSwitch          = []string{"top", "t"}
	usersSwitch        = []string{"users", "u", "user", "us"}
	subSwitch          = []string{"sub", "s", "subtraction", "cmp", "compare", "c"}
	lastSwitch         = []string{"last", "l", "days", "d"}
	pup                = "king"
	usage              = "`g! top [number] last [days]`"
	tatsu              = "Tatsumaki"
	leaderboardTrigger = "Guild Score Leaderboards"
)

// Args represents the returned arguments for the command.
type Args struct {
	Usernames []string
	Sub       bool
	Top       int
	Days      int
	King      bool
}

// CreateParser returns a parser that handles the commands for the graph bot.
func CreateParser() *argparse.Parser {
	return argparse.NewParser("print", "Prints provided string to stdout")
}

// MessageCreate will be called every time a new message is created.
// Designed to be the way to handle all interaction.
func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages by the bot
	if m.Author.ID == s.State.User.ID {
		return
	}

	if /* m.Author.Username == "Crouton" && */ strings.HasPrefix(m.Content, prefix) {
		serveGraph(s, m)
	} else if /*m.Author.Username == "Crouton"*/ m.Author.Username == tatsu &&
		m.Author.Bot && strings.Contains(m.Content, leaderboardTrigger) {
		parseNewEntry(s, m)
	}
}

func parseNewEntry(s *discordgo.Session, m *discordgo.MessageCreate) {
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

func serveGraph(s *discordgo.Session, m *discordgo.MessageCreate) {
	// It's a request for the bot to serve a graph up.
	logger.Log.Info("Handler matched on graph request. ",
		" channel: ", m.ChannelID, " sent by: ", m.Author, " content: ", m.Content)
	args, err := ParseGraphArgs(m.Content)
	if err != nil {
		s.ChannelMessageSend(
			m.ChannelID,
			fmt.Sprintf(
				"Your command wasn't recognized. The correct usage is: %s\nError: %s",
				usage, err.Error()))
		return
	}

	var img []byte
	var path string

	// Need to do usernames instead.
	if args.Usernames != nil {
		if args.Sub == true {
			img, path, err = makeUserSubtractionGraph(&args)
		} else {
			img, path, err = makeUserComparisonGraph(&args)
		}
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
	} else {
		img, path, err = makeRankGraph(&args)
		if err != nil {
			if strings.Contains(err.Error(), NeedMoreRecordsError) {
				s.ChannelMessageSend(m.ChannelID, NeedMoreRecordsError)
				return
			}
		}
	}
	logger.Log.Info("Sending image")
	_, err = s.ChannelFileSend(m.ChannelID, path, bytes.NewReader(img))
	if err != nil {
		logger.Log.Error("Failed to send image. ", err.Error(), "imageLen", len(img), "path", path)
	}
}

// ParseGraphArgs attempts to extract the graph arguments from a string.
// Returns only user-friendly errors.
func ParseGraphArgs(s string) (Args, error) {
	splitted := strings.Split(s, " ")
	if len(splitted) > 10 {
		return Args{}, errors.New("")
	}
	if len(splitted) <= 1 {
		// No args. Return the default.
		return Args{nil, false, 10, 0, false}, nil
	}
	args := Args{}

	// Only stuff after the g!
	splitted = splitted[1:]

	// Check for pup
	if splitted[0] == pup {
		args.King = true
		// Remove pup if he's there
		// If statement here is to avoid over-indexing
		// and simply empty the array if "king" is the
		// last thing in the splitted string.
		if len(splitted) > 1 {
			splitted = splitted[1:]
		} else {
			splitted = []string{}
		}
	}

	if len(splitted) == 0 {
		// All done. Return default again!
		args.Top = 10
		return args, nil
	}

	// Check if they specified top.
	args, err := topCheck(args, splitted)

	if err != nil {
		return Args{}, err
	}

	// Check if they specified last X days.
	args, err = lastCheck(args, splitted)

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

	for _, x := range subSwitch {
		if splitted[0] == x {
			if len(splitted) == 1 {
				return Args{},
					errors.New("please pass at least one username as the argument to subtract")
			}
			names := util.StripUsernames(splitted[1:])
			return Args{Sub: true, Usernames: names}, nil
		}
	}

	if err != nil {
		return Args{}, errors.New("your command wasn't recognized")
	}
	return args, nil
}

func topCheck(args Args, splitted []string) (Args, error) {
	// Now we check for top.
	for _, x := range topSwitch {
		if len(splitted) >= 2 && splitted[0] == x {
			top, err := strconv.Atoi(splitted[1])
			if err != nil {
				return Args{},
					errors.New("please pass a number as the argument to top")
			}

			args.Top = top
			return args, nil
		}
	}
	return args, nil
}

func lastCheck(args Args, splitted []string) (Args, error) {
	// Check for last x days
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
				args.Days = last
				return args, nil
			}
		}
	}
	return args, nil
}

func makeRankGraph(args *Args) (img []byte, path string, err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Error("Recovered in makeRankGraph", r)
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

func makeUserComparisonGraph(args *Args) (img []byte, path string, err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Error("Recovered in makeUserComparisonGraph", r)
			err = errors.New("something went wrong when generating the graph")
		}
	}()

	c, err := db.CreateDB(DBURI)
	if err != nil {
		logger.Log.Error("Couldn't connect to database:", err.Error())
		return nil, "", err
	}

	src, _, err := UserLineChart(&c, args)
	if err != nil {
		logger.Log.Error("Failed to construct chart:", err.Error())
		return nil, "", err
	}

	img, path = src.Make()

	return img, path, nil
}

func makeUserSubtractionGraph(args *Args) (img []byte, path string, err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Error("Recovered in makeUserSubtractionGraph", r)
			err = errors.New("something went wrong when generating the graph")
		}
	}()

	c, err := db.CreateDB(DBURI)
	if err != nil {
		logger.Log.Error("Couldn't connect to database:", err.Error())
		return nil, "", err
	}

	src, _, err := SubLineChart(&c, args)
	if err != nil {
		logger.Log.Error("Failed to construct chart:", err.Error())
		return nil, "", err
	}

	img, path = src.Make()

	return img, path, nil
}
