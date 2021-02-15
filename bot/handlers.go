package bot

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/snyderks/xp-bot/db"

	"github.com/bwmarrin/discordgo"
	"github.com/snyderks/xp-bot/chart"
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
	helpSwitch         = []string{"help", "h", "ahh", "?", "usage"}
	statsSwitch        = []string{"stats", "stat"}
	manualUpdate       = []string{"update"}
	pup                = "king"
	usage              = "`g! help | top [number] | last [days] | compare [user 1, user 2] | users [usernames]`"
	tatsu              = "172002275412279296"
	guildID            = "99199781900910592"
	croot              = "99201752280096768"
	leaderboardTrigger = "Guild Score Leaderboards"
)

// Args represents the returned arguments for the command.
type Args struct {
	Usernames []string
	Sub       bool
	Stats     bool
	Top       int
	Days      int
	King      bool
	Update    bool
	Help      bool
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
	}
}

// AppendRankings retrieves current server rankings and places them in the DB.
func AppendRankings() error {
	ranks, err := GetRankings(guildID)
	if err != nil {
		logger.Log.Error("Couldn't get new rankings: ", err.Error())
		return err
	}

	c, err := db.CreateDB(DBURI)
	if err != nil {
		logger.Log.Error("Couldn't connect to database: ", err.Error())
		return err
	}

	peopleMap, err := RankingsToPeople(ranks)
	if err != nil {
		logger.Log.Error("Conversion failed: ", err.Error())
		return err
	}

	err = c.AddDay(peopleMap)
	if err != nil {
		logger.Log.Error("DB insert failed: ", err.Error())
		return err
	}

	err = c.AddPeople(peopleMap)
	if err != nil {
		logger.Log.Error("DB insert failed: ", err.Error())
		return err
	}

	return nil
}

