"use client"

import { useMutation, useQuery } from "@tanstack/react-query"
import { Player } from "@/types"

const API_URL = process.env.NEXT_PUBLIC_API_URL

// Health check to verify backend is accessible
export async function checkBackendHealth(): Promise<boolean> {
  try {
    const res = await fetch(`${API_URL}/api/health`, {
      method: "GET",
      credentials: "include",
    })
    return res.ok
  } catch (err) {
    console.error("Backend health check failed:", err)
    return false
  }
}

/**
 * Create a new game session
 */
export function useCreateGame() {
  return useMutation({
    mutationFn: async (): Promise<{ session_id: string; code: number }> => {
      // Check if backend is accessible
      const isHealthy = await checkBackendHealth()
      if (!isHealthy) {
        throw new Error(`Backend not accessible at ${API_URL}. Make sure the backend server is running on port 9876.`)
      }

      const res = await fetch(`${API_URL}/api/game`, {
        method: "POST",
        credentials: "include",
      })
      if (!res.ok) {
        const errorText = await res.text()
        throw new Error(`Failed to create game: ${res.status} ${errorText}`)
      }
      return res.json()
    },
  })
}

export function useFetchAllGameSessions() {
  return useQuery<string[]>({
    queryKey: ["fetchAllGameSessions"],
    queryFn: async () => {
      const res = await fetch(`${API_URL}/api/game`, {
        method: "GET",
        credentials: "include",
      })
      if (!res.ok) {
        throw new Error(res.statusText)
      }
      return res.json()
    },
  })
}

/**
 * Join game with a code and get session ID
 */
export function useJoinGameByCode() {
  return useMutation({
    mutationFn: async (code: string): Promise<string> => {
      const res = await fetch(`${API_URL}/api/game/join?code=${code}`, {
        method: "POST",
        credentials: "include",
      })
      if (!res.ok) {
        throw new Error("Failed to join game")
      }
      return res.text()
    },
  })
}

export async function getAvailableTokens(players: Player[], sessionId: string): Promise<number[]> {
  try {
    const takenTokens = new Set(players.map((p) => p.piece_token))

    const availableTokens = [0, 1].filter((token) => !takenTokens.has(token))

    console.log("Available tokens:", availableTokens, "Taken tokens:", Array.from(takenTokens))
    return availableTokens
  } catch (err) {
    console.error("Error fetching available tokens:", err)
    return [0, 1]
  }
}

/**
 * Roll dice
 */
export function useRollDice() {
  return useMutation({
    mutationFn: async ({ playerId, sessionId }: { playerId: string; sessionId: string }) => {
      const formData = new FormData()
      formData.append("player_id", playerId)
      formData.append("session_id", sessionId)

      const res = await fetch(`${API_URL}/api/game/roll`, {
        method: "POST",
        credentials: "include",
        body: formData,
      })
      if (!res.ok) {
        const errMsg = await res.text()
        throw new Error(errMsg)
      }
      return res.json()
    },
  })
}

/**
 * Move player
 */
export function useMovePlayer() {
  return useMutation({
    mutationFn: async ({ playerId, sessionId }: { playerId: string; sessionId: string }) => {
      const formData = new FormData()
      formData.append("player_id", playerId.toString())
      formData.append("session_id", sessionId)

      const res = await fetch(`${API_URL}/api/game/move`, {
        method: "POST",
        credentials: "include",
        body: formData,
      })
      if (!res.ok) {
        const errMsg = await res.text()
        throw new Error(errMsg)
      }
      return res.json()
    },
  })
}

export function useJailRelease() {
  return useMutation({
    mutationFn: async (method: string) => {
      const res = await fetch(`${API_URL}/api/game/jail/release?method=${method}`, {
        method: "POST",
        credentials: "include",
      })
      if (!res.ok) {
        throw new Error(res.statusText)
      }
      return res.json()
    },
  })
}

export function usePayBank() {
  return useMutation({
    mutationFn: async () => {
      const res = await fetch(`${API_URL}/api/game/bank/pay`, {
        method: "POST",
        credentials: "include",
      })
      if (!res.ok) {
        throw new Error(res.statusText)
      }
      return res.json()
    },
  })
}

export function useSetBankPayout() {
  return useMutation({
    mutationFn: async ({ amount, reason }: { amount: string; reason: string }) => {
      const res = await fetch(`${API_URL}/api/game/bank/set?amount=${amount}&reason=${reason}`, {
        method: "POST",
        credentials: "include",
      })
      if (!res.ok) {
        throw new Error(res.statusText)
      }
      return res.json()
    },
  })
}

export function useReceiveBankPayout() {
  return useMutation({
    mutationFn: async () => {
      const res = await fetch(`${API_URL}/api/game/bank/receive`, {
        method: "POST",
        credentials: "include",
      })
      if (!res.ok) {
        throw new Error(res.statusText)
      }
      return res.json()
    },
  })
}

export function usePlayerExchange() {
  return useMutation({
    mutationFn: async () => {
      const res = await fetch(`${API_URL}/api/game/exchange`, {
        method: "POST",
        credentials: "include",
      })
      if (!res.ok) {
        throw new Error(res.statusText)
      }
      return res.json()
    },
  })
}

export function usePlayerBankrupt() {
  return useMutation({
    mutationFn: async () => {
      const res = await fetch(`${API_URL}/api/game/bankrupt`, {
        method: "POST",
        credentials: "include",
      })
      if (!res.ok) {
        throw new Error(res.statusText)
      }
      return res.json()
    },
  })
}

export function usePayRent() {
  return useMutation({
    mutationFn: async ({dst_player, amount} : {dst_player: string, amount: string}) => {
      const res = await fetch(`${API_URL}/api/game/rent?to_player_id=${dst_player}&amount=${amount}`, {
        method: "POST",
        credentials: "include",
      })
      if (!res.ok) {
        throw new Error(res.statusText)
      }
      return res.json()
    },
  })
}

