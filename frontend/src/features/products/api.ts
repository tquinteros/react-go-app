import type { Product } from "./types"

// const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080"
const API_URL = "http://localhost:8080"

export async function fetchProducts(): Promise<Product[]> {
  const res = await fetch(`${API_URL}/products`)
  if (!res.ok) throw new Error("Error fetching products")
  return res.json()
}

export async function createProduct(product: Omit<Product, "id">): Promise<Product> {
  const res = await fetch(`${API_URL}/products`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(product),
  })
  if (!res.ok) throw new Error("Error creating product")
  return res.json()
}

export async function deleteProduct(id: number): Promise<void> {
  const res = await fetch(`${API_URL}/products/${id}`, {
    method: "DELETE",
  })
  if (!res.ok) throw new Error("Error deleting product")
}