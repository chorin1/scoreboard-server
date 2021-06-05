package main

import (
	"github.com/chorin1/scoreboard-server/db"
	"github.com/gofiber/fiber/v2"
	"log"
	"os"
)

var (
	dbAddr = os.Getenv("REDIS_URL")
	auth   = map[string]string{"raidar": "gaMe"}
	port   = os.Getenv("HOST_PORT")
)

func main() {
	database, err := db.NewDatabase(dbAddr)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	app := fiber.New()

	// TODO: uncomment later for basic auth
	// app.Use(basicauth.New(basicauth.Config{Users: auth}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})
	app.Post("/newScore", func(c *fiber.Ctx) error {
		u := new(db.User)
		if err := c.BodyParser(u); err != nil {
			log.Println(err)
			return fiber.ErrBadRequest
		}
		// TODO: check username length
		// TODO: check score is above a certain number
		err := database.SaveUser(u)
		if err != nil {
			log.Println(err)
			return fiber.ErrInternalServerError
		}
		return c.SendString("ok!")
	})
	app.Get("/getScores", func(c *fiber.Ctx) error {
		leaderboard, err := database.GetLeaderboard()
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
	})

	log.Fatal(app.Listen(":" + port))
}

//
//func initRouter(database *db.Database) *gin.Engine {
//	r := gin.Default()
//	r.GET("/points/:username", func (c *gin.Context) {
//		username := c.Param("username")
//		user, err := database.GetUser(username)
//		if err != nil {
//			if err == db.ErrNil {
//				c.JSON(http.StatusNotFound, gin.H{"error": "No record found for " + username})
//				return
//			}
//			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//			return
//		}
//		c.JSON(http.StatusOK, gin.H{"user": user})
//	})
//
//	r.POST("/points", func (c *gin.Context) {
//		var userJson db.User
//		if err := c.ShouldBindJSON(&userJson); err != nil {
//			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//			return
//		}
//		err := database.SaveUser(&userJson)
//		if err != nil {
//			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//			return
//		}
//		c.JSON(http.StatusOK, gin.H{"user": userJson})
//	})
//
//	r.GET("/leaderboard", func(c *gin.Context) {
//		leaderboard, err := database.GetLeaderboard()
//		if err != nil {
//			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//			return
//		}
//		c.JSON(http.StatusOK, gin.H{"leaderboard": leaderboard})
//	})
//
//	return r
//}
