package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/chorin1/scoreboard-server/db"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

var (
	validUser = db.User{
		Name:     "user",
		DeviceID: uuid.NewString(),
		Score:    4000,
	}
	member              = validUser.DeviceID + validUser.Name
	memberRank          = int64(42)
	connectionFailedErr = errors.New("connection failed")
)

func TestUserValidation(t *testing.T) {
	tests := []struct {
		name     string
		user     *db.User
		expected bool
	}{
		{name: "valid user", user: &validUser, expected: true},
		{name: "not a uuid", user: &db.User{Name: "user", DeviceID: "a-b-c-d", Score: 4000}, expected: false},
		{name: "empty name", user: &db.User{Name: "", DeviceID: uuid.NewString(), Score: 4000}, expected: false},
		{name: "too short name", user: &db.User{Name: "ba", DeviceID: uuid.NewString(), Score: 4000}, expected: false},
		{name: "too long name", user: &db.User{Name: "papopepopalala", DeviceID: uuid.NewString(), Score: 4000}, expected: false},
		{name: "too low score", user: &db.User{Name: "user", DeviceID: uuid.NewString(), Score: 1000}, expected: false},
		{name: "impossible score", user: &db.User{Name: "user", DeviceID: uuid.NewString(), Score: 1_000_000_000}, expected: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := validateUser(tt.user); result != tt.expected {
				t.Errorf("validateUser() = %v, should be %v", result, tt.expected)
			}
		})
	}
}

func newUserRequest(t *testing.T, user db.User) *http.Request {
	b, err := json.Marshal(user)
	if err != nil {
		t.Error(err)
	}
	req := httptest.NewRequest("POST", "/", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestNewHighScoreHandler(t *testing.T) {
	tests := []struct {
		name                 string
		user                 db.User
		expectationsInitFunc func(redismock.ClientMock) // mocks database expectations and returned data
		expectedStatusCode   int
		expectedRank         int64
	}{
		{
			name: "new high score, returns the rank",
			user: validUser,
			expectationsInitFunc: func(mock redismock.ClientMock) {
				mock.ExpectZScore(db.LeaderboardKey, member).RedisNil()
				mock.ExpectTxPipeline()
				mock.ExpectZAdd(db.LeaderboardKey, &redis.Z{Score: float64(validUser.Score), Member: member}).SetVal(1)
				mock.ExpectZRevRank(db.LeaderboardKey, member).SetVal(memberRank - 1)
				mock.ExpectTxPipelineExec()
			},
			expectedStatusCode: http.StatusOK,
			expectedRank:       memberRank,
		},
		{
			name: "a higher score for this user exist, returns -1",
			user: validUser,
			expectationsInitFunc: func(mock redismock.ClientMock) {
				mock.ExpectZScore(db.LeaderboardKey, member).SetVal(float64(validUser.Score + 1))
			},
			expectedStatusCode: http.StatusOK,
			expectedRank:       db.ScoreNotUpdated,
		},
		{
			name:                 "invalid user request",
			user:                 db.User{},
			expectationsInitFunc: func(mock redismock.ClientMock) {},
			expectedStatusCode:   http.StatusBadRequest,
		},
		{
			name: "database is down",
			user: validUser,
			expectationsInitFunc: func(mock redismock.ClientMock) {
				mock.ExpectZScore(db.LeaderboardKey, member).SetErr(connectionFailedErr)
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database, mock := redismock.NewClientMock()
			handler := NewScoreHandler(db.Database{Client: database})
			app := fiber.New()
			app.Post("/", handler)

			tt.expectationsInitFunc(mock)

			req := newUserRequest(t, tt.user)
			resp, err := app.Test(req)

			if resp.StatusCode != tt.expectedStatusCode {
				t.Errorf("Expcted status code %v but got %v", tt.expectedStatusCode, resp.StatusCode)
			}

			if resp.StatusCode == http.StatusOK {
				var returnedUser db.User
				err = json.NewDecoder(resp.Body).Decode(&returnedUser)
				if err != nil {
					t.Error(err)
				}
				if tt.expectedRank != returnedUser.Rank {
					t.Errorf("Expcted rank %v but got %v", tt.expectedRank, returnedUser.Rank)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestGetScoresHandler(t *testing.T) {
	database, mock := redismock.NewClientMock()
	handler := GetScoresHandler(db.Database{Client: database})
	app := fiber.New()
	app.Get("/", handler)

	var scoreRecords []redis.Z
	for i := 0; i < 10; i++ {
		scoreRecords = append(scoreRecords, redis.Z{Score: float64(5000 + i), Member: member + strconv.Itoa(i)})
	}

	mock.ExpectZRevRangeWithScores(db.LeaderboardKey, 0, 9).SetVal(scoreRecords)

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fail()
	}

	var leaderboard db.Leaderboard
	err = json.NewDecoder(resp.Body).Decode(&leaderboard)
	if err != nil {
		t.Error(err)
	}

	for i, leader := range leaderboard.Users {
		expected := db.User{
			Name:  validUser.Name + strconv.Itoa(i),
			Score: uint64(scoreRecords[i].Score),
			Rank:  int64(i + 1),
		}
		if expected != *leader {
			t.Errorf("expected user in leaderboard to be %v, got %v instead", expected, *leader)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestGetScoresHandlerDBError(t *testing.T) {
	database, mock := redismock.NewClientMock()
	handler := GetScoresHandler(db.Database{Client: database})
	app := fiber.New()
	app.Get("/", handler)

	mock.ExpectZRevRangeWithScores(db.LeaderboardKey, 0, 9).SetErr(connectionFailedErr)

	req := httptest.NewRequest("GET", "/", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fail()
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
