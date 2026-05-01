"use client"

import { useEffect, useState } from "react"
import { GameState } from "@/types"
import { HandleInitialGameBoardUpdateEvent, HandleMovePlayerEvent } from "./gameEvents"

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

    // map event name with its event handler here
    const eventManager = {
      "InitialGameBoardDataEvent" : HandleInitialGameBoardUpdateEvent,
      "MovePlayerEvent": HandleMovePlayerEvent,
    }

    Object.entries(eventManager).map(([eventName, eventHandler]) => {
      eventSource.addEventListener(eventName, (e: any) => eventHandler(gameState, setGameState, e))
    })

    eventSource.onerror = (err) => {
      console.error("SSE connection error:", err)
      eventSource.close()
    }

    return () => eventSource.close()
  }, [sessionId, playerId, playerName])

  return gameState
}
