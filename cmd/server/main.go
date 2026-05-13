package main

import (
    "log"

    "github.com/cybernote/md-blog/internal/bootstrap"
    "github.com/cybernote/md-blog/internal/config"
)

func main() {
    cfg := config.Load()
    app, err := bootstrap.New(cfg)
    if err != nil {
        log.Fatalf("bootstrap failed: %v", err)
    }

    log.Printf("md-blog listening on %s", cfg.App.Addr)
    if err := app.Start(); err != nil {
        log.Fatalf("server stopped: %v", err)
    }
}
