package bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/snyderks/xp-bot/primitives"
)

// Ranking is a single rank for a user at an (unspecified) point in time.
type Ranking struct {
	Rank   int    `json:"rank"`
	Score  int    `json:"score"`
	UserID string `json:"user_id"`
}

// Rankings is a wrapper object for all rankings, with the
// guild (server) ID listed for tracking.
type Rankings struct {
	GuildID string    `json:"guild_id"`
	Ranks   []Ranking `json:"rankings"`
}

var rootURI = "https://api.tatsu.gg/v1/"

// APIKey is the private API key used to authenticate to the Tatsu API.
var APIKey string

func init() {
	APIKey = os.Getenv("TATSU_API_KEY")
}

func readKey(path string) (string, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

func addAuth(r *http.Request, key string) {
	r.Header.Set("Authorization", key)
}

// RankingsToPeople converts a Rankings object to a map of people for later DB storage.
func RankingsToPeople(ranks Rankings) (map[string]primitives.Person, error) {
	if len(ranks.Ranks) < 1 {
		return nil, errors.New("No ranks available")
	}

	ret := make(map[string]primitives.Person)

	for _, r := range ranks.Ranks {
		ret[r.UserID] = primitives.Person{r.UserID, r.Score, r.Rank}
	}

	return ret, nil
}

// GetRankings retrieves the current XP rankings for a specified guild (server).
func GetRankings(guildID string) (Rankings, error) {
	URI := fmt.Sprintf("%s/guilds/%s/rankings/all", rootURI, guildID)

	req, err := http.NewRequest(http.MethodGet, URI, nil)
	if err != nil {
		return Rankings{}, fmt.Errorf("GetRankings(): failed to build rankings request. Error: %s", err.Error())
	}
	addAuth(req, APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Rankings{}, fmt.Errorf("GetRankings(): failed to retrieve rankings. Error: %s", err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Rankings{}, fmt.Errorf("GetRankings(): failed to retrieve rankings with bad status: %d", resp.StatusCode)
	}

	ranks := Rankings{}

	json.NewDecoder(resp.Body).Decode(&ranks)

	return ranks, nil
}
