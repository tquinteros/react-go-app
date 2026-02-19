package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Product struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	Images      []string `json:"images"`
	Discount    float64  `json:"discount"`
}

func productsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		rows, err := conn.Query(context.Background(),
			"SELECT id, name, description, price, images, discount FROM products ORDER BY id DESC")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var products []Product
		for rows.Next() {
			var p Product
			rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Images, &p.Discount)
			products = append(products, p)
		}
		json.NewEncoder(w).Encode(products)
		return
	}
	if r.Method == http.MethodPost {
		var newProduct Product
		if err := json.NewDecoder(r.Body).Decode(&newProduct); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err := conn.QueryRow(context.Background(),
			"INSERT INTO products (name, description, price, images, discount) VALUES ($1, $2, $3, $4, $5) RETURNING id",
			newProduct.Name, newProduct.Description, newProduct.Price, newProduct.Images, newProduct.Discount,
		).Scan(&newProduct.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newProduct)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func productByIDHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.URL.Path[len("/products/"):]
	id := 0
	fmt.Sscanf(idStr, "%d", &id)
	if r.Method == http.MethodDelete {
		_, err := conn.Exec(context.Background(), "DELETE FROM products WHERE id = $1", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
