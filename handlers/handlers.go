package handlers

import (
	"github.com/chorin1/scoreboard-server/db"
	"github.com/gofiber/fiber/v2"
	"log"
	"strconv"
)

func NewScoreHandler(database db.Database) fiber.Handler {
	return func(c *fiber.Ctx) error {
		u := new(db.User)
		if err := c.BodyParser(u); err != nil {
			log.Println(err)
			return fiber.ErrBadRequest
		}
		// TODO: check username length
		// TODO: check score is above a certain number
		// TODO: check that we're not overriting its
		rank, err := database.SaveUser(c.Context(), u)
		if err != nil {
			log.Println(err)
			return fiber.ErrInternalServerError
		}

		return c.SendString(strconv.FormatInt(rank, 10))
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
