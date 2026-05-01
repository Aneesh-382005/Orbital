package main

import (
	"log"
	
	"github.com/Aneesh-382005/Orbital/internal/api"
	"github.com/Aneesh-382005/Orbital/internal/provisioner"
	"github.com/Aneesh-382005/Orbital/internal/store"
)

func main() {
	s := store.New()

	p, err := provisioner.NewDockerProvisioner()
	if err != nil {
		log.Fatalf("failed to create provisioner: %v", err)
	}

	h := api.NewHandler(s, p)
	r := api.NewRouter(h)
	r.Run(":8080")
}
