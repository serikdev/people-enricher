// В main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"people-enricher/internal/client"
	"people-enricher/internal/handler"
	"people-enricher/internal/repository"
	"people-enricher/internal/service"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {

	logger := logrus.New()

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	connString := os.Getenv("DATABASE_URL")
	fmt.Println("DATABASE_URL from .env file:")

	dbpool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		logger.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbpool.Close()

	repo := repository.NewPersonRepo(dbpool, logger)

	enricherService := client.NewEnricher(logger)

	personService := service.NewPersonService(*repo, enricherService, logger)

	personHandler := handler.NewPersonHandler(personService, logger)

	http.HandleFunc("/persons", personHandler.Create)
	http.HandleFunc("/persons/", func(w http.ResponseWriter, r *http.Request) {

		idStr := r.URL.Path[len("/persons/"):]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		person, err := personHandler.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Fetched person: %+v\n", person)
	})

	http.HandleFunc("/persons/update/", personHandler.Update)
	http.HandleFunc("/persons/delete/", personHandler.Delete)

	port := 8080
	logger.Infof("Starting server on :%d", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		logger.WithError(err).Fatal("Server failed")
	}
}
