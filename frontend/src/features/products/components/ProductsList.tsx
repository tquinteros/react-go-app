import React from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { deleteProduct, fetchProducts } from '../api'
import { Button } from '@/components/ui/button'

const ProductsList = () => {

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
        <div>
            <Button>Create Product</Button>
            {products.map((product) => (
                <div key={product.id} className="border-b border-gray-200 p-4">
                    <h2 className="text-2xl font-bold">{product.name}</h2>
                    <p>{product.description}</p>
                    <p>{product.price}</p>
                    <p>{product.images}</p>
                    <p>{product.discount}</p>
                    <Button onClick={() => deleteProductMutation.mutate(product.id)} disabled={deleteProductMutation.isPending}>{deleteProductMutation.isPending ? 'Deleting...' : 'Delete Product'}</Button>
                </div>
            ))}
        </div>
    )
}

export default ProductsList