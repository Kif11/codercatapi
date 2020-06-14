package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var db *mongo.Database

type applicantRequest struct {
	Email     string     `json:"email"`
	Questions []question `json:"questions"`
}

type question struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// find takes a slice and looks for an element in it. If found it will
// return it's key, otherwise it will return -1 and a bool of false.
func find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func validateEmail(email string) error {
	var rxEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	if len(email) > 254 || !rxEmail.MatchString(email) {
		return fmt.Errorf("%s is not a valid email", email)
	}

	return nil
}

func decodeBody(body io.ReadCloser, dst interface{}) error {
	if body == http.NoBody {
		return fmt.Errorf("please send a request body")
	}
	if err := json.NewDecoder(body).Decode(dst); err != nil {
		return fmt.Errorf("error decoding request body. %v", err)
	}

	return nil
}

func subscribeHandler(w http.ResponseWriter, r *http.Request) {
	applicantReq := &applicantRequest{}

	if err := decodeBody(r.Body, applicantReq); err != nil {
		returnError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := validateEmail(applicantReq.Email); err != nil {
		returnError(w, http.StatusInternalServerError, err.Error())
		return
	}

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err := db.Collection("applicant").InsertOne(ctx, applicantReq)
	if err != nil {
		returnError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func returnError(w http.ResponseWriter, header int, msg string) {
	payload := errorResponse{Error: msg}

	js, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("error marshaling error msg. %v\n", err)
	}

	w.Header().Set("Content-Type", "application/json;")
	w.WriteHeader(header)
	w.Write(js)
}

var allowedOrigins = []string{"http://localhost:3000", "https://codercatclub.github.io"}

func corsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		_, found := find(allowedOrigins, r.Header.Get("Origin"))
		if !found {
			// Do not attach CORS header if origin is not allowed
			h.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func main() {
	var err error

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}

	db = client.Database("codercat")

	r := mux.NewRouter()

	r.HandleFunc("/v1/subscribe", subscribeHandler).Methods("POST", "OPTIONS")

	r.Use(corsMiddleware)

	srvAddress := "127.0.0.1:9000"

	srv := &http.Server{
		Handler:      r,
		Addr:         srvAddress,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Printf("Starting server on %s\n", srvAddress)

	log.Fatal(srv.ListenAndServe())
}
