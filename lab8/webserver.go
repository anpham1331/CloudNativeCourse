package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Initialize MongoDB client
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	// Initialize MongoDB collection
	collection := client.Database("mydb").Collection("products")

	// Initialize HTTP server mux
	mux := http.NewServeMux()
	mux.HandleFunc("/list", listHandler(collection))
	mux.HandleFunc("/price", priceHandler(collection))
	mux.HandleFunc("/create", createHandler(collection))
	mux.HandleFunc("/update", updateHandler(collection))
	mux.HandleFunc("/delete", deleteHandler(collection))

	log.Fatal(http.ListenAndServe(":8000", mux))
}

type Product struct {
	Name  string  `bson:"name"`
	Price float32 `bson:"price"`
}

func listHandler(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		cursor, err := collection.Find(context.Background(), nil)
		if err != nil {
			handleError(w, err)
			return
		}
		defer cursor.Close(context.Background())

		for cursor.Next(context.Background()) {
			var product Product
			err := cursor.Decode(&product)
			if err != nil {
				handleError(w, err)
				return
			}
			fmt.Fprintf(w, "%s: $%.2f\n", product.Name, product.Price)
		}
	}
}
func priceHandler(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		item := req.URL.Query().Get("item")
		var product Product
		err := collection.FindOne(context.Background(), bson.M{"name": item}).Decode(&product)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, "no such item: %q\n", item)
				return
			}
			handleError(w, err)
			return
		}
		fmt.Fprintf(w, "%s: $%.2f\n", product.Name, product.Price)
	}
}

func createHandler(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		item := req.URL.Query().Get("item")
		priceStr := req.URL.Query().Get("price")
		price, err := strconv.ParseFloat(priceStr, 32)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "invalid price: %s\n", err)
			return
		}

		product := Product{Name: item, Price: float32(price)}
		_, err = collection.InsertOne(context.Background(), product)
		if err != nil {
			handleError(w, err)
			return
		}
		fmt.Fprintf(w, "item %s created with price $%.2f\n", item, price)
	}
}

func updateHandler(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		item := req.URL.Query().Get("item")
		priceStr := req.URL.Query().Get("price")
		price, err := strconv.ParseFloat(priceStr, 32)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "invalid price: %s\n", err)
			return
		}

		update := bson.M{"$set": bson.M{"price": float32(price)}}
		result, err := collection.UpdateOne(context.Background(), bson.M{"name": item}, update)
		if err != nil {
			handleError(w, err)
			return
		}

		if result.MatchedCount == 0 {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "no such item: %s\n", item)
			return
		}

		fmt.Fprintf(w, "item %s updated with price $%.2f\n", item, price)
	}
}

func deleteHandler(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		item := req.URL.Query().Get("item")
		result, err := collection.DeleteOne(context.Background(), bson.M{"name": item})
		if err != nil {
			handleError(w, err)
			return
		}

		if result.DeletedCount == 0 {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "no such item: %s\n", item)
			return
		}

		fmt.Fprintf(w, "item %s deleted\n", item)
	}
}

func handleError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "internal server error: %s\n", err)
}
