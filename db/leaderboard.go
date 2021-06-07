package db

import "context"

var leaderboardKey = "leaderboard"

type Leaderboard struct {
	Count int `json:"count"`
	Users []*User
}

func (db *Database) GetTop10(ctx context.Context) (*Leaderboard, error) {
	scores := db.Client.ZRevRangeWithScores(ctx, leaderboardKey, 0, 9)
	if scores == nil {
		return nil, ErrNil
	}
	count := len(scores.Val())
	users := make([]*User, count)
	for idx, member := range scores.Val() {
		users[idx] = &User{
			Username: member.Member.(string),
			Score:    uint64(member.Score),
			Rank:     int64(idx) + 1,
		}
	}
	leaderboard := &Leaderboard{
		Count: count,
		Users: users,
	}
	return leaderboard, nil
}
