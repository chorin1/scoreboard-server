package db

var leaderboardKey = "leaderboard"

type Leaderboard struct {
	Count int `json:"count"`
	Users []*User
}

func (db *Database) GetLeaderboard() (*Leaderboard, error) {
	scores := db.Client.ZRevRangeWithScores(Ctx, leaderboardKey, 0, -1)
	if scores == nil {
		return nil, ErrNil
	}
	count := len(scores.Val())
	users := make([]*User, count)
	for idx, member := range scores.Val() {
		users[idx] = &User{
			Username: member.Member.(string),
			Score:    uint64(int(member.Score)),
			Rank: uint64(idx),
		}
	}
	leaderboard := &Leaderboard{
		Count: count,
		Users: users,
	}
	return leaderboard, nil
}
