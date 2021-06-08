package handlers

import (
	"github.com/chorin1/scoreboard-server/db"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"log"
)

const (
	maxScore      = 999_999_999
	minScore      = 3000
	maxNameLength = 13
)

func validateUser(user *db.User) bool {
	if len(user.Name) > maxNameLength {
		return false
	}
	if user.Score > maxScore || user.Score < minScore {
		return false
	}
	_, err := uuid.Parse(user.DeviceID)
	if err != nil {
		return false
	}

	return true
}

func NewScoreHandler(database db.Database) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := new(db.User)
		if err := c.BodyParser(user); err != nil {
			log.Println(err)
			return fiber.ErrBadRequest
		}
		if !validateUser(user) {
			return fiber.NewError(fiber.StatusBadRequest, "user is invalid")
		}

		// TODO: what to do when user already exist? allow overwrite?
		err := database.SaveUser(c.Context(), user)
		if err != nil {
			log.Println(err)
			return fiber.ErrInternalServerError
		}

		// enriched with rank
		return c.JSON(user)
	}
}

func GetScoresHandler(database db.Database) fiber.Handler {
	return func(c *fiber.Ctx) error {
		leaderboard, err := database.GetTop10(c.Context())
		if err != nil {
			log.Println(err)
			return fiber.ErrInternalServerError
		}
		err = c.JSON(leaderboard)
		if err != nil {
			log.Println(err)
			return fiber.ErrInternalServerError
		}
		return nil
	}
}

func DeleteAllHandler(database db.Database) fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := database.DeleteAllUsers()
		if err != nil {
			return err
		}
		return nil
	}
}
