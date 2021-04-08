package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	clientOptions := options.Client().ApplyURI("mongodb://mongo:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	pointers := client.Database("bf").Collection("pointers")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		result := struct{ Pointer int }{}
		err := pointers.FindOne(r.Context(), bson.M{"session": r.Header.Get("X-Session")}).Decode(&result)
		if err != nil && err != mongo.ErrNoDocuments {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "%d", result.Pointer)
	})
	http.HandleFunc("/right", func(w http.ResponseWriter, r *http.Request) {
		_, err := pointers.UpdateOne(r.Context(), bson.M{"session": r.Header.Get("X-Session")},
			bson.D{{"$inc", bson.D{{"pointer", 1}}}},
			options.Update().SetUpsert(true))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	http.HandleFunc("/left", func(w http.ResponseWriter, r *http.Request) {
		_, err := pointers.UpdateOne(r.Context(), bson.M{"session": r.Header.Get("X-Session")},
			bson.D{{"$inc", bson.D{{"pointer", -1}}}},
			options.Update().SetUpsert(true))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
