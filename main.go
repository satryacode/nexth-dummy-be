// main.go
package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/satryacode/nexth-dummy-be/config"
	"github.com/satryacode/nexth-dummy-be/db"
	"github.com/satryacode/nexth-dummy-be/handlers"
	"github.com/satryacode/nexth-dummy-be/middleware"
)

func main() {
	cfg := config.Load()

	logger, err := middleware.NewLogger()
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer logger.Sync()

	pool, err := db.NewPool(cfg)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer pool.Close()

	r := gin.New()

	// CORS wide-open (intentional vuln)
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "*")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.Use(middleware.RequestLogger(logger))

	r.POST("/register", handlers.Register(pool))
	r.POST("/login", handlers.Login(pool, cfg.JWTSecret))
	r.GET("/home", handlers.Home(pool, cfg.JWTSecret))

	log.Printf("starting dummy-be on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
