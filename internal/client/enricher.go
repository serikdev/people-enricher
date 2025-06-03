package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"people-enricher/internal/config"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Enricher struct {
	httpClient   *http.Client
	logger       *logrus.Entry
	agifyURL     string
	genderizeURL string
	nationalize  string
}

type AgifyResponse struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Count int    `json:"count"`
}

type GenderizeResponse struct {
	Name        string  `json:"name"`
	Gender      string  `json:"gender"`
	Probability float64 `json:"probability"`
	Count       int     `json:"count"`
}

type NationalizeResponse struct {
	Name    string `json:"name"`
	Country []struct {
		CountryID   string  `json:"country_id"`
		Probability float64 `json:"probability"`
	} `json:"country"`
}

type EnrichmentResult struct {
	Age                    *int     `json:"age,omitempty"`
	Gender                 *string  `json:"gender,omitempty"`
	Nationality            *string  `json:"nationality,omitempty"`
	NationalityProbability *float64 `json:"nationality_probability,omitempty"`
}

func NewEnricher(cfg config.ExternalAPIConfig, logger *logrus.Entry) *Enricher {

	return &Enricher{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger:       logger,
		agifyURL:     cfg.AgifyURL,
		genderizeURL: cfg.GenderizeURL,
		nationalize:  cfg.NationalizeURL,
	}

}

func (e *Enricher) EnrichPerson(ctx context.Context, name string) (*EnrichmentResult, error) {
	e.logger.WithField("name", name).Debug("Starting enrich for name data")

	result := &EnrichmentResult{}

	age, err := e.getAge(ctx, name)
	if err != nil {
		e.logger.WithError(err).Warn("Error getting age")
	} else {
		result.Age = &age
		e.logger.WithField("age", age).Debug("Success got age")
	}

	gender, err := e.getGender(ctx, name)
	if err != nil {
		e.logger.WithError(err).Warn("Error getting gender")
	} else {
		result.Gender = &gender
		e.logger.WithField("gender", gender).Debug("Success got gender")
	}

	nationality, probability, err := e.getNationality(ctx, name)
	if err != nil {
		e.logger.WithError(err).Warn("Error getting nationality")
	} else {
		result.Nationality = &nationality
		result.NationalityProbability = &probability
		e.logger.WithFields(logrus.Fields{
			"nationality": nationality,
			"probability": probability,
		}).Debug("Success got nationality")
	}

	return result, nil
}

func (e *Enricher) getAge(ctx context.Context, name string) (int, error) {

	url := fmt.Sprintf("%s?name=%s", e.agifyURL, name)

	e.logger.WithField("url", url).Debug("request to API agify.io")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, errors.Wrap(err, "creating request to API agify.io")
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return 0, errors.Wrap(err, "request to API agify.io")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("status code of API agify.io: %d", resp.StatusCode)
	}

	var agifyResp AgifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&agifyResp); err != nil {
		return 0, errors.Wrap(err, "decode response of API agify.io")
	}

	return agifyResp.Age, nil
}

func (e *Enricher) getGender(ctx context.Context, name string) (string, error) {
	url := fmt.Sprintf("%s?name=%s", e.genderizeURL, name)

	e.logger.WithField("url", url).Debug("request to genderize.io")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", errors.Wrap(err, "creating request to API genderize.io")
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "request to API genderize.io")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code of API genderize.io: %d", resp.StatusCode)
	}

	var genderizeResp GenderizeResponse
	if err := json.NewDecoder(resp.Body).Decode(&genderizeResp); err != nil {
		return "", errors.Wrap(err, "decode response of API genderize.io")
	}

	return genderizeResp.Gender, nil
}

func (e *Enricher) getNationality(ctx context.Context, name string) (string, float64, error) {
	url := fmt.Sprintf("%s?name=%s", e.nationalize, name)

	e.logger.WithField("url", url).Debug("request to API nationalize.io")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", 0, errors.Wrap(err, "creating requset to API nationalize.io")
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return "", 0, errors.Wrap(err, "request to API nationalize.io")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("status code of API nationalize.io: %d", resp.StatusCode)
	}

	var nationalizeResp NationalizeResponse
	if err := json.NewDecoder(resp.Body).Decode(&nationalizeResp); err != nil {
		return "", 0, errors.Wrap(err, "decode reponse of API nationalize.io")
	}

	if len(nationalizeResp.Country) == 0 {
		return "", 0, errors.New("no data about of nationality")
	}

	countryID := nationalizeResp.Country[0].CountryID
	probability := nationalizeResp.Country[0].Probability

	return countryID, probability, nil
}

// func sanitizeName(name string) string {
// 	// Удаляем пробелы и лишние слэши
// 	name = strings.TrimSpace(name)
// 	name = strings.Trim(name, "/")
// 	return name
// }
