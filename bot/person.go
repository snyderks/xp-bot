package bot

import (
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/snyderks/xp-bot/db"
	"github.com/snyderks/xp-bot/primitives"
)

// GetNicknames retrives nicknames for a list of user IDs.
func GetNicknames(userIDs []string, s *discordgo.Session) (map[string]string, error) {
	d, err := db.CreateDB(DBURI)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize DB: %s", err.Error())
	}

	now := time.Now()

	results := d.GetNicknames(userIDs)

	toUpdate := make([]primitives.Nickname, 0)

	// Get live results if needed
	for id, r := range results {
		// Only update every 12 hours at best
		if r.UserID == "" || now.Sub(r.LastUpdated) > time.Hour*12 {
			// No result in cache
			u, err := s.User(id)
			if err != nil {
				// User doesn't exist
				return nil, errors.New("User does not exist at all")
			}
			n := primitives.Nickname{
				UserID:      id,
				Nickname:    u.Username,
				LastUpdated: now}
			results[id] = n
			toUpdate = append(toUpdate, n)
		}
	}
	// Want to return results ASAP, so do this in another thread.
	go d.SetNicknames(toUpdate)

	fmt.Println(results)
	// Extract nicknames from the results.
	ret := make(map[string]string, 0)
	for id, r := range results {
		ret[id] = r.Nickname
	}
	fmt.Println(ret)
	return ret, nil
}
