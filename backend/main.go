package main
import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"github.com/jackc/pgx/v5"
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

	// if r.Method == http.MethodGet {
	// 	json.NewEncoder(w).Encode(posts)
	// 	return
	// }

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

	// if r.Method == http.MethodPost {
	// 	var newPost Post

	// 	err := json.NewDecoder(r.Body).Decode(&newPost)
	// 	if err != nil {
	// 		http.Error(w, err.Error(), http.StatusBadRequest)
	// 		return
	// 	}

	// 	newPost.ID = len(posts) + 1
	// 	posts = append(posts, newPost)

	// 	w.WriteHeader(http.StatusCreated)
	// 	json.NewEncoder(w).Encode(newPost)
	// 	return
	// }

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

func initDB() {
	var err error

	conn, err = pgx.Connect(context.Background(),
		"postgres://postgres:password@localhost:5432/blog")

	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	log.Println("Connected to PostgreSQL")
}


func main() {
	initDB()
	http.HandleFunc("/posts", postsHandler)

	log.Println("server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
