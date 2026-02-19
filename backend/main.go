package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

var conn *pgx.Conn
var jwtSecret []byte

// ─── STRUCTS ───────────────────────────────────────────────

type CartItem struct {
	ID        int      `json:"id"`
	ProductID int      `json:"product_id"`
	Quantity  int      `json:"quantity"`
	Name      string   `json:"name"`
	Price     float64  `json:"price"`
	Images    []string `json:"images"`
	Discount  float64  `json:"discount"`
}

type Cart struct {
	ID     int        `json:"id"`
	UserID int        `json:"user_id"`
	Items  []CartItem `json:"items"`
}

type Post struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

type Product struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	Images      []string `json:"images"`
	Discount    float64  `json:"discount"`
}

type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	User        User   `json:"user"`
}

// ─── HELPERS ───────────────────────────────────────────────

func getUserIDFromToken(r *http.Request) (int, error) {
	authHeader := r.Header.Get("Authorization")
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}
	claims := token.Claims.(jwt.MapClaims)
	return int(claims["user_id"].(float64)), nil
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

// ─── MIDDLEWARE ────────────────────────────────────────────

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func generateAccessToken(userID int, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func generateRefreshToken(userID int, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func authCookie(w http.ResponseWriter, value string, maxAge int) {
	secure := os.Getenv("COOKIE_SECURE") != "false"
	sameSite := http.SameSiteLaxMode
	if secure {
		sameSite = http.SameSiteNoneMode
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    value,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Path:     "/",
		MaxAge:   maxAge,
	})
}

// ─── AUTH HANDLERS ─────────────────────────────────────────

func registerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "error hashing password", http.StatusInternalServerError)
		return
	}
	var user User
	err = conn.QueryRow(context.Background(),
		"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, email",
		req.Email, string(hash),
	).Scan(&user.ID, &user.Email)
	if err != nil {
		http.Error(w, "email already exists", http.StatusConflict)
		return
	}
	accessToken, err := generateAccessToken(user.ID, user.Email)
	if err != nil {
		http.Error(w, "error generating token", http.StatusInternalServerError)
		return
	}
	refreshToken, _ := generateRefreshToken(user.ID, user.Email)
	authCookie(w, refreshToken, 7*24*60*60)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{AccessToken: accessToken, User: user})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	var user User
	var passwordHash string
	err := conn.QueryRow(context.Background(),
		"SELECT id, email, password_hash FROM users WHERE email = $1", req.Email,
	).Scan(&user.ID, &user.Email, &passwordHash)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	accessToken, err := generateAccessToken(user.ID, user.Email)
	if err != nil {
		http.Error(w, "error generating token", http.StatusInternalServerError)
		return
	}
	refreshToken, _ := generateRefreshToken(user.ID, user.Email)
	authCookie(w, refreshToken, 7*24*60*60)
	json.NewEncoder(w).Encode(AuthResponse{AccessToken: accessToken, User: user})
}

func refreshHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "no refresh token", http.StatusUnauthorized)
		return
	}
	token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}
	claims := token.Claims.(jwt.MapClaims)
	userIDVal, ok := claims["user_id"]
	if !ok || userIDVal == nil {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}
	userID := int(userIDVal.(float64))
	var email string
	if emailVal, ok := claims["email"]; ok && emailVal != nil {
		email = emailVal.(string)
	} else {
		if err := conn.QueryRow(context.Background(),
			"SELECT email FROM users WHERE id = $1", userID).Scan(&email); err != nil {
			http.Error(w, "invalid refresh token", http.StatusUnauthorized)
			return
		}
	}
	accessToken, err := generateAccessToken(userID, email)
	if err != nil {
		http.Error(w, "error generating token", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"access_token": accessToken})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	authCookie(w, "", -1)
	w.WriteHeader(http.StatusNoContent)
}

// ─── CART HANDLERS ─────────────────────────────────────────

