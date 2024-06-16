package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
	"kamchatka-satellite/chatgpt"
	"kamchatka-satellite/qgis"
	"log"
	"os"
	"time"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Get database credentials from environment variables
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbIP := os.Getenv("DB_IP")
	dbName := os.Getenv("DB_NAME")
	aiKey := os.Getenv("AI_KEY")
	qgisToken := os.Getenv("QGIS_TOKEN")

	// Connect to PostgreSQL database
	conn, err := pgx.Connect(context.Background(), fmt.Sprintf("postgresql://%s:%s@%s/%s", dbUser, dbPass, dbIP, dbName))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())

	// Ping the database to ensure the connection is established
	err = conn.Ping(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to the database!")

	// Create a new instance of the server, passing the database connection
	serverClient := NewServer(conn)
	qgisClint := qgis.NewQgis(qgisToken)
	chatgptClient := chatgpt.NewChatGpt(aiKey)

	dataChannel := make(chan []*Coordinates)

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				data, _ := serverClient.GetRouteCoordinates(context.Background())
				dataChannel <- data
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				select {
				case data := <-dataChannel:
					for id := range data {
						var x1 float32
						var x2 float32
						var y1 float32
						var y2 float32
						if data[id+1] != nil {
							x1 = data[id].Dot[0]
							y1 = data[id].Dot[1]
							x2 = data[id+1].Dot[0]
							y2 = data[id+1].Dot[1]
						} else {
							x1 = data[id].Dot[0]
							y1 = data[id].Dot[1]
							x2 = data[id+1].Dot[0]
							y2 = data[id+1].Dot[1]
						}
						image, _ := qgisClint.GetSatelliteData(x1, y1, x2, y2)
						hasAny, trouble, err := chatgptClient.Prompt(image)
						if err != nil && hasAny {
							if trouble == "felling" {

							} else if trouble == "burning" {

							}
						}
					}
				default:
				}
			}
		}
	}()

	// Block main from exiting
	select {}
}
