import type { Post } from "./types"

export async function fetchPosts(): Promise<Post[]> {
  const res = await fetch("https://jsonplaceholder.typicode.com/posts")
  if (!res.ok) throw new Error("Error fetching posts")
  return res.json()
}


/** Create post – Postman: POST http://localhost:8080/posts, Body raw JSON: { "title": "...", "body": "..." } */
export async function createPost(title: string, body: string) {
  const res = await fetch("http://localhost:8080/posts", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ title, body }),
  })

  if (!res.ok) {
    throw new Error("Error creating post")
  }

  return res.json()
}

// export async function fetchPostsGo(): Promise<Post[]> {
//     const res = await fetch("http://localhost:8080/posts")
//     if (!res.ok) throw new Error("Error fetching posts")
//     return res.json()
//   }