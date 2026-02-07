package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/metaclaw/metaclaw-registry/internal/httpapi"
	"github.com/metaclaw/metaclaw-registry/internal/service"
	"github.com/metaclaw/metaclaw-registry/internal/store"
)

func main() {
	var addr string
	var dataPath string
	var adminToken string
	flag.StringVar(&addr, "addr", ":8088", "listen address")
	flag.StringVar(&dataPath, "data", "./data/registry.json", "registry data file")
	flag.StringVar(&adminToken, "admin-token", "", "admin bearer token for write API")
	flag.Parse()

	if strings.TrimSpace(adminToken) == "" {
		adminToken = os.Getenv("REGISTRY_ADMIN_TOKEN")
	}

	st, err := store.NewFileStore(dataPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	svc := service.New(st)
	h := httpapi.New(svc, adminToken)

	log.Printf("metaclaw-registry listening on %s", addr)
	if adminToken == "" {
		log.Printf("write API auth: disabled")
	} else {
		log.Printf("write API auth: enabled")
	}
	if err := http.ListenAndServe(addr, h.Routes()); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
