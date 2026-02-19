import { useQuery } from '@tanstack/react-query'

interface Post {
  id: number
  title: string
  body: string
}

const fetchPosts = async (): Promise<Post[]> => {
  const response = await fetch('https://jsonplaceholder.typicode.com/posts')
  if (!response.ok) throw new Error('Failed to fetch posts')
  return response.json()
}

const PostList = () => {
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
      {posts.map((post) => (
        <div key={post.id} className="border-b border-gray-200 p-4">
          <h2 className="text-2xl font-bold">{post.title}</h2>
          <p>{post.body}</p>
        </div>
      ))}
    </div>
  )
}

export default PostList