// GET /cart → devuelve el carrito completo con items y datos del producto
func getCartHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	cartID, err := getOrCreateCart(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := conn.Query(context.Background(), `
		SELECT ci.id, ci.product_id, ci.quantity,
		       p.name, p.price, p.images, p.discount
		FROM cart_items ci
		JOIN products p ON p.id = ci.product_id
		WHERE ci.cart_id = $1
		ORDER BY ci.id ASC
	`, cartID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []CartItem
	for rows.Next() {
		var item CartItem
		rows.Scan(&item.ID, &item.ProductID, &item.Quantity,
			&item.Name, &item.Price, &item.Images, &item.Discount)
		items = append(items, item)
	}

	if items == nil {
		items = []CartItem{}
	}

	json.NewEncoder(w).Encode(Cart{ID: cartID, UserID: userID, Items: items})
}

// POST /cart/items → agregar producto al carrito (o sumar quantity si ya existe)
func addCartItemHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		ProductID int `json:"product_id"`
		Quantity  int `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ProductID == 0 {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.Quantity <= 0 {
		body.Quantity = 1
	}

	cartID, err := getOrCreateCart(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Si el producto ya está en el carrito, suma la quantity
	var item CartItem
	err = conn.QueryRow(context.Background(), `
		INSERT INTO cart_items (cart_id, product_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (cart_id, product_id)
		DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity
		RETURNING id, product_id, quantity
	`, cartID, body.ProductID, body.Quantity).Scan(&item.ID, &item.ProductID, &item.Quantity)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

// PATCH /cart/items/{id} → actualizar quantity de un item
func updateCartItemHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.URL.Path[len("/cart/items/"):]
	itemID := 0
	fmt.Sscanf(idStr, "%d", &itemID)

	var body struct {
		Quantity int `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Quantity <= 0 {
		http.Error(w, "invalid quantity", http.StatusBadRequest)
		return
	}

	// Verificamos que el item pertenece al carrito del usuario
	var updated CartItem
	err = conn.QueryRow(context.Background(), `
		UPDATE cart_items ci
		SET quantity = $1
		FROM carts c
		WHERE ci.id = $2
		  AND ci.cart_id = c.id
		  AND c.user_id = $3
		RETURNING ci.id, ci.product_id, ci.quantity
	`, body.Quantity, itemID, userID).Scan(&updated.ID, &updated.ProductID, &updated.Quantity)

	if err != nil {
		http.Error(w, "item not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(updated)
}

// DELETE /cart/items/{id} → eliminar un item del carrito
func deleteCartItemHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.URL.Path[len("/cart/items/"):]
	itemID := 0
	fmt.Sscanf(idStr, "%d", &itemID)

	_, err = conn.Exec(context.Background(), `
		DELETE FROM cart_items ci
		USING carts c
		WHERE ci.id = $1
		  AND ci.cart_id = c.id
		  AND c.user_id = $2
	`, itemID, userID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DELETE /cart → vaciar carrito completo
func clearCartHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	_, err = conn.Exec(context.Background(), `
		DELETE FROM cart_items ci
		USING carts c
		WHERE ci.cart_id = c.id AND c.user_id = $1
	`, userID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Router del carrito — una sola función que despacha por método y path
func cartRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// /cart/items/{id}
	if strings.HasPrefix(path, "/cart/items/") && len(path) > len("/cart/items/") {
		switch r.Method {
		case http.MethodPatch:
			updateCartItemHandler(w, r)
		case http.MethodDelete:
			deleteCartItemHandler(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	// /cart/items
	if path == "/cart/items" {
		if r.Method == http.MethodPost {
			addCartItemHandler(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// /cart
	if path == "/cart" {
		switch r.Method {
		case http.MethodGet:
			getCartHandler(w, r)
		case http.MethodDelete:
			clearCartHandler(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	http.NotFound(w, r)
}

// ─── POSTS HANDLERS ────────────────────────────────────────

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

// ─── PRODUCTS HANDLERS ─────────────────────────────────────

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

// ─── INIT ──────────────────────────────────────────────────

func initDB() {
	godotenv.Load()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL not set")
	}
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		log.Fatal("JWT_SECRET not set")
	}
	var err error
	conn, err = pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	log.Println("Connected to PostgreSQL")
}

func main() {
	initDB()

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