"use client"

import { GameState, GameStateUpdate } from "@/types"
import { Dispatch, SetStateAction } from "react"

const parse = <T>(raw: string): T | null => {
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
  e: any,
) {
  const data = parse<GameState>(e.data)
  if (data) setGameState(data)
}

export function HandleMovePlayerEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
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

export function HandleGameStateUpdateEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {
  const data = parse<GameStateUpdate>(e.data)
  if (!data) return
  setGameState((prev) => {
    if (!prev) return prev
    return {
      ...prev,
      current_turn: data.current_turn,
      players: data.players,
    }
  })
}

export function HandleBankPaymentEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {}

export function HandleBankPayoutEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {}

export function HandleBankPaymentDueEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {}

export function HandleBankPayoutDueEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {}

export function HandleGameReadyEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {
  setGameState((prev) => {
    if (!prev) return prev
    return {
      ...prev,
      current_turn: 0,
    }
  })
}

export function HandleRollDiceEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {}

export function HandleRentDueEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {}

export function HandlePayToLeaveJailEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {}

export function HandleUseGetOutOfJailCardEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {}

export function HandleBankruptcyEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {}

export function HandleRentPaidEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {}

export function HandleHousePurchaseEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {}

export function HandleHotelPurchasedEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {}

export function HandleHouseSoldEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {}

export function HandleHotelSoldEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {}

export function HandlePropertyPurchasedEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {}

export function HandlePropertyMortgagedEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {}

export function HandlePropertyUnmortgagedEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {}
