package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Post struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

func postsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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
		if err := json.NewDecoder(r.Body).Decode(&newPost); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err := conn.QueryRow(context.Background(),
			"INSERT INTO posts (title, body) VALUES ($1, $2) RETURNING id",
			newPost.Title, newPost.Body,
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
	w.Header().Set("Content-Type", "application/json")
	idStr := r.URL.Path[len("/posts/"):]
	id := 0
	fmt.Sscanf(idStr, "%d", &id)
	if r.Method == http.MethodDelete {
		_, err := conn.Exec(context.Background(), "DELETE FROM posts WHERE id = $1", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
