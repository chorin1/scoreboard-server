package db

import (
	"context"
	"github.com/go-redis/redis/v8"
	"strings"
)

// Member is a combination of the picked name by the user and the device id
// uuid + name
type Member string

type User struct {
	Name     string `json:"name,omitempty"`
	DeviceID string `json:"uuid,omitempty"`
	Score    uint64 `json:"score,omitempty"`
	Rank     int64  `json:"rank,omitempty"`
}

func GenerateMember(user *User) Member {
	var sb strings.Builder
	sb.Grow(50)
	sb.WriteString(user.DeviceID)
	sb.WriteString(user.Name)
	return Member(sb.String())
}

func (member Member) ExtractName() string {
	return string(member[36:])
}

func (db *Database) SaveUser(ctx context.Context, user *User) error {
	member := string(GenerateMember(user))

	// if the member has an existing higher score -> abort
	// TODO: remove this check once ZADD with GT is added to go-redis
	existingScore, err := db.Client.ZScore(ctx, leaderboardKey, member).Result()
	if err != nil && err != redis.Nil {
		return err
	}
	if uint64(existingScore) >= user.Score {
		*user = User{Rank: -1} // convention that score wasn't updated
		return nil
	}

	record := &redis.Z{
		Score:  float64(user.Score),
		Member: member,
	}
	pipe := db.Client.TxPipeline()
	pipe.ZAdd(ctx, leaderboardKey, record)
	rank := pipe.ZRevRank(ctx, leaderboardKey, member)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}
	// saves space on returned message
	*user = User{Rank: rank.Val() + 1}
	return nil
}
