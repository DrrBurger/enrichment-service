package api

import (
	"database/sql"
	"encoding/json"
	"enrichment-service/internal/entity"
	"enrichment-service/internal/service"
	"errors"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

type PersonAPI struct {
	service *service.PersonService
}

func NewPersonAPI(service *service.PersonService) *PersonAPI {
	return &PersonAPI{service: service}
}

func (api *PersonAPI) AddPersonHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	personData := &entity.Person{}
	if err := json.NewDecoder(r.Body).Decode(personData); err != nil {
		log.Println("Failed to decode request body:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(personData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := api.service.EnrichAndSave(*personData)
	if err != nil {
		log.Println("Failed to add person:", err)
		http.Error(w, "Error adding person", http.StatusInternalServerError)
		return
	}

	log.Println("Successfully added person:", personData.Name)
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte("Person added successfully"))
	if err != nil {
		return
	}
}

func (api *PersonAPI) GetPersonsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	limit := 10
	offset := 0

	if l, ok := r.URL.Query()["limit"]; ok {
		parsedLimit, err := strconv.Atoi(l[0])
		if err == nil {
			limit = parsedLimit
		}
	}

	if o, ok := r.URL.Query()["offset"]; ok {
		parsedOffset, err := strconv.Atoi(o[0])
		if err == nil {
			offset = parsedOffset
		}
	}

	persons, err := api.service.GetPersons(limit, offset)
	if err != nil {
		log.Println("Failed to fetch persons:", err)
		http.Error(w, "Error fetching persons", http.StatusInternalServerError)
		return
	}

	responseData, err := json.Marshal(persons)
	if err != nil {
		log.Println("Failed to marshal response:", err)
		http.Error(w, "Error marshalling response", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully fetched %d persons", len(persons))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(responseData)
	if err != nil {
		return
	}
}

func (api *PersonAPI) DeletePersonHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	idStr, ok := r.URL.Query()["id"]
	if !ok || len(idStr[0]) < 1 {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr[0])
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	err = api.service.DeletePersonByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Error ID not found", http.StatusNotFound)
			return
		}

		log.Printf("Failed to delete person with ID %d: %v", id, err)
		http.Error(w, "Error deleting person", http.StatusInternalServerError)
		return
	}
	log.Printf("Successfully deleted person with ID %d", id)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Person delete successfully"))
	if err != nil {
		return
	}
}

func (api *PersonAPI) UpdatePersonHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	personData := &entity.Person{}
	if err := json.NewDecoder(r.Body).Decode(personData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := api.service.UpdatePerson(*personData)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Error ID not found", http.StatusNotFound)
			return
		}
		log.Println("Failed to update person:", err)
		http.Error(w, "Error updating person", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully updated person: %s %s", personData.Name, personData.Surname)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Person updated successfully"))
	if err != nil {
		return
	}
}
