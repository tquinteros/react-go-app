package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

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
