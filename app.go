package main

import (
	"context"
	"fmt"
	"log"

	"personal-digital-inventory/internal/infrastructure/config"
	"personal-digital-inventory/internal/infrastructure/database"

	"gorm.io/gorm"
)

type App struct {
	ctx context.Context
	db  *gorm.DB
	cfg *config.Config
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Printf("WARNING: Failed to load config: %v", err)
		log.Println("The application will start with limited functionality.")
		return
	}
	a.cfg = cfg

	dbCfg := database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
	}

	db, err := database.Connect(dbCfg)
	if err != nil {
		log.Printf("WARNING: Database connection failed: %v", err)
		log.Println("The application will start without database connectivity.")
		return
	}

	a.db = db
	log.Println("Database connected successfully.")
}

func (a *App) shutdown(ctx context.Context) {
	if a.db != nil {
		if err := database.Close(a.db); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}
}

func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
