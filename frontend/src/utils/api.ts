const rawApiUrl = process.env.NEXT_PUBLIC_API_URL ?? ""

export const API_URL = rawApiUrl.trim().replace(/^["']|["']$/g, "").replace(/\/$/, "") || "http://localhost:9876"
