package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type Comment struct {
	ID      primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Time    time.Time          `json:"_time,omitempty" bson:"time"`
	Comment string             `json:"comment" bson:"comment"`
}

const RESTFULAPI_PORT = "1313"
const FORUM_DATABASE_NAME = "forum"
const COMMENT_COLLECTION_NAME = "comments"
const MONGO_URI = "mongodb://mongo:27017"
const MONGO_URI_LOCAL = "mongodb://192.168.1.3:27017"

var client *mongo.Client

func main() {
	// Set client options
	clientOptions := options.Client().ApplyURI(MONGO_URI)

	// Connect to MongoDB
	var err error
	client, err = mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	// Initialize router
	router := mux.NewRouter()
	handler := cors.Default().Handler(router)

	// Route handles & endpoints
	router.HandleFunc("/"+COMMENT_COLLECTION_NAME, GetComments).Methods("GET")
	router.HandleFunc("/"+COMMENT_COLLECTION_NAME, CreateComment).Methods("POST")
	router.HandleFunc("/"+COMMENT_COLLECTION_NAME+"/{id}", GetComment).Methods("GET")
	router.HandleFunc("/"+COMMENT_COLLECTION_NAME+"/{id}", UpdateComment).Methods("PATCH")
	router.HandleFunc("/"+COMMENT_COLLECTION_NAME+"/{id}", DeleteComment).Methods("DELETE")

	// Start server
	log.Fatal(http.ListenAndServe(":"+RESTFULAPI_PORT, handler))
}

// GetComments retrieves all comments
func GetComments(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get All")

	// Define the collection
	collection := client.Database(FORUM_DATABASE_NAME).Collection(COMMENT_COLLECTION_NAME)

	// Find all documents in the collection
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	var comments []Comment
	// Iterate through the cursor to retrieve all documents
	for cursor.Next(context.TODO()) {
		var comment Comment
		cursor.Decode(&comment)
		comments = append(comments, comment)
	}

	json.NewEncoder(w).Encode(comments)
}

// GetComment retrieves a single comment by id
func GetComment(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get")
	params := mux.Vars(r)

	// Convert the ID string to a MongoDB ObjectID
	id, _ := primitive.ObjectIDFromHex(params["id"])

	// Create a filter to find the specific comment
	filter := bson.M{"_id": id}

	// Find the comment in the "comments" collection
	var result Comment
	err := client.Database(FORUM_DATABASE_NAME).Collection(COMMENT_COLLECTION_NAME).FindOne(context.TODO(), filter).Decode(&result)

	if err != nil {
		// Return an empty comment if the ID is not found
		json.NewEncoder(w).Encode(&Comment{})
	} else {
		json.NewEncoder(w).Encode(result)
	}
}

// CreateComment creates a new comment
func CreateComment(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Post")
	var comment Comment
	_ = json.NewDecoder(r.Body).Decode(&comment)
	comment.ID = primitive.NewObjectID()
	comment.Time = time.Now()

	// Add the comment to the MongoDB database
	collection := client.Database(FORUM_DATABASE_NAME).Collection(COMMENT_COLLECTION_NAME)
	_, err := collection.InsertOne(context.TODO(), comment)
	if err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(comment)
}

// UpdateComment updates an existing comment by id
func UpdateComment(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Patch")
	params := mux.Vars(r)
	var comment Comment
	_ = json.NewDecoder(r.Body).Decode(&comment)

	// Define the collection
	collection := client.Database(FORUM_DATABASE_NAME).Collection(COMMENT_COLLECTION_NAME)

	// Convert the id string to an ObjectID
	id, _ := primitive.ObjectIDFromHex(params["id"])

	// Create the update filter
	filter := bson.M{"_id": id}

	// Create the update document
	update := bson.M{"$set": bson.M{"comment": comment.Comment}}

	// Update the document in the collection
	_, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal(err)
	}

	// Find the updated document
	var updatedComment Comment
	err = collection.FindOne(context.TODO(), filter).Decode(&updatedComment)
	if err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(updatedComment)
}

// DeleteComment deletes a comment by id
func DeleteComment(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delete")
	params := mux.Vars(r)

	// Get the collection
	collection := client.Database(FORUM_DATABASE_NAME).Collection(COMMENT_COLLECTION_NAME)

	// Convert the ID from hex string to an ObjectID
	id, _ := primitive.ObjectIDFromHex(params["id"])

	// Delete the document with the matching ID
	_, err := collection.DeleteOne(context.TODO(), bson.M{"_id": id})
	if err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode("Comment Deleted")
}
