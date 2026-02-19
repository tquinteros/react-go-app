import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { deleteProduct, fetchProducts } from '../api'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogTrigger } from '@/components/ui/dialog'
import ProductCard from './ProductCard'
import CreateProductForm from './CreateProductForm'

const ProductsList = () => {
    const [createOpen, setCreateOpen] = useState(false)
    const queryClient = useQueryClient()

    const deleteProductMutation = useMutation({
        mutationFn: (id: number) => deleteProduct(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['products'] })
        },
    })

    const { data: products = [], isLoading } = useQuery({
        queryKey: ['products'],
        queryFn: fetchProducts,
    })

    if (isLoading) {
        return <div>Loading...</div>
    }

    return (
        <div className="space-y-6">
            <Dialog open={createOpen} onOpenChange={setCreateOpen}>
                <DialogTrigger asChild>
                    <Button>Create Product</Button>
                </DialogTrigger>
                <DialogContent className="sm:max-w-md">
                    <CreateProductForm onSuccess={() => setCreateOpen(false)} />
                </DialogContent>
            </Dialog>
            <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
                {products.map((product) => (
                    <ProductCard
                        key={product.id}
                        product={product}
                        onDelete={(id) => deleteProductMutation.mutate(id)}
                        isDeleting={deleteProductMutation.isPending}
                    />
                ))}
            </div>
        </div>
    )
}

export default ProductsList