func serveGraph(s *discordgo.Session, m *discordgo.MessageCreate) {
	// It's a request for the bot to serve a graph up.
	logger.Log.Info("Handler matched on graph request. ",
		" channel: ", m.ChannelID, " sent by: ", m.Author, " content: ", m.Content)
	args, err := ParseGraphArgs(m.Content, m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(
			m.ChannelID,
			fmt.Sprintf(
				"Your command wasn't recognized. \nThe correct usage is: %s",
				usage))
		return
	}

	if args.Update {
		AppendRankings()
	}
	if args.Help {
		s.ChannelMessageSend(m.ChannelID, usage)
		return
	}

	// For now, putting this here.
	// TODO: pull out the ParseGraphArgs above into the calling function
	// and make this called from there as well.
	if args.Stats {
		sendStatsMessage(s, m, args)
		return
	}

	var img []byte
	var path string

	// Need to do usernames instead.
	if args.Usernames != nil {
		if args.Sub == true {
			img, path, err = makeUserSubtractionGraph(&args, s)
		} else {
			img, path, err = makeUserComparisonGraph(&args, s)
		}
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
	} else {
		img, path, err = makeRankGraph(&args, s)
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

func sendStatsMessage(s *discordgo.Session, m *discordgo.MessageCreate, args Args) {
	// Get the XP for the person.
	c, err := db.CreateDB(DBURI)
	if err != nil {
		logger.Log.Error("Couldn't connect to database:", err.Error())
		s.ChannelMessageSend(m.ChannelID,
			"Failed to create the stats lookup. "+
				"(Something is probs wrong with the DB.)")
		return
	}

	if args.Usernames == nil {
		args.Usernames = []string{m.Author.Username}
	}

	// Looking to get lots of time.
	notFound, people, err := c.ReadPeople(args.Usernames, 365)

	if err != nil {
		logger.Log.Error("Couldn't get the people for some reason:",
			err.Error())
		s.ChannelMessageSend(m.ChannelID, "Failed to create the stats lookup. "+
			"(Something is probs wrong with the DB.)")
		return
	}

	names, err := GetNicknames(args.Usernames, s)
	if err != nil {
		logger.Log.Error("One or more people didn't exist at all.",
			err.Error())
		s.ChannelMessageSend(m.ChannelID,
			"One or more users do not exist. Did you @ the user?")
		return
	}

	if len(notFound) != 0 || len(people) == 0 {
		s.ChannelMessageSend(m.ChannelID, "This user does not exist on this server. "+
			"Did you @ the user?")
		return
	}

	weeklyXP, err := AverageForTimeRange(TimeRange{MonthsAgoStart: 0, DaysAgoStart: 7, MonthsAgoEnd: 0, DaysAgoEnd: 0}, people[0])
	var weeklyStat string
	if err != nil {
		weeklyStat = "Not enough data"
	} else {
		weeklyStat = fmt.Sprintf("%d XP/day", int(weeklyXP))
	}

	prevWeekXP, err := AverageForTimeRange(TimeRange{MonthsAgoStart: 0, DaysAgoStart: 14, MonthsAgoEnd: 0, DaysAgoEnd: 7}, people[0])
	if err == nil && prevWeekXP > 0 {
		percentChange := float64(weeklyXP-prevWeekXP) / float64(prevWeekXP) * 100
		prefix := ""
		if percentChange > 0 {
			prefix = "+"
		} else if percentChange == 0 {
			prefix = "±"
		}
		weeklyStat = fmt.Sprintf("**%s\n%s%.2f%%** from previous week's daily XP",
			weeklyStat, prefix, percentChange)
	} else if err != nil {
		logger.Log.Error("Failed to construct prevWeekXP", err.Error())
	}

	monthlyXP, err := AverageForTimeRange(TimeRange{MonthsAgoStart: 1, DaysAgoStart: 0, MonthsAgoEnd: 0, DaysAgoEnd: 0}, people[0])
	var monthlyStat string
	if err != nil {
		monthlyStat = "Not enough data"
	} else {
		monthlyStat = fmt.Sprintf("%d XP/day", int(monthlyXP))
	}

	currentXP := people[0].History[len(people[0].History)-1].XP

	nextTier, err := NextXPTier(currentXP, chart.GlobalChartConfig.Milestones)
	var nextTierMsg string
	var nextTierDate string
	if err != nil {
		nextTierMsg = "You're already at the highest tier!"
		nextTierDate = "None"
	} else {
		d, err := ExpectedDate(currentXP, nextTier.XP, int(weeklyXP))
		nextTierMsg = fmt.Sprintf("%s: %d XP", nextTier.Name, nextTier.XP)
		if err != nil {
			nextTierDate = "We don't have enough data to determine this."
		} else {
			nextTierDate = d.Format("January 2, 2006")
		}
	}

	s.ChannelMessageSendEmbed(m.ChannelID,
		&discordgo.MessageEmbed{Author: &discordgo.MessageEmbedAuthor{},
			Color: 0x4444ff,
			Title: fmt.Sprintf("Stats for %s", names[people[0].UserID]),
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:   "Average XP last week",
					Value:  weeklyStat,
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Average XP last month",
					Value:  monthlyStat,
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Next XP tier",
					Value:  nextTierMsg,
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Estimated date to reach next tier",
					Value:  nextTierDate,
					Inline: true,
				},
			},
			Timestamp: time.Now().Format(time.RFC3339),
		},
	)
}

// ParseGraphArgs attempts to extract the graph arguments from a string.
// Returns only user-friendly errors.
func ParseGraphArgs(s string, senderID string) (Args, error) {
	// Ensure the correct number of spaces between args
	s = util.SpaceNormalizer(s)
	splitted := strings.Split(s, " ")

	// Allow for g!command as well.
	if splitted[0] != "g!" && len(splitted[0]) > 2 {
		splitted[0] = splitted[0][2:]
		splitted = append([]string{"g!"}, splitted...)
	}

	if len(splitted) > 10 {
		return Args{}, errors.New("you entered too many users or too long of a command. Please try again")
	}
	if len(splitted) <= 1 {
		// No args. Return the default.
		return Args{
				Usernames: nil,
				Sub:       false,
				Stats:     false,
				Top:       10,
				Days:      0,
				King:      false,
				Help:      false,
				Update:    false,
			},
			nil
	}
	args := Args{}

	// Only stuff after the g!
	splitted = splitted[1:]

	// Default check
	if len(splitted) == 0 {
		args.Top = 10
		return args, nil
	}

	// Big if-else time.
	// Structure of the parser is as follows:
	// 1. Check for the top-level command, whatever it may be.
	// 2. Enter a function that will either return the arguments to be passed or an error.
	// 3. Return that function *or* an error if necessary.

	// Declared here for readability.
	startingArg := splitted[0]
	caseInsensitive := true
	if util.StringChecker(startingArg, topSwitch, caseInsensitive) {
		// We have a top request.
		return parseTop(splitted)
	} else if util.StringChecker(startingArg, helpSwitch, caseInsensitive) {
		// Help request
		return Args{Help: true}, nil
	} else if util.StringChecker(startingArg, lastSwitch, caseInsensitive) {
		// We have a last request.
		return parseLast(splitted)
	} else if util.StringChecker(startingArg, usersSwitch, caseInsensitive) {
		// Users request.
		return parseUsers(splitted)
	} else if util.StringChecker(startingArg, subSwitch, caseInsensitive) {
		// Comparison request.
		return parseSub(splitted)
	} else if util.StringChecker(startingArg, statsSwitch, caseInsensitive) {
		// Stats request.
		return parseStats(splitted)
	} else if util.StringChecker(startingArg, manualUpdate, caseInsensitive) && senderID == croot {
		// Manual update
		return Args{Update: true}, nil
	}

	return Args{}, errors.New("your command wasn't recognized")
}

