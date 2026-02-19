package main
import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"github.com/jackc/pgx/v5"
	"os"
	"fmt"
	"github.com/joho/godotenv"
)

var conn *pgx.Conn

type Post struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

var posts = []Post{
	{ID: 1, Title: "Hola Go", Body: "Mi primer backend"},
}

func corsHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func postsHandler(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w)
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method == http.MethodGet {
		rows, err := conn.Query(context.Background(),
			"SELECT id, title, body FROM posts ORDER BY id DESC")
	
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
	
		var posts []Post
	
		for rows.Next() {
			var p Post
			rows.Scan(&p.ID, &p.Title, &p.Body)
			posts = append(posts, p)
		}
	
		json.NewEncoder(w).Encode(posts)
		return
	}

	if r.Method == http.MethodPost {
		var newPost Post
	
		err := json.NewDecoder(r.Body).Decode(&newPost)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	
		err = conn.QueryRow(
			context.Background(),
			"INSERT INTO posts (title, body) VALUES ($1, $2) RETURNING id",
			newPost.Title,
			newPost.Body,
		).Scan(&newPost.ID)
	
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newPost)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func postByIDHandler(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w)
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	idStr := r.URL.Path[len("/posts/"):]
	id := 0
	fmt.Sscanf(idStr, "%d", &id)

	if r.Method == http.MethodDelete {
		_, err := conn.Exec(context.Background(),
			"DELETE FROM posts WHERE id = $1", id)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func initDB() {
	var err error
	err = godotenv.Load()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	conn, err = pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	log.Println("Connected to PostgreSQL")
}


func main() {
	initDB()
	http.HandleFunc("/posts", postsHandler)
	http.HandleFunc("/posts/", postByIDHandler)
	log.Println("server running at http://localhost:8080")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
