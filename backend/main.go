package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	initDB()
	initJWTSecret()

	mux := http.NewServeMux()

	// Auth
	mux.HandleFunc("/auth/register", registerHandler)
	mux.HandleFunc("/auth/login", loginHandler)
	mux.HandleFunc("/auth/refresh", refreshHandler)
	mux.HandleFunc("/auth/logout", logoutHandler)

	// Posts
	mux.HandleFunc("/posts", postsHandler)
	mux.HandleFunc("/posts/", authMiddleware(postByIDHandler))

	// Products
	mux.HandleFunc("/products", productsHandler)
	mux.HandleFunc("/products/", authMiddleware(productByIDHandler))

	// Cart — todo protegido
	mux.HandleFunc("/cart", authMiddleware(cartRouter))
	mux.HandleFunc("/cart/", authMiddleware(cartRouter))

	handler := corsMiddleware(mux)

	log.Println("server running at http://localhost:8080")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
