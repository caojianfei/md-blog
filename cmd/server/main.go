package main

import (
	"flag"
	"log"

	"github.com/cybernote/md-blog/internal/bootstrap"
	"github.com/cybernote/md-blog/internal/config"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to YAML config file")
	flag.StringVar(&configPath, "c", "", "shorthand for --config")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	app, err := bootstrap.New(cfg)
	if err != nil {
		log.Fatalf("bootstrap failed: %v", err)
	}

	log.Printf("md-blog listening on %s", cfg.App.Addr)
	if err := app.Start(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
