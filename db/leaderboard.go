package db

import (
	"context"
	"log"
	"time"
)

const (
	leaderboardKey           = "leaderboard"
	maxEntries               = 10000
	cleanBottomRanksInterval = 1 * time.Minute
)

type Leaderboard struct {
	Count int `json:"count"`
	Users []*User
}

func (db *Database) GetTop10(ctx context.Context) (*Leaderboard, error) {
	records, err := db.Client.ZRevRangeWithScores(ctx, leaderboardKey, 0, 9).Result()
	if err != nil {
		return nil, err
	}
	if records == nil {
		return nil, ErrNil
	}
	users := make([]*User, len(records))
	for i, record := range records {
		users[i] = &User{
			Name:  Member(record.Member.(string)).ExtractName(),
			Score: uint64(record.Score),
			Rank:  int64(i) + 1,
		}
	}
	leaderboard := &Leaderboard{
		Count: len(records),
		Users: users,
	}
	return leaderboard, nil
}

func (db *Database) removeBottomRanks() {
	for range time.NewTicker(cleanBottomRanksInterval).C {
		card, err := db.Client.ZCard(context.Background(), leaderboardKey).Result()
		if err != nil {
			log.Printf("error cleaning records: %v\n", err)
			return
		}
		if card > maxEntries {
			// removes the bottom entries
			removed, err := db.Client.
				ZRemRangeByRank(context.Background(), leaderboardKey, 0, card-maxEntries-1).
				Result()
			if err != nil {
				log.Printf("error cleaning records: %v\n", err)
				return
			}
			log.Printf("cleaned %d entries\n", removed)
		}
	}
}

func (db *Database) DeleteAllUsers() error {
	_, err := db.Client.Del(context.Background(), leaderboardKey).Result()
	if err != nil {
		return err
	}
	log.Printf("deleted all entries")
	return nil
}
