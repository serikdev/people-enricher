package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"people-enricher/internal/adapter/repository"
	"people-enricher/internal/client"
	"people-enricher/internal/config"
	"people-enricher/internal/handler"
	"people-enricher/internal/service"
	"people-enricher/pkg/database"
	"people-enricher/pkg/logger"
	"strconv"
	"strings"
	"time"

	_ "people-enricher/docs"

	"github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           People Information API
// @version         1.0
// @description     API for managing and enriching people data
// @host            localhost:8080
// @BasePath        /

func main() {

	type key string

	const (
		personKey key = "personID"
	)
	log := logger.NewLogger(os.Getenv("LOG_LEVEL"))

	cfg, err := config.LoadCfg(".env")
	if err != nil {
		log.WithError(err).Fatal("Error loading .env file")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dbpool, err := database.NewPool(ctx, &cfg.DBConfig, &logrus.Logger{})
	if err != nil {
		log.WithError(err).Fatal("Failed to connect database")
	}
	defer dbpool.Close()

	repo := repository.NewPersonRepo(dbpool, log)
	enricherService := client.NewEnricher(cfg.ExternalAPI, log)
	personService := service.NewPersonService(*repo, enricherService, log)
	personHandler := handler.NewPersonHandler(personService, log)

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
			r = r.WithContext(context.WithValue(r.Context(), personKey, id))
			personHandler.Update(w, r)
		case http.MethodDelete:
			r = r.WithContext(context.WithValue(r.Context(), personKey, id))
			personHandler.Delete(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	port := 8080
	log.Infof("Starting server on :%d", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
	if err != nil {
		log.WithError(err).Fatal("Server failed")
	}
}
