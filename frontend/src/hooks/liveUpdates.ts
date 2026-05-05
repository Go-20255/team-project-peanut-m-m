"use client"

import { useEffect, useState } from "react"
import { GameState } from "@/types"
import {
  HandleBankPaymentDueEvent,
  HandleBankPaymentEvent,
  HandleBankPayoutDueEvent,
  HandleBankPayoutEvent,
  HandleBankruptcyEvent,
  HandleCardDrawAvailableEvent,
  HandleCardResolvedEvent,
  HandleDrawCardEvent,
  HandleGameReadyEvent,
  HandleGameStateUpdateEvent,
  HandleHotelPurchasedEvent,
  HandleHotelSoldEvent,
  HandleHousePurchaseEvent as HandleHousePurchasedEvent,
  HandleHouseSoldEvent,
  HandleInitialGameBoardUpdateEvent,
  HandleMovePlayerEvent,
  HandlePayToLeaveJailEvent,
  HandlePropertyPurchaseAvailableEvent,
  HandlePropertyPurchaseIgnoredEvent,
  HandlePropertyMortgagedEvent,
  HandlePropertyPurchasedEvent,
  HandlePropertyUnmortgagedEvent,
  HandlePlayerExchangeDueEvent,
  HandlePlayerExchangeEvent,
  HandleRentDueEvent,
  HandleRentPaidEvent,
  HandleRollDiceEvent,
  HandleUseGetOutOfJailCardEvent,
} from "./gameEvents"
import { API_URL } from "@/utils/api"

export function useLiveGameUpdates(sessionId: string | null, playerId: number | null, playerName: string | null) {
  const [gameState, setGameState] = useState<GameState | null>(null)

  useEffect(() => {
    if (!sessionId || !playerId || !playerName) return

    const eventSource = new EventSource(`${API_URL}/api/game/join/live`, {
      withCredentials: true,
    })

    // map event name with its event handler here
    const eventManager = {
      InitialGameBoardDataEvent: HandleInitialGameBoardUpdateEvent,
      GameStateUpdateEvent: HandleGameStateUpdateEvent,
      BankPaymentDueEvent: HandleBankPaymentDueEvent,
      BankPayoutDueEvent: HandleBankPayoutDueEvent,
      BankPaymentEvent: HandleBankPaymentEvent,
      BankPayoutEvent: HandleBankPayoutEvent,
      GameReadyEvent: HandleGameReadyEvent,
      RollDiceEvent: HandleRollDiceEvent,
      MovePlayerEvent: HandleMovePlayerEvent,
      CardDrawAvailableEvent: HandleCardDrawAvailableEvent,
      CardResolvedEvent: HandleCardResolvedEvent,
      PropertyPurchaseAvailableEvent: HandlePropertyPurchaseAvailableEvent,
      PropertyPurchaseIgnoredEvent: HandlePropertyPurchaseIgnoredEvent,
      DrawCardEvent: HandleDrawCardEvent,
      RentDueEvent: HandleRentDueEvent,
      PayToLeaveJailEvent: HandlePayToLeaveJailEvent,
      UseGetOutOfJailCardEvent: HandleUseGetOutOfJailCardEvent,
      BankruptcyEvent: HandleBankruptcyEvent,
      RentPaidEvent: HandleRentPaidEvent,
      HousePurchasedEvent: HandleHousePurchasedEvent,
      HotelPurchasedEvent: HandleHotelPurchasedEvent,
      HouseSoldEvent: HandleHouseSoldEvent,
      HotelSoldEvent: HandleHotelSoldEvent,
      PropertyPurchasedEvent: HandlePropertyPurchasedEvent,
      PropertyMortgagedEvent: HandlePropertyMortgagedEvent,
      PropertyUnmortgagedEvent: HandlePropertyUnmortgagedEvent,
      PlayerExchangeDueEvent: HandlePlayerExchangeDueEvent,
      PlayerExchangeEvent: HandlePlayerExchangeEvent,
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
