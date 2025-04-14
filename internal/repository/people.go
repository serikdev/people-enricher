package repository

import (
	"context"
	"fmt"
	"people-enricher/internal/domain"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type PersonRepo struct {
	poll   *pgxpool.Pool
	logger *logrus.Logger
}

func NewPersonRepo(pool *pgxpool.Pool, logger *logrus.Logger) *PersonRepo {
	return &PersonRepo{
		poll:   pool,
		logger: logger,
	}
}

func (r *PersonRepo) Create(ctx context.Context, person *domain.Person) (*domain.Person, error) {
	logger := r.logger.WithField("operation", "Create")
	logger.Debug("Creating a new record about people")

	query := `
        INSERT INTO people(
            name, surname, patronymic, age, gender, nationality, nationality_probability, created_at, updated_at
        )   VALUES(
            $1, $2, $3, $4, $5, $6, $7, $8, $9   
        )
            RETURNING id, name, surname, patronymic, age, gender, nationality, nationality_probability, created_at, updated_at
    `
	now := time.Now()
	person.CreatedAt = now
	person.UpdatedAt = now

	row := r.poll.QueryRow(
		ctx,
		query,
		person.Name,
		person.Surname,
		person.Patronymic,
		person.Age,
		person.Gender,
		person.Nationality,
		person.NationalityProbability,
		person.CreatedAt,
		person.UpdatedAt,
	)

	var ceatedPerson domain.Person
	err := row.Scan(
		&ceatedPerson.ID,
		&ceatedPerson.Name,
		&ceatedPerson.Surname,
		&ceatedPerson.Patronymic,
		&ceatedPerson.Age,
		&ceatedPerson.Gender,
		&ceatedPerson.Nationality,
		&ceatedPerson.NationalityProbability,
		&ceatedPerson.CreatedAt,
		&ceatedPerson.UpdatedAt,
	)
	if err != nil {
		logger.WithError(err).Error("Failed Creating record about person")
		return nil, fmt.Errorf("creating record about person: %w", err)
	}
	logger.WithField("person_id", ceatedPerson.ID).Info("Successfull creating person")
	return &ceatedPerson, nil
}
