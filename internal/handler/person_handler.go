// handler/person_handler.go
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"people-enricher/internal/domain"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

type PersonHandler struct {
	service domain.PersonService
	log     *logrus.Logger
}

func NewPersonHandler(service domain.PersonService, log *logrus.Logger) *PersonHandler {
	return &PersonHandler{
		service: service,
		log:     log,
	}
}

func extractIDFromURL(path string) (int, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return 0, errors.New("invalid URL format")
	}
	return strconv.Atoi(parts[len(parts)-1])
}

func (h *PersonHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input domain.PersonInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.log.WithError(err).Debug("Error decoding request body")
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if input.Name == "" || input.Surname == "" {
		h.log.Debug("Invalid input: name or surname is empty")
		respondWithError(w, http.StatusBadRequest, "Name and surname are required")
		return
	}

	h.log.WithFields(logrus.Fields{
		"name":       input.Name,
		"surname":    input.Surname,
		"patronymic": input.Patronymic,
	}).Debug("Creating new person")

	person := &domain.Person{
		Name:       input.Name,
		Surname:    input.Surname,
		Patronymic: input.Patronymic,
	}

	createdPerson, err := h.service.Create(r.Context(), person)
	if err != nil {
		h.log.WithError(err).Error("Error creating person")
		respondWithError(w, http.StatusInternalServerError, "Error creating person")
		return
	}

	h.log.WithField("id", createdPerson.ID).Info("Person created successfully")
	respondWithJSON(w, http.StatusCreated, createdPerson)
}

func (h *PersonHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := extractIDFromURL(r.URL.Path)
	if err != nil {
		h.log.WithError(err).Debug("Invalid ID parameter")
		respondWithError(w, http.StatusBadRequest, "Invalid person ID")
		return
	}

	var input domain.Person
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.log.WithError(err).Debug("Error decoding request body")
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if input.Name == "" || input.Surname == "" {
		h.log.Debug("Invalid input: name or surname is empty")
		respondWithError(w, http.StatusBadRequest, "Name and surname are required")
		return
	}

	h.log.WithFields(logrus.Fields{
		"id":         id,
		"name":       input.Name,
		"surname":    input.Surname,
		"patronymic": input.Patronymic,
	}).Debug("Updating person")

	person, err := h.service.Update(r.Context(), &input)
	if err != nil {
		h.log.WithFields(logrus.Fields{
			"id":    id,
			"error": err,
		}).Error("Error updating person")
		respondWithError(w, http.StatusInternalServerError, "Error updating person")
		return
	}

	h.log.WithField("id", id).Info("Person updated successfully")
	respondWithJSON(w, http.StatusOK, person)
}

func (h *PersonHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := extractIDFromURL(r.URL.Path)
	if err != nil {
		h.log.WithError(err).Debug("Invalid ID parameter")
		respondWithError(w, http.StatusBadRequest, "Invalid person ID")
		return
	}

	h.log.WithField("id", id).Debug("Deleting person")

	if err := h.service.Delete(r.Context(), int64(id)); err != nil {
		h.log.WithFields(logrus.Fields{
			"id":    id,
			"error": err,
		}).Error("Error deleting person")
		respondWithError(w, http.StatusInternalServerError, "Error deleting person")
		return
	}

	h.log.WithField("id", id).Info("Person deleted successfully")
	w.WriteHeader(http.StatusNoContent)
}

type PaginatedResponse struct {
	Data       []domain.Person `json:"data"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, ErrorResponse{Error: message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {

		logrus.WithError(err).Error("Failed to marshal JSON response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
func (h *PersonHandler) GetByID(ctx context.Context, id int64) (*domain.Person, error) {
	logger := h.log.WithFields(logrus.Fields{
		"operation": "GetByID",
		"person_id": id,
	})

	logger.Debug("Fetching person by ID")

	person, err := h.service.GetById(ctx, id)
	if err != nil {
		// Логируем ошибку
		if err.Error() == "person not found" {
			logger.Warn("Person not found")
			return nil, fmt.Errorf("person not found with id %d", id)
		}
		logger.Error("Error fetching person:", err)
		return nil, fmt.Errorf("error fetching person with id %d: %w", id, err)
	}

	logger.Info("Successfully fetched person")
	return person, nil
}
func (h *PersonHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	filter := &domain.PersonFilter{}

	if name := query.Get("name"); name != "" {
		filter.Name = &name
	}
	if surname := query.Get("surname"); surname != "" {
		filter.Surname = &surname
	}
	if patronymic := query.Get("patronymic"); patronymic != "" {
		filter.Patronymic = &patronymic
	}

	if ageMinStr := query.Get("age_min"); ageMinStr != "" {
		if ageMin, err := strconv.Atoi(ageMinStr); err == nil {
			filter.AgeFrom = &ageMin
		}
	}
	if ageMaxStr := query.Get("age_max"); ageMaxStr != "" {
		if ageMax, err := strconv.Atoi(ageMaxStr); err == nil {
			filter.AgeTo = &ageMax
		}
	}
	page := 1
	if p := query.Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	pageSize := 10
	if ps := query.Get("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}
	filter.Page = page
	filter.PageSize = pageSize

	h.log.WithFields(logrus.Fields{
		"name":       filter.Name,
		"surname":    filter.Surname,
		"patronymic": filter.Patronymic,
		"age_min":    filter.AgeFrom,
		"age_max":    filter.AgeTo,
		"page":       filter.Page,
		"page_size":  filter.PageSize,
	}).Debug("Listing persons with filter")

	persons, total, err := h.service.List(r.Context(), filter)
	if err != nil {
		h.log.WithError(err).Error("Error listing persons")
		respondWithError(w, http.StatusInternalServerError, "Error listing persons")
		return
	}

	totalPages := (total + pageSize - 1) / pageSize

	response := PaginatedResponse{
		Data:       toFlatList(persons),
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	respondWithJSON(w, http.StatusOK, response)
}
func toFlatList(persons []*domain.Person) []domain.Person {
	result := make([]domain.Person, len(persons))
	for i, p := range persons {
		result[i] = *p
	}
	return result
}
