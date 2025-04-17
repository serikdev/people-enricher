package entity

import (
	"context"
	"time"
)

// Person represents enriched person data in database
// @Description Information about a person
type Person struct {
	ID                     int64     `json:"id"`
	Name                   string    `json:"name"`
	Surname                string    `json:"surname"`
	Patronymic             *string   `json:"patronymic,omitempty"`
	Age                    *int      `json:"age,omitempty"`
	Gender                 *string   `json:"gender,omitempty"`
	Nationality            *string   `json:"nationality,omitempty"`
	NationalityProbability *string   `json:"nationality_probability,omitempty"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type PersonFilter struct {
	Name        *string `json:"name,omitempty"`
	Surname     *string `json:"surname,omitempty"`
	Patronymic  *string `json:"patronymic,omitempty"`
	Gender      *string `json:"gender,omitempty"`
	AgeFrom     *int    `json:"age_from,omitempty"`
	AgeTo       *int    `json:"age_to,omitempty"`
	Nationality *string `json:"nationality,omitempty"`
	Page        int     `json:"page"`
	PageSize    int     `json:"page_size"`
}

type PersonInput struct {
	Name       string  `json:"name"`
	Surname    string  `json:"surname"`
	Patronymic *string `json:"patronymic,omitempty"`
}

type PersonService interface {
	Create(ctx context.Context, person *Person) (*Person, error)
	Update(ctx context.Context, person *Person) (*Person, error)
	Delete(ctx context.Context, id int64) error
	GetById(ctx context.Context, id int64) (*Person, error)
	List(ctx context.Context, filter *PersonFilter) ([]*Person, int, error)
}
