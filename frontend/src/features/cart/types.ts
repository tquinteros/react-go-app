export interface CartItem {
  id: number
  product_id: number
  quantity: number
  name: string
  price: number
  images: string[]
  discount: number
}

export interface Cart {
  id: number
  user_id: number
  items: CartItem[]
}
