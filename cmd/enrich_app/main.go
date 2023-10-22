package main

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"

	"enrichment-service/internal/api"
	"enrichment-service/internal/repository"
	"enrichment-service/internal/service"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := sql.Open("postgres", os.Getenv("DB_CONN_STRING"))
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	repo := repository.NewPersonRepository(db)
	serv := service.NewPersonService(repo)
	apis := api.NewPersonAPI(serv)

	http.HandleFunc("/add", apis.AddPersonHandler)
	http.HandleFunc("/get", apis.GetPersonsHandler)
	http.HandleFunc("/delete", apis.DeletePersonHandler)
	http.HandleFunc("/update", apis.UpdatePersonHandler)

	port := os.Getenv("API_PORT")
	log.Printf("Server started on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
