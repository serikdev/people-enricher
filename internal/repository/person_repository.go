package repository

import (
	"context"
	"errors"
	"fmt"
	"people-enricher/internal/domain"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type PersonRepo struct {
	pool   *pgxpool.Pool
	logger *logrus.Logger
}

func NewPersonRepo(pool *pgxpool.Pool, logger *logrus.Logger) *PersonRepo {
	return &PersonRepo{
		pool:   pool,
		logger: logger,
	}
}

func (r *PersonRepo) Create(ctx context.Context, person *domain.Person) (*domain.Person, error) {
	logger := r.logger.WithField("operation", "Create")
	logger.Debug("Creating new record about person")

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

	row := r.pool.QueryRow(
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

func (r *PersonRepo) Update(ctx context.Context, person *domain.Person) (*domain.Person, error) {
	logger := r.logger.WithField("operation", "Update").WithField("person_id", person.ID)
	logger.Debug("Update person")

	query := `
		UPDATE people
		SET name = $1,
			surname = $2,
			patronymic = $3,
			age = $4,
			gender = $5,
			nationality = $6,
			nationality_probability = $7,
			updated_at = $8
		WHERE id = $9
		RETURNING id, name, surname, patronymic, age, gender, nationality, nationality_probability, created_at, updated_at
	`
	person.UpdatedAt = time.Now()

	row := r.pool.QueryRow(
		ctx,
		query,
		person.Name,
		person.Surname,
		person.Patronymic,
		person.Age,
		person.Gender,
		person.Nationality,
		person.NationalityProbability,
		person.UpdatedAt,
		person.ID,
	)

	var updatedPerson domain.Person
	err := row.Scan(
		&updatedPerson.ID,
		&updatedPerson.Name,
		&updatedPerson.Surname,
		&updatedPerson.Patronymic,
		&updatedPerson.Age,
		&updatedPerson.Gender,
		&updatedPerson.Nationality,
		&updatedPerson.NationalityProbability,
		&updatedPerson.CreatedAt,
		&updatedPerson.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.WithError(err).Warn("Person not found")
			return nil, errors.New("person not found")
		}
		logger.WithError(err).Error("error update person")
		return nil, fmt.Errorf("update record about person: %w", err)
	}

	logger.Info("Successfully updated person")
	return &updatedPerson, nil
}
func (r *PersonRepo) Delete(ctx context.Context, id int64) error {
	logger := r.logger.WithField("operation", "Delete").WithField("person_id", id)
	logger.Debug("Remove person ")

	query := "DELETE FROM people WHERE id = $1"

	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		logger.WithError(err).Error("Error delete person")
		return fmt.Errorf("removing person: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		logger.Warn("Person not found")
		return errors.New("person not found")
	}

	logger.Info("Successfully remove person")
	return nil
}

func (r *PersonRepo) GetByID(ctx context.Context, id int64) (*domain.Person, error) {
	logger := r.logger.WithField("operation", "GetByID").WithField("person_id", id)
	logger.Debug("Получение записи о человеке по ID")

	query := `
        SELECT id, name, surname, patronymic, age, gender, nationality, nationality_probability, created_at, updated_at
        FROM people
        WHERE id = $1
    `

	row := r.pool.QueryRow(ctx, query, id)

	var person domain.Person
	err := row.Scan(
		&person.ID,
		&person.Name,
		&person.Surname,
		&person.Patronymic,
		&person.Age,
		&person.Gender,
		&person.Nationality,
		&person.NationalityProbability,
		&person.CreatedAt,
		&person.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.WithError(err).Warn("Person not found")
			return nil, fmt.Errorf("person not found with ID %d", id)
		}
		logger.WithError(err).Error("Error executing query or scanning result")
		return nil, fmt.Errorf("error getting person with ID %d: %w", id, err)
	}

	logger.Info("Successfully retrieved person")
	return &person, nil
}

func (r *PersonRepo) List(ctx context.Context, filter *domain.PersonFilter) ([]*domain.Person, int, error) {
	logger := r.logger.WithField("operation", "List")
	logger.WithField("filter", filter).Debug("Getting person list")

	whereConditions := []string{}
	args := []interface{}{}
	argCounter := 1

	if filter.Name != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("name ILIKE $%d", argCounter))
		args = append(args, "%"+*filter.Name+"%")
		argCounter++
	}

	if filter.Surname != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("surname ILIKE $%d", argCounter))
		args = append(args, "%"+*filter.Surname+"%")
		argCounter++
	}

	if filter.Patronymic != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("patronymic ILIKE $%d", argCounter))
		args = append(args, "%"+*filter.Patronymic+"%")
		argCounter++
	}

	if filter.Gender != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("gender = $%d", argCounter))
		args = append(args, *filter.Gender)
		argCounter++
	}

	if filter.AgeFrom != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("age >= $%d", argCounter))
		args = append(args, *filter.AgeFrom)
		argCounter++
	}

	if filter.AgeTo != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("age <= $%d", argCounter))
		args = append(args, *filter.AgeTo)
		argCounter++
	}

	if filter.Nationality != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("nationality = $%d", argCounter))
		args = append(args, *filter.Nationality)
		argCounter++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM people %s", whereClause)

	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		logger.WithError(err).Error("Error getting all lists")
		return nil, 0, fmt.Errorf("getting all lists: %w", err)
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 10
	}

	offset := (filter.Page - 1) * filter.PageSize

	query := fmt.Sprintf(`
		SELECT id, name, surname, patronymic, age, gender, nationality, nationality_probability, created_at, updated_at
		FROM people
		%s
		ORDER BY id DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCounter, argCounter+1)

	args = append(args, filter.PageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		logger.WithError(err).Error("Error getting lists")
		return nil, 0, fmt.Errorf("getting lists: %w", err)
	}
	defer rows.Close()

	people := []*domain.Person{}
	for rows.Next() {
		var person domain.Person
		err := rows.Scan(
			&person.ID,
			&person.Name,
			&person.Surname,
			&person.Patronymic,
			&person.Age,
			&person.Gender,
			&person.Nationality,
			&person.NationalityProbability,
			&person.CreatedAt,
			&person.UpdatedAt,
		)
		if err != nil {
			logger.WithError(err).Error("Error scanning rows")
			return nil, 0, fmt.Errorf("scannig rows: %w", err)
		}
		people = append(people, &person)
	}

	if err := rows.Err(); err != nil {
		logger.WithError(err).Error("Ошибка при обработке строк")
		return nil, 0, fmt.Errorf("обработка строк: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"total":  total,
		"count":  len(people),
		"page":   filter.Page,
		"limit":  filter.PageSize,
		"offset": offset,
	}).Info("Successfull geting person")

	return people, total, nil
}
