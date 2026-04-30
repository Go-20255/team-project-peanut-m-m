"use client"

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Player } from "@/types"
import { storage } from "@/utils"
const API_URL = process.env.NEXT_PUBLIC_API_URL

/**
 * Create a player in a game session
 */
export function useCreatePlayer() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({
      playerName,
      sessionId,
      pieceToken,
    }: {
      playerName: string
      sessionId: string
      pieceToken: number
    }): Promise<Player> => {
      const formData = new FormData()
      formData.append("player_name", playerName)
      formData.append("session_id", sessionId)
      formData.append("piece_token", pieceToken.toString())

      const res = await fetch(`${API_URL}/api/player`, {
        method: "POST",
        credentials: "include",
        body: formData,
      })

      if (!res.ok) {
        const errorText = await res.text()
        throw new Error(`Failed to create player: ${res.status} ${errorText}`)
      }

      return res.json()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["fetchPlayersForSession"] })
    },
  })
}


export function useUpdatePlayerToken() {
  return useMutation({
    mutationFn: async ({
      playerId,
      sessionId,
      pieceToken,
    }: {
      playerId: number
      sessionId: string
      pieceToken: number
    }) => {
      const formData = new FormData()
      formData.append("player_id", playerId.toString())
      formData.append("session_id", sessionId)
      formData.append("piece_token", pieceToken.toString())

      const res = await fetch(`${API_URL}/api/player`, {
        method: "PATCH",
        credentials: "include",
        body: formData,
      })
      if (!res.ok) {
        const errorText = await res.text()
        throw new Error(`Failed to update player token: ${res.status} ${errorText}`)
      }
      return res.json()
    },
  })
}


export function useFetchPlayersForSession() {
  return useQuery<Player[]>({
    queryKey: ["fetchPlayersForSession"],
    retry: 3,
    queryFn: async (): Promise<Player[]> => {
      const sessionId = storage.getSessionId()
      const res = await fetch(`${API_URL}/api/game/players?session_id=${sessionId}`, {
        method: "GET",
        credentials: "include",
      })
      console.log("Players fetch response status:", res.status)
      if (!res.ok) {
        const errorText = await res.text()
        console.error("Players fetch error:", res.status, errorText)
        throw new Error(`Failed to fetch players: ${res.status} ${errorText}`)
      }

      return res.json()
    },
  })
}

export async function fetchPlayersForSession(sessionId: string): Promise<any[]> {
  try {
    console.log("Fetching players for session:", sessionId)
    const res = await fetch(`${API_URL}/api/game/players?session_id=${sessionId}`, {
      method: "GET",
      credentials: "include",
    })
    console.log("Players fetch response status:", res.status)

    if (!res.ok) {
      const errorText = await res.text()
      console.error("Players fetch error:", res.status, errorText)
      throw new Error(`Failed to fetch players: ${res.status} ${errorText}`)
    }

    const data = await res.json()
    console.log("Players fetched successfully:", data)
    return data || []
  } catch (err) {
    console.error("Error fetching players:", err)
    return []
  }
}

export function useLoginPlayer() {
  return useMutation({
    mutationFn: async (p: Player) => {
      const response = await fetch(`${API_URL}/api/player/join?player_id=${p.id}&player_name=${p.name}&session_id=${p.session_id}`, {
        method:"POST",
        credentials: "include"
      })
      if (!response.ok) {
        throw new Error(response.statusText)
      }
      return response
    }
  })

}

