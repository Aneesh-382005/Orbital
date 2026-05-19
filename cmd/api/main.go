package main

import (
	"log"

	"github.com/Aneesh-382005/Orbital/internal/api"
	"github.com/Aneesh-382005/Orbital/internal/auth"
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

	j, err := auth.NewJWTService("secrets/private.pem", "secrets/public.pem")
	if err != nil {
		log.Fatalf("failed to load JWT keys: %v", err)
	}

	h := api.NewHandler(s, p, j)
	r := api.NewRouter(h, j)
	r.Run(":8080")
}
