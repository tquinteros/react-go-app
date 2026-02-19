import type { Post } from "./types"


const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080"
// export async function fetchPosts(): Promise<Post[]> {
//   const res = await fetch("https://jsonplaceholder.typicode.com/posts")
//   if (!res.ok) throw new Error("Error fetching posts")
//   return res.json()
// }


export async function fetchPosts(): Promise<Post[]> {
  const res = await fetch(`${API_URL}/posts`)
  if (!res.ok) throw new Error("Error fetching posts")
  return res.json()
}


export async function createPost(title: string, body: string): Promise<Post> {
  const res = await fetch(`${API_URL}/posts`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ title, body }),
  })
  if (!res.ok) throw new Error("Error creating post")
  return res.json()
}

/** Create post – Postman: POST http://localhost:8080/posts, Body raw JSON: { "title": "...", "body": "..." } */
// export async function createPost(title: string, body: string) {
//   const res = await fetch("http://localhost:8080/posts", {
//     method: "POST",
//     headers: {
//       "Content-Type": "application/json",
//     },
//     body: JSON.stringify({ title, body }),
//   })

//   if (!res.ok) {
//     throw new Error("Error creating post")
//   }

//   return res.json()
// }

// export async function fetchPostsGo(): Promise<Post[]> {
//     const res = await fetch("http://localhost:8080/posts")
//     if (!res.ok) throw new Error("Error fetching posts")
//     return res.json()
//   }