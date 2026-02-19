package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

var conn *pgx.Conn

func initDB() {
	godotenv.Load()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL not set")
	}
	var err error
	conn, err = pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	log.Println("Connected to PostgreSQL")
}

// Obtiene o crea el carrito del usuario
func getOrCreateCart(userID int) (int, error) {
	var cartID int
	err := conn.QueryRow(context.Background(),
		"SELECT id FROM carts WHERE user_id = $1", userID,
	).Scan(&cartID)

	if err != nil {
		// No existe, lo creamos
		err = conn.QueryRow(context.Background(),
			"INSERT INTO carts (user_id) VALUES ($1) RETURNING id", userID,
		).Scan(&cartID)
		if err != nil {
			return 0, err
		}
	}
	return cartID, nil
}
