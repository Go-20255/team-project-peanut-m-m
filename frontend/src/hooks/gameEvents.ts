"use client"

import { GameState } from "@/types";
import { Dispatch, SetStateAction } from "react";

const parse = <T,>(raw: string): T | null => {
  try {
    return JSON.parse(raw) as T
  } catch (e) {
    console.error("Failed to parse SSE payload:", e)
    return null
  }
}

export function HandleInitialGameBoardUpdateEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any
) {
  const data = parse<GameState>(e.data)
  if (data) setGameState(data)
}

export function HandleMovePlayerEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any
) {
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
}
