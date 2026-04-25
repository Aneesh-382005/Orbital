package main

import (
	"github.com/Aneesh-382005/Orbital/internal/api"
	"github.com/Aneesh-382005/Orbital/internal/store"
)

func main() {
	s := store.New()
	h := api.NewHandler(s)
	r := api.NewRouter(h)
	r.Run(":8080")
}
