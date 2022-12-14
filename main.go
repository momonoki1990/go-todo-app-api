package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

type Task struct {
	ID          string    `json:"_id"`
	Title       string    `json:"title"`
	CompletedAt time.Time `json:"completedAt"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Cnnect to MongoDB and return client
func connectToDB() (*mongo.Client, error) {
	uri := os.Getenv("MONGO_URI")
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	return client, err
}

// List all tasks
func listTasks(w http.ResponseWriter, r *http.Request) {
	// Cnnect to MongoDB
	client, err := connectToDB()
	if err != nil {
		log.Println("Error at connect to db: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer client.Disconnect(context.TODO())

	// Query to DB
	coll := client.Database("todo").Collection("tasks")
	cur, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Println("Error at query to db: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer cur.Close(context.Background())

	var results []bson.M
	if err = cur.All(context.Background(), &results); err != nil {
		log.Println("Error at get data by cur.All(): ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println("Succeeded in query")

	// Respond JSON
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	// respond empty slice when no results
	if len(results) == 0 {
		fmt.Fprint(w, make([]string, 0))
		return
	}

	if json.NewEncoder(w).Encode(results); err != nil {
		log.Println("Error at JSON decode of query resutls: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

type TaskCreateRequest struct {
	Title string `json:"title"`
}

// Create task
func createTask(w http.ResponseWriter, r *http.Request) {
	// Get POST body data
	var task TaskCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		log.Println("Error at JSON decode of POST body data: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Printf("%+v\n", task)

	// Connect to DB
	client, err := connectToDB()
	if err != nil {
		log.Println("Error at connect to db: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer client.Disconnect(context.TODO())

	// Create task
	coll := client.Database("todo").Collection("tasks")
	coll.InsertOne(context.TODO(), bson.M{
		"title":       task.Title,
		"completedAt": nil,
		"createdAt":   time.Now(),
		"updatedAt":   time.Now(),
	})

	w.WriteHeader(http.StatusOK)
}

type TaskUpdateRequest struct {
	Title       string    `json:"title"`
	CompletedAt time.Time `json:"completedAt"`
}

// Update Task
func updateTask(w http.ResponseWriter, r *http.Request) {
	// Get id from URL path
	id := path.Base(r.URL.Path)
	log.Println("Requested id:", id)
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("Error at converting id to objectId ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get POST body data
	var task TaskUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		log.Println("Error at JSON decode of POST body data: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Printf("%+v\n", task)

	// Connect to DB
	client, err := connectToDB()
	if err != nil {
		log.Println("Error at connect to db: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer client.Disconnect(context.TODO())

	// Find task(check task presence)
	coll := client.Database("todo").Collection("tasks")
	var _task bson.M
	if err := coll.FindOne(context.TODO(), bson.M{"_id": oid}).Decode(&task); err != nil {
		// NOTE: including ErrNoDocuments here
		log.Println("Error at find task ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("%+v\n", _task)

	// Update task
	result, err := coll.UpdateByID(context.TODO(), oid, bson.M{"$set": bson.M{"title": task.Title, "completedAt": task.CompletedAt, "updatedAt": time.Now()}})
	if result.ModifiedCount == 0 {
		log.Println("Updated count is 0")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("%+v", result)

	w.WriteHeader(http.StatusOK)
}

// Delete task
func deletTask(w http.ResponseWriter, r *http.Request) {
	// Get id from URL path
	id := path.Base(r.URL.Path)
	log.Println("Requested id:", id)
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("Error at converting id to objectId ", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	// Connect to DB
	client, err := connectToDB()
	if err != nil {
		log.Println("Error at connect to db: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer client.Disconnect(context.TODO())

	// Find task(check task presence)
	coll := client.Database("todo").Collection("tasks")
	var task bson.M
	if err := coll.FindOne(context.TODO(), bson.M{"_id": oid}).Decode(&task); err != nil {
		// NOTE: including ErrNoDocuments here
		log.Println("Error at find task ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("%+v\n", task)

	// Delete task
	result, err := coll.DeleteOne(context.TODO(), bson.M{"_id": oid})
	if err != nil {
		log.Println("Error at delete task ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if result.DeletedCount == 0 {
		log.Println("Deleted count is 0")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("%+v", result)

	w.WriteHeader(http.StatusOK)
}

// Handle request to '/tasks'
func handleTaskRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		listTasks(w, r)
	case "POST":
		createTask(w, r)
	case "PUT":
		updateTask(w, r)
	case "DELETE":
		deletTask(w, r)
	default:
		w.WriteHeader(405)
	}
}

func main() {
	// Load dotenv
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Register handler
	http.HandleFunc("/tasks/", handleTaskRequest)

	// Start server
	log.Fatal(http.ListenAndServe(":8080", nil))
}
