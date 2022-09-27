package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"gopkg.in/mgo.v2/bson"
)

// type result struct {
// 	title string
// 	completedAt time.Time
// }

func listTasks(w http.ResponseWriter, r *http.Request) {
	// Cnnect to MongoDB
	uri := os.Getenv("MONGO_URI")
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("Error at connecting to DB: ", err)
	}

	// Ping the primary
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully connected and pinged.")

	// Query to DB
	collection := client.Database("todo").Collection("tasks")
	cur, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal("Error at query to db: ", err)
	}
	defer cur.Close(context.Background())

	var results []bson.M
	if err = cur.All(context.Background(), &results); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Succeeded in query")

	// Write to response as JSON
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(http.StatusOK)
	
	// respond empty slice when no results
	if len(results) == 0 {
		fmt.Fprint(w, make([]string, 0))
		return
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(results); err != nil {
		log.Fatal(err)
	}
	fmt.Fprint(w, buf.String())
	// buf.WriteTo(w)
}

func main() {
	// Load dotenv
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Register handler
	http.HandleFunc("/tasks", listTasks)

	// Start server
	log.Fatal(http.ListenAndServe(":8080", nil))
}
