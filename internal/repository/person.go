package repository

import (
	"database/sql"
	"enrichment-service/internal/entity"
	log "github.com/sirupsen/logrus"
)

type PersonRepository struct {
	db *sql.DB
}

func NewPersonRepository(db *sql.DB) *PersonRepository {
	return &PersonRepository{db: db}
}

// Save сохраняет обогащенное сообщение в БД
func (r *PersonRepository) Save(person entity.Person) error {
	exists, err := r.ExistsByUniqueData(person)
	if err != nil {
		log.WithError(err).Error("Error checking unique data for person")
		return err
	}
	if exists {
		log.Warn("Person already exists")
		return sql.ErrNoRows // или любая другая ошибка, указывающая на дубликат
	}

	query := `INSERT INTO persons(name, surname, patronymic, age, gender, nationality) VALUES($1, $2, $3, $4, $5, $6)`
	_, err = r.db.Exec(query, person.Name, person.Surname, person.Patronymic, person.Age, person.Gender, person.Nationality)
	if err != nil {
		log.WithError(err).Error("Error saving person to DB")
	}
	return err
}

// Get возвращает данные по фильтрам и пагинации
func (r *PersonRepository) Get(limit int, offset int) ([]entity.Person, error) {
	query := `SELECT id, name, surname, patronymic, age, gender, nationality FROM persons LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		log.WithError(err).Error("Error fetching persons from DB")
		return nil, err
	}
	defer rows.Close()

	var persons []entity.Person
	for rows.Next() {
		var p entity.Person
		if err := rows.Scan(&p.ID, &p.Name, &p.Surname, &p.Patronymic, &p.Age, &p.Gender, &p.Nationality); err != nil {
			log.WithError(err).Error("Error scanning person data")
			return nil, err
		}
		persons = append(persons, p)
	}

	return persons, rows.Err()
}

// DeleteByID удаляет запись по идентификатору
func (r *PersonRepository) DeleteByID(id int) error {
	exists, err := r.ExistsByID(id)
	if err != nil {
		log.WithError(err).Error("Error checking if person exists by ID")
		return err
	}
	if !exists {
		log.Warnf("No person found with ID: %d", id)
		return sql.ErrNoRows // ошибка, если запись не найдена
	}

	query := `DELETE FROM persons WHERE id = $1`
	_, err = r.db.Exec(query, id)
	if err != nil {
		log.WithError(err).Errorf("Error deleting person with ID: %d", id)
	}

	return err
}

// Update обновляет сущность в БД
func (r *PersonRepository) Update(person entity.Person) error {
	exists, err := r.ExistsByID(person.ID)
	if err != nil {
		log.WithError(err).Error("Error checking if person exists by ID for update")
		return err
	}
	if !exists {
		log.Warnf("No person found with ID: %d for update", person.ID)
		return sql.ErrNoRows // ошибка, если запись не найдена
	}

	query := `UPDATE persons SET name=$1, surname=$2, patronymic=$3, age=$4, gender=$5, nationality=$6 WHERE id=$7`
	_, err = r.db.Exec(query, person.Name, person.Surname, person.Patronymic, person.Age, person.Gender, person.Nationality, person.ID)
	if err != nil {
		log.WithError(err).Error("Error updating person in DB")
	}
	return err
}

// ExistsByID проверяет, существует ли запись с переданным ID
func (r *PersonRepository) ExistsByID(id int) (bool, error) {
	var exists bool
	query := `SELECT exists(SELECT 1 FROM persons WHERE id=$1)`
	err := r.db.QueryRow(query, id).Scan(&exists)
	if err != nil {
		log.WithError(err).Errorf("Error checking if person exists by ID: %d", id)
	}
	return exists, err
}

// ExistsByUniqueData проверяет, существует ли запись с такими данными
func (r *PersonRepository) ExistsByUniqueData(person entity.Person) (bool, error) {
	var exists bool
	query := `SELECT exists(SELECT 1 FROM persons WHERE
                name=$1 AND 
                surname=$2 AND 
                patronymic=$3 AND 
				age=$4 AND 
				gender=$5 AND 
				nationality=$6)`
	err := r.db.QueryRow(query, person.Name, person.Surname, person.Patronymic, person.Age, person.Gender, person.Nationality).Scan(&exists)
	if err != nil {
		log.WithError(err).Errorf("Error checking unique data for person: %s %s", person.Name, person.Surname)
	}
	return exists, err
}
