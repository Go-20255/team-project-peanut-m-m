"use client"

import { useEffect, useRef } from "react"

const API_URL = process.env.NEXT_PUBLIC_API_URL

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
      withCredentials: true,
    })

    console.log("SSE connection established for session:", sessionId, "player:", playerId)

    // Special case: InitialGameBoardDataEvent contains full game state
    eventSource.addEventListener("InitialGameBoardDataEvent", (event: any) => {
      handleSSEEvent(event.data, "InitialGameBoardDataEvent", (_, data) => {
        // data shape: { tiles, current_turn, players: PlayerInfo[] }
        onUpdateRef.current({
          type: "initialGameState",
          data: {
            current_turn: data.current_turn,
            tiles: data.tiles,
            players: data.players,
          },
        })
      })
    })

    //// Map event names to their corresponding update types
    //const eventMap: { [key: string]: string } = {
      //RollDiceEvent: "diceRoll",
      //MovePlayerEvent: "playerMove",
      //RentPaidEvent: "rentPaid",
      //PropertyPurchased: "propertyPurchased",
      //HousePurchased: "housePurchased",
      //HotelPurchased: "hotelPurchased",
      //HouseSold: "houseSold",
      //HotelSold: "hotelSold",
      //PropertyMortgaged: "propertyMortgaged",
      //PropertyUnmortgaged: "propertyUnmortgaged",
    //}

    //Object.entries(eventMap).forEach(([eventName, updateType]) => {
      //eventSource.addEventListener(eventName, (event: any) => {
        //handleSSEEvent(event.data, eventName, (_, data) => {
          //onUpdateRef.current({ type: updateType, data })
        //})
      //})
    //})

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
