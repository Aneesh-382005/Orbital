package main

import (
	"log"

	"github.com/Aneesh-382005/Orbital/internal/api"
	"github.com/Aneesh-382005/Orbital/internal/db"
	"github.com/Aneesh-382005/Orbital/internal/provisioner"
	"github.com/Aneesh-382005/Orbital/internal/store"
)

func main() {
	database, err := db.New("orbital.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	s := store.New(database)

	p, err := provisioner.NewDockerProvisioner(s)
	if err != nil {
		log.Fatalf("failed to create provisioner: %v", err)
	}

	h := api.NewHandler(s, p)
	r := api.NewRouter(h)
	r.Run(":8080")
}
