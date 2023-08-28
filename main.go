package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

type CatFactWorker struct {
	client *mongo.Client
}

type Server struct {
	client *mongo.Client
}

func ServerConstructor(c *mongo.Client) *Server {
	return &Server{
		client: c,
	}
}

func (s *Server) handleGetAllFacts(w http.ResponseWriter, r *http.Request) {
	coll := s.client.Database("catfact").Collection("facts")
	query := bson.M{}
	cursor, err := coll.Find(context.TODO(), query)
	if err != nil {
		log.Fatal(err)
	}

	results := []bson.M{}

	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)

}

func CatFactWorkerConstructor(c *mongo.Client) *CatFactWorker {
	return &CatFactWorker{
		client: c,
	}
}

func (cfw *CatFactWorker) start() error {
	coll := cfw.client.Database("catfact").Collection("facts")
	ticker := time.NewTicker(2 * time.Second)
	for {
		resp, err := http.Get("https://catfact.ninja/fact")
		if err != nil {
			return err
		}

		var catFact bson.M

		if err := json.NewDecoder(resp.Body).Decode(&catFact); err != nil {
			return err
		}

		fmt.Println(catFact)

		_, errs := coll.InsertOne(context.TODO(), catFact)
		if errs != nil {
			return errs
		}

		<-ticker.C
	}

}

func main() {

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("").SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	worker := CatFactWorkerConstructor(client)

	go worker.start()

	server := ServerConstructor(client)
	http.HandleFunc("/facts", server.handleGetAllFacts)
	http.ListenAndServe(":3000", nil)
}
