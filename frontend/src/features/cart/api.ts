import type { Cart, CartItem } from "./types"
import { API_URL } from "@/lib/api"

export const cartQueryKey = ["cart"] as const

function cartHeaders(accessToken: string): HeadersInit {
  return {
    "Content-Type": "application/json",
    Authorization: `Bearer ${accessToken}`,
  }
}

export async function getCart(accessToken: string): Promise<Cart> {
  const res = await fetch(`${API_URL}/cart`, {
    headers: cartHeaders(accessToken),
  })
  if (!res.ok) throw new Error("Error fetching cart")
  return res.json()
}

export async function addCartItem(
  accessToken: string,
  productId: number,
  quantity: number = 1
): Promise<CartItem> {
  const res = await fetch(`${API_URL}/cart/items`, {
    method: "POST",
    headers: cartHeaders(accessToken),
    body: JSON.stringify({ product_id: productId, quantity }),
  })
  if (!res.ok) throw new Error("Error adding to cart")
  return res.json()
}

export async function updateCartItem(
  accessToken: string,
  itemId: number,
  quantity: number
): Promise<CartItem> {
  const res = await fetch(`${API_URL}/cart/items/${itemId}`, {
    method: "PATCH",
    headers: cartHeaders(accessToken),
    body: JSON.stringify({ quantity }),
  })
  if (!res.ok) throw new Error("Error updating cart item")
  return res.json()
}

export async function deleteCartItem(
  accessToken: string,
  itemId: number
): Promise<void> {
  const res = await fetch(`${API_URL}/cart/items/${itemId}`, {
    method: "DELETE",
    headers: cartHeaders(accessToken),
  })
  if (!res.ok) throw new Error("Error removing from cart")
}

export async function clearCart(accessToken: string): Promise<void> {
  const res = await fetch(`${API_URL}/cart`, {
    method: "DELETE",
    headers: cartHeaders(accessToken),
  })
  if (!res.ok) throw new Error("Error clearing cart")
}
