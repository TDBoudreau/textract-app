package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/textract"
	_ "github.com/joho/godotenv/autoload"

	"textract-app/internal/database"
)

type Server struct {
	port int

	db     database.Service
	client *textract.Client
}

func NewServer() (*http.Server, error) {
	client, err := setupTextractClient()
	if err != nil {
		return nil, err
	}

	port, _ := strconv.Atoi(os.Getenv("PORT"))
	NewServer := &Server{
		port: port,

		db:     database.New(),
		client: client,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server, nil
}

func setupTextractClient() (*textract.Client, error) {
	ctx := context.Background()

	// Manually set your AWS credentials
	accessKey := os.Getenv("AWS_ACCESS_KEY")
	secretKey := os.Getenv("AWS_SECRET_KEY")
	sessionToken := "" // If using temporary credentials, set the session token here

	// Load the AWS configuration with your region and credentials.
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-west-2"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, sessionToken)),
	)
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		return nil, err
	}

	// Create a Textract client.
	client := textract.NewFromConfig(cfg)
	return client, nil
}
