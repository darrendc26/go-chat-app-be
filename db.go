package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

var db *pgx.Conn
var redisClient *redis.Client

func connectDb() {
	connStr := //TODO: Add your postgreSQL Connection string here.
	var err error
	db, err = pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to PostgreSQL")
}

func connectRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	fmt.Println("Connected to Redis")
}

func InitDB() {
	connectDb()
	connectRedis()
}
