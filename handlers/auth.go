package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/satryacode/nexth-dummy-be/models"
)

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Register(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var userID int
		// No input sanitization, password stored plaintext (intentional vulns)
		err := pool.QueryRow(context.Background(),
			"INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id",
			req.Username, req.Email, req.Password,
		).Scan(&userID)

		if err != nil {
			// Expose raw DB error (intentional vuln)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "user registered", "user_id": userID})
	}
}

func Login(pool *pgxpool.Pool, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// SQL Injection vulnerability: string interpolation instead of parameterized query
		query := fmt.Sprintf(
			"SELECT id, username, email, password, created_at, blocked FROM users WHERE username='%s' AND password='%s'",
			req.Username, req.Password,
		)

		var user models.User
		err := pool.QueryRow(context.Background(), query).Scan(
			&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.Blocked,
		)
		if err != nil {
			// Hardcoded default admin fallback (intentional vuln)
			if req.Username == "admin" && req.Password == "admin123" {
				user = models.User{
					ID:        1,
					Username:  "admin",
					Email:     "admin@dummy.com",
					Password:  "admin123",
					CreatedAt: time.Now(),
				}
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
				return
			}
		}

		if user.Blocked == 1 {
			c.JSON(http.StatusForbidden, gin.H{"error": "account blocked"})
			return
		}

		// JWT signed with weak hardcoded secret
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":  user.ID,
			"username": user.Username,
			"exp":      time.Now().Add(24 * time.Hour).Unix(),
		})
		tokenStr, _ := token.SignedString([]byte(jwtSecret))

		// Returns full user object including plaintext password (intentional vuln)
		c.JSON(http.StatusOK, gin.H{
			"token": tokenStr,
			"user":  user,
		})
	}
}
