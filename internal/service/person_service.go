// internal/service/person_service.go
package service

import (
	"context"
	"fmt"

	"people-enricher/internal/client"
	"people-enricher/internal/domain"
	"people-enricher/internal/repository"

	"github.com/sirupsen/logrus"
)

type personService struct {
	repo     repository.PersonRepo
	enricher *client.Enricher
	log      *logrus.Logger
}

func NewPersonService(repo repository.PersonRepo, enricher *client.Enricher, log *logrus.Logger) domain.PersonService {
	return &personService{
		repo:     repo,
		enricher: enricher,
		log:      log,
	}
}

func (s *personService) Create(ctx context.Context, input *domain.Person) (*domain.Person, error) {
	s.log.WithFields(logrus.Fields{
		"name":       input.Name,
		"surname":    input.Surname,
		"patronymic": input.Patronymic,
	}).Info("Creating person")

	// Получаем обогащённые данные по имени.
	enrichedResult, err := s.enricher.EnrichPerson(ctx, input.Name)
	if err != nil {
		s.log.WithError(err).Error("Failed to enrich person data")

	}

	person := &domain.Person{
		Name:       input.Name,
		Surname:    input.Surname,
		Patronymic: input.Patronymic,
	}

	if enrichedResult != nil {
		if enrichedResult.Age != nil {
			person.Age = enrichedResult.Age
		}
		if enrichedResult.Gender != nil {
			person.Gender = enrichedResult.Gender
		}
		if enrichedResult.Nationality != nil {
			person.Nationality = enrichedResult.Nationality
		}

		if enrichedResult.NationalityProbability != nil {
			person.NationalityProbability = ptrString(fmt.Sprintf("%.2f", *enrichedResult.NationalityProbability))
		}
	}

	createdPerson, err := s.repo.Create(ctx, person)
	if err != nil {
		s.log.WithError(err).Error("Failed to create person in DB")
		return nil, err
	}

	s.log.WithField("id", createdPerson.ID).Info("Successfully created person")
	return createdPerson, nil
}

func (s *personService) GetById(ctx context.Context, id int64) (*domain.Person, error) {
	s.log.WithField("id", id).Info("Fetching person by ID")
	person, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"id":    id,
			"error": err,
		}).Error("Failed to fetch person")
		return nil, err
	}
	s.log.WithField("id", id).Info("Successfully fetched person")
	return person, nil
}

func (s *personService) Update(ctx context.Context, person *domain.Person) (*domain.Person, error) {
	s.log.WithFields(logrus.Fields{
		"id":         person.ID,
		"name":       person.Name,
		"surname":    person.Surname,
		"patronymic": person.Patronymic,
	}).Info("Updating person")

	existing, err := s.repo.GetByID(ctx, person.ID)
	if err != nil {
		s.log.WithFields(logrus.Fields{"id": person.ID, "error": err}).Error("Person not found")
		return nil, err
	}
	_ = existing

	enrichedResult, err := s.enricher.EnrichPerson(ctx, person.Name)
	if err != nil {
		s.log.WithError(err).Error("Failed to enrich updated person data")
	} else if enrichedResult != nil {
		if enrichedResult.Age != nil {
			person.Age = enrichedResult.Age
		}
		if enrichedResult.Gender != nil {
			person.Gender = enrichedResult.Gender
		}
		if enrichedResult.Nationality != nil {
			person.Nationality = enrichedResult.Nationality
		}
		if enrichedResult.NationalityProbability != nil {
			person.NationalityProbability = ptrString(fmt.Sprintf("%.2f", *enrichedResult.NationalityProbability))
		}
	}

	updated, err := s.repo.Update(ctx, person)
	if err != nil {
		s.log.WithError(err).Error("Failed to update person in DB")
		return nil, err
	}

	s.log.WithField("id", person.ID).Info("Successfully updated person")
	return updated, nil
}

func (s *personService) Delete(ctx context.Context, id int64) error {
	s.log.WithField("id", id).Info("Deleting person")

	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.WithFields(logrus.Fields{"id": id, "error": err}).Error("Person not found")
		return err
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		s.log.WithError(err).Error("Failed to delete person")
		return err
	}
	s.log.WithField("id", id).Info("Successfully deleted person")
	return nil
}

func ptrString(s string) *string {
	return &s
}

func (s *personService) List(ctx context.Context, filter *domain.PersonFilter) ([]*domain.Person, int, error) {

	s.log.Infof("got request: %+v", filter)

	persons, total, err := s.repo.List(ctx, filter)
	if err != nil {
		s.log.Errorf("error to getting: %v", err)
		return nil, 0, err
	}

	return persons, total, nil
}
