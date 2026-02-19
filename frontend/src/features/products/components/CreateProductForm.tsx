import { useMutation, useQueryClient } from "@tanstack/react-query"
import { createProduct } from "../api"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { useState } from "react"

const HARDCODED_IMAGE =
  "https://cdn.shopify.com/s/files/1/2987/0758/files/100101222-100201-blackwhite-02.jpg?v=1749120579&width=400&height=515&crop=center"

type CreateProductFormProps = {
  onSuccess?: () => void
}

const CreateProductForm = ({ onSuccess }: CreateProductFormProps) => {
  const queryClient = useQueryClient()
  const [name, setName] = useState("")
  const [description, setDescription] = useState("")
  const [price, setPrice] = useState("")
  const [discount, setDiscount] = useState("")

  const createMutation = useMutation({
    mutationFn: () =>
      createProduct({
        name,
        description,
        price: Number(price) || 0,
        discount: Number(discount) || 0,
        images: [HARDCODED_IMAGE],
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["products"] })
      setName("")
      setDescription("")
      setPrice("")
      setDiscount("")
      onSuccess?.()
    },
  })

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    createMutation.mutate()
  }

  return (
    <Card className="w-full max-w-md border-border/60 shadow-sm">
      <CardHeader className="space-y-1">
        <CardTitle className="text-xl">New product</CardTitle>
        <CardDescription>
          Add a new product. Image is set automatically for now.
        </CardDescription>
      </CardHeader>
      <form onSubmit={handleSubmit}>
        <CardContent className="flex flex-col gap-4">
          <div className="grid gap-2">
            <Label htmlFor="product-name">Name</Label>
            <Input
              id="product-name"
              placeholder="e.g. Classic Tee"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              className="transition-[border-color,box-shadow]"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="product-description">Description</Label>
            <Textarea
              id="product-description"
              placeholder="Short description of the product..."
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              required
              rows={3}
              className="min-h-20 resize-y transition-[border-color,box-shadow]"
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="grid gap-2">
              <Label htmlFor="product-price">Price (USD)</Label>
              <Input
                id="product-price"
                type="number"
                min={0}
                step={0.01}
                placeholder="0"
                value={price}
                onChange={(e) => setPrice(e.target.value)}
                required
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="product-discount">Discount (%)</Label>
              <Input
                id="product-discount"
                type="number"
                min={0}
                max={100}
                step={1}
                placeholder="0"
                value={discount}
                onChange={(e) => setDiscount(e.target.value)}
              />
            </div>
          </div>
        </CardContent>
        <CardFooter className="flex flex-col gap-3">
          {createMutation.isError && (
            <p className="text-destructive text-sm w-full" role="alert">
              {createMutation.error.message}
            </p>
          )}
          <Button
            type="submit"
            className="w-full"
            disabled={createMutation.isPending}
          >
            {createMutation.isPending ? "Creating…" : "Create product"}
          </Button>
        </CardFooter>
      </form>
    </Card>
  )
}

export default CreateProductForm
