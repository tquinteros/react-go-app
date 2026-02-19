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

// Extrae y valida el JWT del header Authorization
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

	// Hashear el password con bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "error hashing password", http.StatusInternalServerError)
		return
	}

	var user User
	err = conn.QueryRow(
		context.Background(),
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

	// Refresh token en cookie httpOnly
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
	err := conn.QueryRow(
		context.Background(),
		"SELECT id, email, password_hash FROM users WHERE email = $1",
		req.Email,
	).Scan(&user.ID, &user.Email, &passwordHash)

	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// Comparar password con el hash
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
		// Old refresh tokens may not have email; fetch from DB
		if err := conn.QueryRow(context.Background(), "SELECT email FROM users WHERE id = $1", userID).Scan(&email); err != nil {
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

// authCookie sets the refresh_token cookie. In dev (COOKIE_SECURE=false) use Secure=false
// and SameSite=Lax so the cookie is stored and sent on localhost (HTTP).
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

func generateRefreshToken(userID int, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
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
		err := conn.QueryRow(
			context.Background(),
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
		err := conn.QueryRow(
			context.Background(),
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

	// Auth — rutas públicas
	mux.HandleFunc("/auth/register", registerHandler)
	mux.HandleFunc("/auth/login", loginHandler)
	mux.HandleFunc("/auth/refresh", refreshHandler)
	mux.HandleFunc("/auth/logout", logoutHandler)

	// Posts — GET público, POST protegido
	mux.HandleFunc("/posts", postsHandler)
	mux.HandleFunc("/posts/", authMiddleware(postByIDHandler))

	// Products — GET público, resto protegido
	mux.HandleFunc("/products", productsHandler)
	mux.HandleFunc("/products/", authMiddleware(productByIDHandler))

	handler := corsMiddleware(mux)

	log.Println("server running at http://localhost:8080")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, handler))
}