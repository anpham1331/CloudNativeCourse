package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var collection *mongo.Collection

func main() {
	// Connect to MongoDB
	ctx := context.Background()
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	// Ping the MongoDB
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")

	// Set up the collection
	collection = client.Database("testdb").Collection("items")

	// HTTP Server
	mux := http.NewServeMux()
	mux.HandleFunc("/list", listHandler)
	mux.HandleFunc("/price", priceHandler)
	mux.HandleFunc("/create", createHandler)
	mux.HandleFunc("/update", updateHandler)
	mux.HandleFunc("/delete", deleteHandler)
	log.Fatal(http.ListenAndServe("localhost:8000", mux))
}

type Item struct {
	Name  string  `json:"name"`
	Price float32 `json:"price"`
}

func listHandler(w http.ResponseWriter, req *http.Request) {
	cursor, err := collection.Find(context.Background(), nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error listing items: %s", err)
		return
	}
	defer cursor.Close(context.Background())

	// Check if cursor is nil
	if cursor == nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error listing items: cursor is nil")
		return
	}

	items := []Item{}
	// Iterate over the cursor
	for cursor.Next(context.Background()) {
		var item Item
		// Decode each document
		if err := cursor.Decode(&item); err != nil {
			log.Println("Error decoding item:", err)
			continue // Skip this item and continue with the next one
		}
		items = append(items, item)
	}
	// Check for cursor error after iterating
	if err := cursor.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error listing items: %s", err)
		return
	}

	// Encode the items to JSON and send the response
	json.NewEncoder(w).Encode(items)
}

func priceHandler(w http.ResponseWriter, req *http.Request) {
	itemName := req.URL.Query().Get("item")

	var item Item
	err := collection.FindOne(context.Background(), Item{Name: itemName}).Decode(&item)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "No such item: %s", itemName)
		return
	}

	json.NewEncoder(w).Encode(item)
}

func createHandler(w http.ResponseWriter, req *http.Request) {
	itemName := req.URL.Query().Get("item")
	priceStr := req.URL.Query().Get("price")
	price, err := strconv.ParseFloat(priceStr, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid price: %s", err)
		return
	}

	item := Item{Name: itemName, Price: float32(price)}
	_, err = collection.InsertOne(context.Background(), item)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error creating item: %s", err)
		return
	}

	json.NewEncoder(w).Encode(item)
}

func updateHandler(w http.ResponseWriter, req *http.Request) {
	itemName := req.URL.Query().Get("item")
	priceStr := req.URL.Query().Get("price")
	price, err := strconv.ParseFloat(priceStr, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid price: %s", err)
		return
	}

	filter := Item{Name: itemName}
	update := Item{Name: itemName, Price: float32(price)}

	result, err := collection.ReplaceOne(context.Background(), filter, update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error updating item: %s", err)
		return
	}

	if result.ModifiedCount == 0 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "No such item: %s", itemName)
		return
	}

	json.NewEncoder(w).Encode(update)
}

func deleteHandler(w http.ResponseWriter, req *http.Request) {
	itemName := req.URL.Query().Get("item")

	result, err := collection.DeleteOne(context.Background(), Item{Name: itemName})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error deleting item: %s", err)
		return
	}

	if result.DeletedCount == 0 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "No such item: %s", itemName)
		return
	}

	fmt.Fprintf(w, "Item %s deleted", itemName)
}
