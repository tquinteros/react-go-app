import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createPost, deletePost, fetchPosts } from '../features/posts/api'
import { useAuth } from '@/hooks/use-auth'
import { useState } from 'react'

const PostList = () => {
    const [title, setTitle] = useState('')
    const [body, setBody] = useState('')
    const queryClient = useQueryClient()
    const { accessToken } = useAuth()

    const createPostMutation = useMutation({
        mutationFn: ({ title, body }: { title: string; body: string }) => createPost(title, body),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['posts'] })
            setTitle('')
            setBody('')
        },
    })

    const deletePostMutation = useMutation({
        mutationFn: (id: number) => deletePost(id, accessToken),
        onSuccess: () => {
            console.log('Post deleted successfully')
            queryClient.invalidateQueries({ queryKey: ['posts'] })
        },
        onError: (error) => {
            console.error('Error deleting post:', error)
        },
    })

    const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault()
        createPostMutation.mutate({ title, body })
    }

    const { data: posts = [], isLoading } = useQuery({
        queryKey: ['posts'],
        queryFn: fetchPosts,
    })

    if (isLoading) {
        return (
            <div className="flex justify-center items-center py-16" aria-busy="true">
                <div
                    className="h-10 w-10 border-2 border-gray-600 border-t-white rounded-full animate-spin"
                    role="status"
                    aria-label="Loading posts"
                />
            </div>
        )
    }

    return (
        <div className="">

            <h2>Create Post</h2>

            <form onSubmit={handleSubmit}>
                <input
                    value={title}
                    onChange={e => setTitle(e.target.value)}
                    placeholder="Title"
                />

                <textarea
                    value={body}
                    onChange={e => setBody(e.target.value)}
                    placeholder="Body"
                />

                <button type="submit" disabled={createPostMutation.isPending}>
                    {createPostMutation.isPending ? 'Creating…' : 'Create'}
                </button>
                {createPostMutation.isError && (
                    <p className="text-red-500 text-sm mt-1">{createPostMutation.error.message}</p>
                )}
            </form>

            <hr />

            {posts.map((post) => (
                <div key={post.id} className="border-b border-gray-200 p-4">
                    <h2 className="text-2xl font-bold">{post.title}</h2>
                    <p>{post.body}</p>
                    <button
                        onClick={() => deletePostMutation.mutate(post.id)}
                        disabled={deletePostMutation.isPending}
                        className="bg-red-500 text-white px-4 py-2 rounded-md cursor-pointer"
                    >
                        {deletePostMutation.isPending ? 'Eliminando…' : 'Eliminar'}
                    </button>
                </div>
            ))}
        </div>
    )
}

export default PostList
