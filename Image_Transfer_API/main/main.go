package main

import (
	"Image-Transfer-API/pkg/routes"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error while loading environment variables.")
		os.Exit(1)
	}

	port := os.Getenv("SERVER_PORT")

	if port == "" {
		port = "8000"
	}

	router := gin.New()

	routes.InitiateRoutes(router)
	routes.UploadRoutes(router)
	routes.TerminationRoutes(router)

	router.Run(":", port)
}
