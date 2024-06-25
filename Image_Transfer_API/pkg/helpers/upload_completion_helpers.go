package helpers

import (
	"Image-Transfer-API/pkg/models"
	redis_database "Image-Transfer-API/pkg/redis"
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

func CheckReceivedStatus(c *gin.Context, chunkNumber int, session models.Session) bool {
	// Get the redis client.
	redisClient := redis_database.GetRedisClient()

	// Run a loop and whether all the chunks recieved status is true or not.
	for i := 1; i <= chunkNumber; i++ {
		// Get the corresponding session and chunk value from the redis.
		val, err := redisClient.ChunksClient.Get(*redisClient.Ctx, fmt.Sprintf("%s__%d", session.SessionID, i)).Bytes()
		if c.Err() != nil {
			if c.Err() == context.DeadlineExceeded {
				c.JSON(http.StatusRequestTimeout, gin.H{"error": "Error: Request Timeout"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Internal Server error."})
			}

			return false
		}

		// If it does not exist, reply back with bad request, chunk size wrong.
		if err == redis.Nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error: All chunks not received."})
			return false
		}
		if c.Err() != nil {
			if c.Err() == context.DeadlineExceeded {
				c.JSON(http.StatusRequestTimeout, gin.H{"error": "Error: Request Timeout"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Internal Server error."})
			}

			return false
		}

		// If it is false, reply back with upload chunk failed, with chunk id number.
		var chunkInfo models.ChunkInfo
		dec := gob.NewDecoder(bytes.NewReader(val))
		err = dec.Decode(&chunkInfo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Failed to decode chunkinfo.",
				"chunk number": chunkNumber})

			return false
		}
		if c.Err() != nil {
			if c.Err() == context.DeadlineExceeded {
				c.JSON(http.StatusRequestTimeout, gin.H{"error": "Error: Request Timeout"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Internal Server error."})
			}

			return false
		}

	}

	// Check for same session id and chunkNumber + 1 if it exists in the redis or not. If not then reply back with upload complete message.
	_, err := redisClient.ChunksClient.Get(*redisClient.Ctx, fmt.Sprintf("%s__%d", session.SessionID, chunkNumber+1)).Bytes()
	if c.Err() != nil {
		if c.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusRequestTimeout, gin.H{"error": "Error: Request Timeout"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error: Internal Server error."})
		}

		return false
	}

	if err == redis.Nil {
		c.JSON(http.StatusOK, gin.H{"message": "Message: All chunks are received and uploaded successfully."})
		return true
	}

	// If yes, reply back with bad request, chunk size wrong
	c.JSON(http.StatusBadRequest, gin.H{"error": "Error: Total number of chunks mismatch."})
	return true
}
