"use client"

import { useEffect, useRef } from "react"
import { useMutation } from "@tanstack/react-query"
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

function handleSSEEvent(eventData: string, eventType: string, callback: (type: string, data: any) => void) {
  try {
    const data = JSON.parse(eventData)
    console.log(`Received ${eventType}:`, data)
    callback(eventType, data)
  } catch (e) {
    console.error(`Failed to parse ${eventType}:`, e)
  }
}

/**
 * Connect to live game updates via SSE and set up event listeners
 */
export function useLiveGameUpdates(
  sessionId: string | null,
  playerId: number | null,
  playerName: string | null,
  onUpdate: (data: any) => void,
  enabled = true,
) {
  const onUpdateRef = useRef(onUpdate)

  // Update the ref when onUpdate changes, but don't re-establish the connection
  useEffect(() => {
    onUpdateRef.current = onUpdate
  }, [onUpdate])

  useEffect(() => {
    if (!enabled || !sessionId || !playerId || !playerName) {
      return
    }

    const eventSource = new EventSource(`${API_URL}/api/game/join/live`, {
      withCredentials: true
    })

    console.log("SSE connection established for session:", sessionId, "player:", playerId)

    // Special case: InitialGameBoardDataEvent sends multiple updates
    eventSource.addEventListener("InitialGameBoardDataEvent", (event: any) => {
      handleSSEEvent(event.data, "InitialGameBoardDataEvent", (_, data) => {
        onUpdateRef.current({ type: "playersUpdate", data: data.Players })
        onUpdateRef.current({ type: "gameStateUpdate", data: { turn_number: data.CurrentTurn } })
      })
    })

    // Map event names to their corresponding update types
    const eventMap: { [key: string]: string } = {
      RollDiceEvent: "diceRoll",
      MovePlayerEvent: "playerMove",
      RentPaidEvent: "rentPaid",
      PropertyPurchased: "propertyPurchased",
      HousePurchased: "housePurchased",
      HotelPurchased: "hotelPurchased",
      HouseSold: "houseSold",
      HotelSold: "hotelSold",
      PropertyMortgaged: "propertyMortgaged",
      PropertyUnmortgaged: "propertyUnmortgaged",
    }

    Object.entries(eventMap).forEach(([eventName, updateType]) => {
      eventSource.addEventListener(eventName, (event: any) => {
        handleSSEEvent(event.data, eventName, (_, data) => {
          onUpdateRef.current({ type: updateType, data })
        })
      })
    })

    eventSource.onerror = (error) => {
      console.error("SSE connection error:", error)
      eventSource.close()
    }

    return () => {
      console.log("Closing SSE connection for session:", sessionId)
      eventSource.close()
    }
  }, [sessionId, playerId, playerName, enabled])
}

/**
 * Roll dice
 */
export function useRollDice() {
  return useMutation({
    mutationFn: async ({ playerId, sessionId }: { playerId: number; sessionId: string }) => {
      const formData = new FormData()
      formData.append("player_id", playerId.toString())
      formData.append("session_id", sessionId)

      const res = await fetch(`${API_URL}/api/game/roll`, {
        method: "POST",
        credentials: "include",
        body: formData,
      })
      if (!res.ok) {
        throw new Error("Failed to roll dice")
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
    mutationFn: async ({ playerId, sessionId }: { playerId: number; sessionId: string }) => {
      const formData = new FormData()
      formData.append("player_id", playerId.toString())
      formData.append("session_id", sessionId)

      const res = await fetch(`${API_URL}/api/game/move`, {
        method: "POST",
        credentials: "include",
        body: formData,
      })
      if (!res.ok) {
        throw new Error("Failed to move player")
      }
      return res.json()
    },
  })
}

/**
 * Purchase property
 */
export function usePurchaseProperty() {
  return useMutation({
    mutationFn: async ({
      playerId,
      sessionId,
      propertyId,
    }: {
      playerId: number
      sessionId: string
      propertyId: number
    }) => {
      const formData = new FormData()
      formData.append("player_id", playerId.toString())
      formData.append("session_id", sessionId)
      formData.append("property_id", propertyId.toString())

      const res = await fetch(`${API_URL}/api/game/property`, {
        method: "POST",
        credentials: "include",
        body: formData,
      })
      if (!res.ok) {
        throw new Error("Failed to purchase property")
      }
      return res.json()
    },
  })
}
