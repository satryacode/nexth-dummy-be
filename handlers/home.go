package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/satryacode/nexth-dummy-be/models"
)

func Home(pool *pgxpool.Pool, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// Broken auth: token parsed but signature NOT verified (intentional vuln)
		if authHeader != "" {
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			parser := jwt.NewParser()
			token, _, _ := parser.ParseUnverified(tokenStr, jwt.MapClaims{})

			if token != nil {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					username, _ := claims["username"].(string)
					if username != "" && pool != nil {
						var user models.User
						err := pool.QueryRow(context.Background(),
							"SELECT id, username, email, created_at FROM users WHERE username=$1",
							username,
						).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
						if err == nil {
							c.JSON(http.StatusOK, gin.H{
								"message": "welcome",
								"user": gin.H{
									"id":         user.ID,
									"username":   user.Username,
									"email":      user.Email,
									"created_at": user.CreatedAt,
								},
							})
							return
						}
					}
				}
			}
		}

		// No valid token or no DB — still return 200 (intentional broken auth)
		c.JSON(http.StatusOK, gin.H{
			"message": "welcome",
			"user":    nil,
		})
	}
}
