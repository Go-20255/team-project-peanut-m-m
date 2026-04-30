"use client"

import { useEffect, useState } from "react"
import { GameState } from "@/types"

const API_URL = process.env.NEXT_PUBLIC_API_URL

export function useLiveGameUpdates(
  sessionId: string | null,
  playerId: number | null,
  playerName: string | null,
) {
  const [gameState, setGameState] = useState<GameState | null>(null)

  useEffect(() => {
    if (!sessionId || !playerId || !playerName) return

    const eventSource = new EventSource(`${API_URL}/api/game/join/live`, {
      withCredentials: true,
    })

    const parse = <T,>(raw: string): T | null => {
      try {
        return JSON.parse(raw) as T
      } catch (e) {
        console.error("Failed to parse SSE payload:", e)
        return null
      }
    }

    eventSource.addEventListener("InitialGameBoardDataEvent", (e: any) => {
      const data = parse<GameState>(e.data)
      if (data) setGameState(data)
    })

    eventSource.addEventListener("MovePlayerEvent", (e: any) => {
      const data = parse<{
        player_id: number
        new_position: number
        turn_number?: number
      }>(e.data)
      if (!data) return

      setGameState((prev) => {
        if (!prev) return prev
        return {
          ...prev,
          current_turn: data.turn_number ?? prev.current_turn,
          players: prev.players.map((pi) =>
            pi.player.id === data.player_id
              ? {
                  ...pi,
                  player: { ...pi.player, position: data.new_position },
                }
              : pi,
          ),
        }
      })
    })

    eventSource.onerror = (err) => {
      console.error("SSE connection error:", err)
      eventSource.close()
    }

    return () => eventSource.close()
  }, [sessionId, playerId, playerName])

  return gameState
}
