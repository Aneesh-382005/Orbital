package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Aneesh-382005/Orbital/internal/api"
	"github.com/Aneesh-382005/Orbital/internal/auth"
	"github.com/Aneesh-382005/Orbital/internal/db"
	"github.com/Aneesh-382005/Orbital/internal/provisioner"
	"github.com/Aneesh-382005/Orbital/internal/reconciler"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rec := reconciler.New(s, p, 30*time.Second)
	go rec.Start(ctx)

	h := api.NewHandler(s, p, j)
	r := api.NewRouter(h, j)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("shutting down ...")
		cancel()
	}()

	r.Run(":8080")
}
