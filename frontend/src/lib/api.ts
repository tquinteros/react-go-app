/**
 * API base URL: uses local URL in development (VITE_API_URL_LOCAL),
 * production URL otherwise (VITE_API_URL). No need to comment/uncomment.
 */
export const API_URL =
  import.meta.env.DEV
    ? (import.meta.env.VITE_API_URL_LOCAL ?? "http://localhost:8080")
    : (import.meta.env.VITE_API_URL ?? "http://localhost:8080")