func parseTop(splitted []string) (Args, error) {
	// Determine if we have a number here
	if len(splitted) > 1 {
		topNum, err := strconv.Atoi(splitted[1])
		if err != nil && topNum > 0 {
			return Args{}, errors.New("please type a whole number after g! top")
		}
		return Args{Top: topNum}, nil
	}
	// Default
	return Args{Top: 10}, nil
}

func parseLast(splitted []string) (Args, error) {
	// Determine if we have a number here
	if len(splitted) > 1 {
		lastNum, err := strconv.Atoi(splitted[1])
		if err != nil && lastNum > 0 {
			return Args{}, errors.New("please type a whole number after g! last")
		}
		return Args{Days: lastNum, Top: 10}, nil
	}
	// Default
	return Args{Days: 365, Top: 10}, nil
}

func parseUsers(splitted []string) (Args, error) {
	if len(splitted) > 1 {
		args := Args{Usernames: make([]string, 0)}
		for _, user := range splitted[1:] {
			args.Usernames = append(args.Usernames, user)
		}
		args.Usernames = util.StripUsernames(args.Usernames)
		return args, nil
	}
	// Default error
	return Args{}, errors.New("You must pass at least one user to g! users")
}

func parseSub(splitted []string) (Args, error) {
	if len(splitted) > 1 && len(splitted) < 4 {
		args := Args{Usernames: make([]string, 0), Sub: true}
		for _, user := range splitted[1:] {
			args.Usernames = append(args.Usernames, user)
		}
		args.Usernames = util.StripUsernames(args.Usernames)
		return args, nil
	}
	// Default
	return Args{}, errors.New("You must pass two users to g! compare")
}

func parseStats(splitted []string) (Args, error) {
	if len(splitted) > 1 {
		return Args{Usernames: util.StripUsernames([]string{splitted[1]}), Stats: true}, nil
	} else if len(splitted) == 1 {
		return Args{Usernames: nil, Stats: true}, nil
	}
	// Default error
	return Args{}, errors.New("You called g! stats incorrectly")
}

func makeRankGraph(args *Args, s *discordgo.Session) (img []byte, path string, err error) {
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

	src, _, err := RankLineChart(&c, args, s)
	if err != nil {
		logger.Log.Error("Failed to construct chart:", err.Error())
		return nil, "", err
	}

	img, path = src.Make()

	return img, path, nil
}

func makeUserComparisonGraph(args *Args, s *discordgo.Session) (img []byte, path string, err error) {
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

	src, _, err := UserLineChart(&c, args, s)
	if err != nil {
		logger.Log.Error("Failed to construct chart:", err.Error())
		return nil, "", err
	}

	img, path = src.Make()

	return img, path, nil
}

func makeUserSubtractionGraph(args *Args, s *discordgo.Session) (img []byte, path string, err error) {
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

	src, _, err := SubLineChart(&c, args, s)
	// No milestones on this graph.
	src.ShowMilestones = false
	if err != nil {
		logger.Log.Error("Failed to construct chart:", err.Error())
		return nil, "", err
	}

	img, path = src.Make()

	return img, path, nil
}
