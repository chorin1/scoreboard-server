package db

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type User struct {
	Username string `json:"username" binding:"required"`
	Score    uint64 `json:"score" binding:"required"`
	Rank     int64  `json:"rank"`
}

func (db *Database) SaveUser(ctx context.Context, user *User) (int64, error) {
	member := &redis.Z{
		Score:  float64(user.Score),
		Member: user.Username,
	}
	pipe := db.Client.TxPipeline()
	pipe.ZAdd(ctx, leaderboardKey, member)
	rank := pipe.ZRevRank(ctx, leaderboardKey, user.Username)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return rank.Val() + 1, nil
}

func (db *Database) GetUser(ctx context.Context, username string) (*User, error) {
	pipe := db.Client.TxPipeline()
	score := pipe.ZScore(ctx, leaderboardKey, username)
	rank := pipe.ZRevRank(ctx, leaderboardKey, username)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}
	if score == nil {
		return nil, ErrNil
	}
	return &User{
		Username: username,
		Score:    uint64(score.Val()),
		Rank:     rank.Val() + 1,
	}, nil
}
