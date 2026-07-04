package main

import (
	"context"
	"fmt"
	"log"

	"personal-digital-inventory/internal/infrastructure/database"

	"gorm.io/gorm"
)

type App struct {
	ctx context.Context
	db  *gorm.DB
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	dbCfg := database.Config{
		Host:     "localhost",
		Port:     5432,
		User:     "inventory",
		Password: "",
		DBName:   "personal_digital_inventory",
		SSLMode:  "disable",
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
