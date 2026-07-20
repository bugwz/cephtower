package main

import (
	"flag"
	"log"
	"net/http"

	"cephtower/backend/internal/api"
	"cephtower/backend/internal/config"
	"cephtower/backend/internal/integrations/ceph"
	"cephtower/backend/internal/store"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "Path to the YAML configuration file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := store.Open(cfg.Database)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}

	cephClient := ceph.NewDashboardClient(cfg.Ceph)
	server := api.NewServer(cfg, cephClient, db)
	defer func() {
		if err := server.Close(); err != nil {
			log.Printf("close database: %v", err)
		}
	}()

	log.Printf("cephtower database engine: %s", cfg.Database.Engine)
	log.Printf("cephtower backend listening on %s", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, server.Routes()); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
