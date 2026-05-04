"use client"

import { GameState, GameStateUpdate, PendingBankPayment, PropertyPurchaseAvailable } from "@/types"
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
  if (data) {
    setGameState({
      ...data,
      current_roll: null,
      last_move: null,
      pending_property_purchase: null,
      pending_bank_payment: null,
    })
  }
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
    old_position: number
    total: number
    passed_go: boolean
    rent_due: boolean
    rent_amount: number
    rent_to_id: number
    property_id: number
    roll_again: boolean
  }>(e.data)
  if (!data) return

  setGameState((prev) => {
    if (!prev) return prev
    return {
      ...prev,
      current_turn: data.turn_number ?? prev.current_turn,
      current_roll: null,
      pending_property_purchase: null,
      pending_bank_payment: null,
      last_move: {
        player_id: data.player_id,
        session_id: prev.players.find((pi) => pi.player.id === data.player_id)?.player.session_id ?? "",
        old_position: data.old_position,
        new_position: data.new_position,
        total: data.total,
        passed_go: data.passed_go,
        turn_number: data.turn_number ?? prev.current_turn,
        rent_due: data.rent_due,
        rent_amount: data.rent_amount,
        rent_to_id: data.rent_to_id,
        property_id: data.property_id,
        roll_again: data.roll_again,
      },
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
    const turnChanged = data.current_turn !== prev.current_turn
    return {
      ...prev,
      current_turn: data.current_turn,
      current_roll: turnChanged ? null : prev.current_roll,
      last_move: turnChanged ? null : prev.last_move,
      pending_property_purchase: turnChanged ? null : prev.pending_property_purchase,
      pending_bank_payment: turnChanged ? null : prev.pending_bank_payment,
      players: data.players,
    }
  })
}

export function HandleBankPaymentEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {
  setGameState((prev) => {
    if (!prev) return prev
    return {
      ...prev,
      pending_bank_payment: null,
    }
  })
}

export function HandleBankPayoutEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {}

export function HandleBankPaymentDueEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {
  const data = parse<PendingBankPayment>(e.data)
  if (!data) return

  setGameState((prev) => {
    if (!prev) return prev
    return {
      ...prev,
      pending_bank_payment: data,
    }
  })
}

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
      current_roll: null,
      last_move: null,
      pending_property_purchase: null,
      pending_bank_payment: null,
    }
  })
}

export function HandleRollDiceEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {
  const data = parse<{
    player_id: number
    session_id: string
    die_one: number
    die_two: number
    total: number
    is_double: boolean
    roll_again: boolean
    released_from_jail: boolean
    sent_to_jail: boolean
    jailed: number
  }>(e.data)
  if (!data) return

  setGameState((prev) => {
    if (!prev) return prev
    return {
      ...prev,
      current_roll: data,
    }
  })
}

export function HandleRentDueEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {}

export function HandlePropertyPurchaseAvailableEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {
  const data = parse<PropertyPurchaseAvailable>(e.data)
  if (!data) return

  setGameState((prev) => {
    if (!prev) return prev
    return {
      ...prev,
      pending_property_purchase: data,
    }
  })
}

export function HandleDrawCardEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {}

export function HandlePayToLeaveJailEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {
  setGameState((prev) => {
    if (!prev) return prev
    return {
      ...prev,
      pending_bank_payment: null,
      current_roll: null,
    }
  })
}

export function HandleUseGetOutOfJailCardEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {
  setGameState((prev) => {
    if (!prev) return prev
    return {
      ...prev,
      pending_bank_payment: null,
      current_roll: null,
    }
  })
}

export function HandleBankruptcyEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {}

export function HandlePlayerExchangeDueEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {}

export function HandlePlayerExchangeEvent(
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
) {
  setGameState((prev) => {
    if (!prev) return prev
    return {
      ...prev,
      pending_property_purchase: null,
    }
  })
}

export function HandlePropertyPurchaseIgnoredEvent(
  gameState: GameState | null,
  setGameState: Dispatch<SetStateAction<GameState | null>>,
  e: any,
) {
  setGameState((prev) => {
    if (!prev) return prev
    return {
      ...prev,
      pending_property_purchase: null,
    }
  })
}

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
