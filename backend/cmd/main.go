package main

import (
	"flag"
	"log"
	"net/http"

	"cephtower/backend/internal/ceph"
	"cephtower/backend/internal/config"
	"cephtower/backend/internal/httpapi"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "Path to the YAML configuration file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	cephClient := ceph.NewDashboardClient(cfg.Ceph)
	server := httpapi.NewServer(cfg, cephClient)

	log.Printf("cephtower backend listening on %s", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, server.Routes()); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
