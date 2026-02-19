import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  Carousel,
  CarouselContent,
  CarouselItem,
  CarouselNext,
  CarouselPrevious,
} from "@/components/ui/carousel"
import { Button } from "@/components/ui/button"
import type { Product } from "../types"
import { ShoppingCart, Tag } from "lucide-react"

interface ProductCardProps {
  product: Product
  onDelete?: (id: number) => void
  isDeleting?: boolean
}

const placeholderImage = "https://placehold.co/400x300/f4f4f5/71717a?text=No+image"

function formatPrice(price: number) {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(price)
}

const ProductCard = ({ product, onDelete, isDeleting }: ProductCardProps) => {
  const images = product.images?.length ? product.images : [placeholderImage]
  const hasDiscount = product.discount > 0

  return (
    <Card className="group overflow-hidden border-border/60 bg-card shadow-sm transition-all duration-300 hover:shadow-md hover:border-border">
      <CardHeader className="relative gap-0 overflow-hidden p-0 pb-2">
        <div className="relative aspect-4/3 w-full bg-muted">
          <Carousel
            opts={{ align: "start", loop: true }}
            className="w-full"
          >
            <CarouselContent className="ml-0">
              {images.map((src, i) => (
                <CarouselItem key={i} className="pl-0">
                  <img
                    src={src}
                    alt={`${product.name} ${i + 1}`}
                    className="h-full w-full object-cover transition-transform duration-300 group-hover:scale-[1.02]"
                  />
                </CarouselItem>
              ))}
            </CarouselContent>
            {images.length > 1 && (
              <>
                <CarouselPrevious className="left-2 size-8 border-0 bg-black/40 text-white opacity-0 transition-opacity group-hover:opacity-100" />
                <CarouselNext className="right-2 size-8 border-0 bg-black/40 text-white opacity-0 transition-opacity group-hover:opacity-100" />
              </>
            )}
          </Carousel>
          {hasDiscount && (
            <span className="absolute right-2 top-2 inline-flex items-center gap-1 rounded-full bg-destructive px-2.5 py-1 text-xs font-semibold text-destructive-foreground shadow-sm">
              <Tag className="size-3" />
              -{product.discount}%
            </span>
          )}
        </div>
        <div className="space-y-1 px-4 pt-3">
          <CardTitle className="line-clamp-1 text-lg font-semibold tracking-tight">
            {product.name}
          </CardTitle>
          <CardDescription className="line-clamp-2 text-sm">
            {product.description}
          </CardDescription>
        </div>
      </CardHeader>
      <CardContent className="flex flex-col gap-3 px-4 pb-2">
        <div className="flex items-baseline gap-2">
          <span className="text-xl font-bold tracking-tight text-foreground">
            {formatPrice(hasDiscount ? product.price * (1 - product.discount / 100) : product.price)}
          </span>
          {hasDiscount && (
            <span className="text-sm text-muted-foreground line-through">
              {formatPrice(product.price)}
            </span>
          )}
        </div>
      </CardContent>
      <CardFooter className="flex gap-2 px-4 pb-4 pt-0">
        <Button className="flex-1" size="sm">
          <ShoppingCart className="size-4" />
          Add to cart
        </Button>
        {onDelete && (
          <Button
            variant="outline"
            size="sm"
            onClick={() => onDelete(product.id)}
            disabled={isDeleting}
            className="text-destructive hover:bg-destructive/10 hover:text-destructive"
          >
            {isDeleting ? "…" : "Delete"}
          </Button>
        )}
      </CardFooter>
    </Card>
  )
}

export default ProductCard
