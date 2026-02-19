import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetFooter,
} from "@/components/ui/sheet"
import { Button } from "@/components/ui/button"
import { useCartStore } from "./store"
import { Minus, Plus, ShoppingCart, Trash2 } from "lucide-react"

const placeholderImage =
  "https://placehold.co/80/f4f4f5/71717a?text=No+image"

function formatPrice(price: number) {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(price)
}

const CartDrawer = () => {
  const {
    items,
    isCartOpen,
    closeCart,
    updateQuantity,
    removeItem,
    clearCart,
  } = useCartStore()

  const total = items.reduce(
    (sum, item) => {
      const hasDiscount = item.product.discount > 0
      const price = hasDiscount
        ? item.product.price * (1 - item.product.discount / 100)
        : item.product.price
      return sum + price * item.quantity
    },
    0
  )

  return (
    <Sheet open={isCartOpen} onOpenChange={(open) => !open && closeCart()}>
      <SheetContent side="right" className="flex w-full flex-col sm:max-w-md">
        <SheetHeader>
          <SheetTitle className="flex items-center gap-2">
            <ShoppingCart className="size-5" />
            Cart ({items.reduce((n, i) => n + i.quantity, 0)} items)
          </SheetTitle>
        </SheetHeader>
        <div className="flex-1 overflow-y-auto py-4">
          {items.length === 0 ? (
            <p className="text-muted-foreground text-center text-sm">
              Your cart is empty.
            </p>
          ) : (
            <ul className="flex flex-col gap-4">
              {items.map(({ product, quantity }) => {
                const hasDiscount = product.discount > 0
                const unitPrice = hasDiscount
                  ? product.price * (1 - product.discount / 100)
                  : product.price
                const image = product.images?.[0] ?? placeholderImage
                return (
                  <li
                    key={product.id}
                    className="flex gap-3 rounded-lg border border-border/60 p-3"
                  >
                    <img
                      src={image}
                      alt={product.name}
                      className="size-16 shrink-0 rounded-md object-cover"
                    />
                    <div className="min-w-0 flex-1">
                      <p className="truncate font-medium text-foreground">
                        {product.name}
                      </p>
                      <p className="text-muted-foreground text-sm">
                        {formatPrice(unitPrice)} each
                      </p>
                      <div className="mt-2 flex items-center gap-2">
                        <div className="flex items-center rounded-md border border-input">
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon-xs"
                            onClick={() =>
                              updateQuantity(product.id, quantity - 1)
                            }
                            disabled={quantity <= 1}
                            aria-label="Decrease quantity"
                          >
                            <Minus className="size-3" />
                          </Button>
                          <span className="min-w-6 text-center text-sm tabular-nums">
                            {quantity}
                          </span>
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon-xs"
                            onClick={() =>
                              updateQuantity(product.id, quantity + 1)
                            }
                            aria-label="Increase quantity"
                          >
                            <Plus className="size-3" />
                          </Button>
                        </div>
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon-xs"
                          onClick={() => removeItem(product.id)}
                          className="text-destructive hover:bg-destructive/10 hover:text-destructive"
                          aria-label="Remove from cart"
                        >
                          <Trash2 className="size-4" />
                        </Button>
                      </div>
                    </div>
                    <p className="shrink-0 font-medium tabular-nums">
                      {formatPrice(unitPrice * quantity)}
                    </p>
                  </li>
                )
              })}
            </ul>
          )}
        </div>
        {items.length > 0 && (
          <SheetFooter className="flex-col gap-2 border-t pt-4 sm:flex-col">
            <div className="flex w-full items-center justify-between text-base font-semibold">
              <span>Total</span>
              <span className="tabular-nums">{formatPrice(total)}</span>
            </div>
            <div className="flex w-full gap-2">
              <Button
                variant="outline"
                className="flex-1"
                onClick={clearCart}
              >
                Clear cart
              </Button>
              <Button className="flex-1">Checkout</Button>
            </div>
          </SheetFooter>
        )}
      </SheetContent>
    </Sheet>
  )
}

export default CartDrawer
