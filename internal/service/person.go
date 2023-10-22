package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"enrichment-service/internal/entity"
	"enrichment-service/internal/repository"
)

type PersonService struct {
	repo       *repository.PersonRepository
	httpClient *http.Client
}

func NewPersonService(repo *repository.PersonRepository) *PersonService {
	return &PersonService{
		repo: repo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *PersonService) EnrichAndSave(person entity.Person) error {
	person, err := s.enrichPerson(person)
	if err != nil {
		log.WithError(err).Error("Failed to enrich person")
		return err
	}
	return s.repo.Save(person)
}

func (s *PersonService) GetPersons(limit int, offset int) ([]entity.Person, error) {
	return s.repo.Get(limit, offset)
}

func (s *PersonService) DeletePersonByID(id int) error {
	return s.repo.DeleteByID(id)
}

func (s *PersonService) UpdatePerson(person entity.Person) error {
	return s.repo.Update(person)
}

func (s *PersonService) fetchAndDecodeJSON(url string, target interface{}) error {
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received HTTP %d from %s", resp.StatusCode, url)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func (s *PersonService) enrichPerson(person entity.Person) (entity.Person, error) {
	// Возраст
	var ageResponse struct {
		Age int `json:"age"`
	}
	if err := s.fetchAndDecodeJSON("https://api.agify.io/?name="+person.Name, &ageResponse); err != nil {
		log.WithError(err).Error("Failed to fetch and decode age data")
		return person, err
	}
	person.Age = ageResponse.Age

	// Гендер
	var genderResponse struct {
		Gender string `json:"gender"`
	}
	if err := s.fetchAndDecodeJSON("https://api.genderize.io/?name="+person.Name, &genderResponse); err != nil {
		log.WithError(err).Error("Failed to fetch and decode gender data")
		return person, err
	}
	person.Gender = genderResponse.Gender

	// Национальность
	var nationalityResponse struct {
		Country []struct {
			CountryID   string  `json:"country_id"`
			Probability float64 `json:"probability"`
		}
	}
	if err := s.fetchAndDecodeJSON("https://api.nationalize.io/?name="+person.Name, &nationalityResponse); err != nil {
		log.WithError(err).Error("Failed to fetch and decode nationality data")
		return person, err
	}
	if len(nationalityResponse.Country) > 0 {
		person.Nationality = nationalityResponse.Country[0].CountryID
	}

	return person, nil
}
