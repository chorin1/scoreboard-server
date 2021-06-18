package main

import (
	"github.com/chorin1/scoreboard-server/db"
	"github.com/chorin1/scoreboard-server/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"log"
	"os"
	"time"
)

var (
	dbAddr = os.Getenv("REDIS_URL")
	auth   = map[string]string{os.Getenv("HTTP_USER"): os.Getenv("HTTP_PASS")}
	port   = os.Getenv("HOST_PORT")
)

func main() {
	database, err := db.NewDatabase(dbAddr)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	app := fiber.New()
	app.Use(limiter.New(limiter.Config{
		Max:        5,
		Expiration: 10 * time.Second,
	}))
	app.Use(logger.New(logger.Config{
		TimeFormat: "2006-01-02T15:04:05",
		TimeZone:   "UTC",
	}))
	app.Use(recover.New())

	app.Use(basicauth.New(basicauth.Config{Users: auth}))

	app.Post("/newScore", handlers.NewScoreHandler(*database))
	app.Get("/getScores", handlers.GetScoresHandler(*database)) // can be cached later
	app.Delete("/deleteAllScores", handlers.DeleteAllHandler(*database))

	app.Use(func(c *fiber.Ctx) error { return fiber.ErrNotFound })

	log.Fatal(app.Listen(":" + port))
}
