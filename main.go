package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EntryType string

const (
	WORKED EntryType = "WORKED"
	LEAVE  EntryType = "LEAVE"
)

type Employee struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

type Entry struct {
	Id         *string   `json:"id"`
	EmployeeId string    `json:"employeeId" binding:"required"`
	Date       string    `json:"date" binding:"required"`
	Type       EntryType `json:"type" binding:"required"`
}

type Stats struct {
	Worked  int `json:"worked"`
	Leave   int `json:"leave"`
	Balance int `json:"balance"`
}

func auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		auth := ctx.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		fmt.Println(tokenStr)

		//To be filled after parsing jwt
		claims := jwt.MapClaims{}
		fmt.Println("The JWT secret is: ", os.Getenv("JWT_SECRET"))
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		ctx.Set("user", claims)

		ctx.Next() // Pass to the next handler
	}
}

type EntryWithStats struct {
	ID         string    `json:"id"`
	EmployeeId string    `json:"employeeId"`
	Type       string    `json:"type"`
	Date       time.Time `json:"date"`
	Worked     int       `json:"worked"`
	Leave      int       `json:"leave"`
	Balance    int       `json:"balance"`
}

func getEmployees(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pool.Query(c.Request.Context(), `SELECT name, id FROM employees`)
		if err != nil {
			log.Printf("query error: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "DB query failed"})
			return
		}
		defer rows.Close()

		emps, err := pgx.CollectRows(rows, pgx.RowToStructByName[Employee])
		if err != nil {
			log.Printf("scan error: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "DB scan failed"})
			return
		}

		c.JSON(http.StatusOK, emps)

	}
}
func createEntry(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		sql, err := os.ReadFile("./sql/create_entry.sql")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to read SQL"})
			return
		}
		var entry Entry
		if err := c.ShouldBindJSON(&entry); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		id := uuid.NewString()

		row := pool.QueryRow(c.Request.Context(), string(sql),
			entry.EmployeeId, id, entry.Type, entry.Date)

		var ews EntryWithStats
		if err := row.Scan(
			&ews.ID, &ews.EmployeeId, &ews.Type, &ews.Date,
			&ews.Worked, &ews.Leave, &ews.Balance,
		); err != nil {
			if err == pgx.ErrNoRows {
				c.JSON(400, gin.H{"error": "insufficient balance"})
				return
			}
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(201, ews)
	}
}

func main() {
	// if err := godotenv.Load(); err != nil {
	// 	log.Fatal("Error loading .env file")

	// }

	ctx, stopCtx := context.WithTimeout(context.Background(), 10*time.Second)
	defer stopCtx()
	pool := InitPool(ctx, os.Getenv("DATABASE_URL"))

	router := gin.Default()
	router.GET("/ping", func(ctx *gin.Context) {
		fmt.Println(ctx)
		ctx.JSON(200, gin.H{
			"message": "pong",
		})
	})

	{
		api := router.Group("/api")
		api.GET("/employees", getEmployees(pool))
		api.POST("/entries", createEntry(pool))

	}
	router.Run(":8080")
}
