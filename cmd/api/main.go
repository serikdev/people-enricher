package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"people-enricher/internal/client"
	"people-enricher/internal/handler"
	"people-enricher/internal/repository"
	"people-enricher/internal/service"
	"strconv"
	"strings"

	_ "people-enricher/docs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           People Information API
// @version         1.0
// @description     API for managing and enriching people data
// @host            localhost:8080
// @BasePath        /

func main() {
	logger := logrus.New()

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	connString := os.Getenv("DATABASE_URL")
	fmt.Println("Connecting to database:", connString)

	dbpool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		logger.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbpool.Close()

	repo := repository.NewPersonRepo(dbpool, logger)
	enricherService := client.NewEnricher(logger)
	personService := service.NewPersonService(*repo, enricherService, logger)
	personHandler := handler.NewPersonHandler(personService, logger)

	mux := http.NewServeMux()

	// Swagger
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	mux.HandleFunc("/persons", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			personHandler.Create(w, r)
		case http.MethodGet:
			personHandler.List(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/persons/", func(w http.ResponseWriter, r *http.Request) {

		path := strings.TrimPrefix(r.URL.Path, "/persons/")
		id, err := strconv.ParseInt(path, 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID in URL", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			person, err := personHandler.GetByID(r.Context(), id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(person)
		case http.MethodPut:
			r = r.WithContext(context.WithValue(r.Context(), "personID", id))
			personHandler.Update(w, r)
		case http.MethodDelete:
			r = r.WithContext(context.WithValue(r.Context(), "personID", id))
			personHandler.Delete(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	port := 8080
	logger.Infof("Starting server on :%d", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
	if err != nil {
		logger.WithError(err).Fatal("Server failed")
	}
}
