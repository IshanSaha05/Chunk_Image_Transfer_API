package redis_database

import (
	"context"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

// Declaring structure for redis client.
type RedisClient struct {
	SessionClient *redis.Client
	ChunksClient  *redis.Client
	Ctx           *context.Context
	Once          sync.Once
}

// Declaring variable for redis client using the structure above.
var redisClient RedisClient

// Creating a init function to load all the variables in the beginning and connect with the redis database.
func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error while loading environment variables.")
		os.Exit(1)
	}

	address := os.Getenv("REDIS_ADDRESS")
	password := os.Getenv("REDIS_PASSWORD")
	sessionDatabase, err := strconv.Atoi(os.Getenv("REDIS_SESSION_DB"))
	if err != nil {
		log.Fatalf("Error: Problem while converting redis db value from string to integer.")
		os.Exit(1)
	}
	chunksDatabase, err := strconv.Atoi(os.Getenv("REDIS_CHUNKS_DB"))
	if err != nil {
		log.Fatalf("Error: Problem while converting redis db value from string to integer.")
		os.Exit(1)
	}

	// Once Do is used to make sure only one time this code within is executed in case of multi-threaded excution.
	redisClient.Once.Do(func() {
		redisClient.SessionClient = redis.NewClient(&redis.Options{
			Addr:     address,
			Password: password,
			DB:       sessionDatabase,
		})

		redisClient.ChunksClient = redis.NewClient(&redis.Options{
			Addr:     address,
			Password: password,
			DB:       chunksDatabase,
		})

		_, err = redisClient.SessionClient.Ping(*redisClient.Ctx).Result()
		if err != nil {
			log.Fatalf("Error: Could not connect to redis client")
			os.Exit(1)
		}
		log.Println("Message: Successfully connected to redis client")

		_, err = redisClient.ChunksClient.Ping(*redisClient.Ctx).Result()
		if err != nil {
			log.Fatalf("Error: Could not connect to redis client")
			os.Exit(1)
		}
	})
}

// Creating function to retun the redis client.
func GetRedisClient() *RedisClient {
	return &redisClient
}
