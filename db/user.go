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
	Name     string `json:"name" binding:"required"`
	DeviceID string `json:"uuid" binding:"required"`
	Score    uint64 `json:"score" binding:"required"`
	Rank     int64  `json:"rank"`
}

func GenerateMember(user *User) Member {
	var sb strings.Builder
	sb.Grow(32)
	sb.WriteString(user.DeviceID)
	sb.WriteString(user.Name)
	return Member(sb.String())
}

func (member Member) ExtractName() string {
	return string(member[32:])
}

func (db *Database) SaveUser(ctx context.Context, user *User) error {
	member := string(GenerateMember(user))
	record := &redis.Z{
		Score:  float64(user.Score),
		Member: member,
	}
	pipe := db.Client.TxPipeline()
	// TODO: only update if greater
	pipe.ZAdd(ctx, leaderboardKey, record)
	rank := pipe.ZRevRank(ctx, leaderboardKey, member)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	user.Rank = rank.Val() + 1
	return nil
}

//func (db *Database) GetUser(ctx context.Context, username string) (*User, error) {
//	pipe := db.Client.TxPipeline()
//	score := pipe.ZScore(ctx, leaderboardKey, username)
//	rank := pipe.ZRevRank(ctx, leaderboardKey, username)
//	_, err := pipe.Exec(ctx)
//	if err != nil {
//		return nil, err
//	}
//	if score == nil {
//		return nil, ErrNil
//	}
//	return &User{
//		Name:  username,
//		Score: uint64(score.Val()),
//		Rank:  rank.Val() + 1,
//	}, nil
//}
