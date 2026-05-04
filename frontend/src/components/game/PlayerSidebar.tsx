"use client"

import { useState, useEffect } from "react"
import { getTokenIcon, getTokenName } from "@/utils/tokens"
import { Player, GameState, OwnedProperty } from "@/types"
import { useEndTurn } from "@/hooks/playerHooks"
import { storage } from "@/utils"

interface PlayerSidebarProps {
  sessionId: string
  playerId: string
  playerName: string
  players: Player[]
  currentPlayerTurnId?: number | null
  gameState?: GameState | null
}

export default function PlayerSidebar({
  sessionId,
  playerId,
  playerName,
  players,
  currentPlayerTurnId,
  gameState,
}: PlayerSidebarProps) {
  const [joinCode, setJoinCode] = useState<string>("")
  const [actionError, setActionError] = useState<string | null>(null)
  const endTurnMutation = useEndTurn()
  const propertyOrder = [
    "BROWN",
    "LIGHTBLUE",
    "PINK",
    "ORANGE",
    "RED",
    "YELLOW",
    "GREEN",
    "DARKBLUE",
    "RAILROAD",
    "UTILITY",
  ]

  const getPropertyColor = (property: OwnedProperty) => {
    const propertyType = property.property_info.property_type

    switch (propertyType) {
      case "BROWN":
        return "#8B5A2B"
      case "LIGHTBLUE":
        return "#87CEEB"
      case "PINK":
        return "#FF69B4"
      case "ORANGE":
        return "#F5A623"
      case "RED":
        return "#D32F2F"
      case "YELLOW":
        return "#F7D54A"
      case "GREEN":
        return "#4CAF50"
      case "DARKBLUE":
        return "#1E3A8A"
      case "RAILROAD":
        return "#000000"
      case "UTILITY":
        return "#FFFFFF"
      default:
        return "#FFFFFF"
    }
  }

  const isCurrentPlayerTurn = currentPlayerTurnId?.toString() === playerId
  const isGameStarted = (gameState?.current_turn ?? -1) >= 0
  const isTurnOrderPhase = (gameState?.current_turn ?? -1) >= 0 && !!gameState?.players.some((playerInfo) => playerInfo.player.player_order === -1)

  const currentPlayer = players.find((p) => p.id === currentPlayerTurnId)

  useEffect(() => {
    const code = storage.getGameCode()
    if (code) setJoinCode(code)
  }, [])

  const handleEndTurn = () => {
    if (!isCurrentPlayerTurn) return

    setActionError(null)
    endTurnMutation.mutate(undefined, {
      onError: (error: Error) => {
        setActionError(error.message)
      },
    })
  }

  return (
    <div className="w-full h-full flex flex-col p-4 overflow-y-auto" style={{ backgroundColor: "#FFFFFF" }}>
      <div className="flex flex-col text-center border-2 px-3 py-1 mb-3" style={{ borderColor: "#D0D3D4" }}>
        <div className="text-xs" style={{ color: "#7C878E" }}>
          Game Join Code
        </div>
        <div className="text-lg font-bold" style={{ color: "#F76902" }}>
          {joinCode || "..."}
        </div>
      </div>
      <h3 className="text-xl font-bold mb-4" style={{ color: "#F76902" }}>
        Players
      </h3>

      {/* Current Turn Info */}
      <div
        className="mb-4 p-3 border-2"
        style={{
          borderColor: isGameStarted && isCurrentPlayerTurn ? "#00AA00" : "#D0D3D4",
          backgroundColor: isGameStarted && isCurrentPlayerTurn ? "#E8F5E9" : "#F9F9F9",
        }}
      >
        <div className="text-xs font-bold mb-2" style={{ color: "#7C878E" }}>
          {isGameStarted ? (isTurnOrderPhase ? "TURN ORDER" : "CURRENT TURN") : "LOBBY"}
        </div>
        <div className="text-lg font-bold mb-2" style={{ color: isGameStarted && isCurrentPlayerTurn ? "#00AA00" : "#F76902" }}>
          {isGameStarted ? (currentPlayer ? currentPlayer.name : "Waiting...") : "Waiting"}
        </div>

        {isGameStarted && isCurrentPlayerTurn && (
          <button
            type="button"
            onClick={handleEndTurn}
            disabled={endTurnMutation.isPending}
            style={{
              width: "100%",
              padding: "10px 12px",
              marginBottom: 12,
              backgroundColor: endTurnMutation.isPending ? "#D0D3D4" : "#F76902",
              color: "#FFFFFF",
              fontWeight: 700,
              cursor: endTurnMutation.isPending ? "not-allowed" : "pointer",
            }}
          >
            End Turn
          </button>
        )}

        {isGameStarted && !isCurrentPlayerTurn && currentPlayer && (
          <div className="text-xs mb-3" style={{ color: "#7C878E" }}>
            Waiting for {currentPlayer.name} to play...
          </div>
        )}

        {!isGameStarted && (
          <div className="text-xs mb-3" style={{ color: "#7C878E" }}>
            Ready up to start
          </div>
        )}
      </div>

      {/* Actions Section */}
      <div className="mb-4 p-3 border-2 flex flex-col gap-4" style={{ borderColor: "#D0D3D4" }}>
        {!isGameStarted ? (
          <div className="text-sm" style={{ color: "#7C878E" }}>
            Use the center panel to ready up.
          </div>
        ) : (
          <>
            <div className="text-sm" style={{ color: "#7C878E" }}>
              Use the center panel for turn actions.
            </div>
            {actionError ? (
              <div className="text-xs" style={{ color: "#D32F2F" }}>
                {actionError}
              </div>
            ) : null}
          </>
        )}
      </div>

      {/* Players List */}
      <div className="space-y-3 flex-1">
        <div className="text-xs font-bold mb-2" style={{ color: "#7C878E" }}>
          PLAYERS ({players.length})
        </div>

        {players.length === 0 ? (
          <div style={{ color: "#7C878E" }} className="text-sm">
            Waiting for players...
          </div>
        ) : (
          players.map((player) => {
            const isCurrentPlayer = player.id.toString() === playerId
            const isPlayerTurn = player.id === currentPlayerTurnId
            const ownedProperties =
              gameState?.players
                .find((playerInfo) => playerInfo.player.id === player.id)
                ?.owned_properties.slice()
                .sort((a, b) => {
                  const typeDiff =
                    propertyOrder.indexOf(a.property_info.property_type) -
                    propertyOrder.indexOf(b.property_info.property_type)

                  if (typeDiff !== 0) {
                    return typeDiff
                  }

                  return a.property_info.purchase_cost - b.property_info.purchase_cost
                }) ?? []

            return (
              <div
                key={player.id}
                className="border-2 p-3"
                style={{
                  borderColor: isPlayerTurn ? "#FFD700" : isCurrentPlayer ? "#F76902" : "#D0D3D4",
                  backgroundColor: isPlayerTurn ? "#FFFACD" : isCurrentPlayer ? "#FFF3E0" : "#FFFFFF",
                  borderWidth: isPlayerTurn ? "3px" : "2px",
                  boxShadow: isPlayerTurn ? "0 0 8px #FFD700" : "none",
                }}
              >
                {/* Player token icon and name */}
                <div
                  className="font-bold mb-2 flex items-center gap-2"
                  style={{
                    color: isPlayerTurn ? "#F57F17" : isCurrentPlayer ? "#000000" : "#000000",
                  }}
                >
                  <img
                    src={getTokenIcon(player.piece_token)}
                    alt={getTokenName(player.piece_token)}
                    style={{
                      width: "18px",
                      height: "18px",
                      border: isPlayerTurn ? "2px solid #FFD700" : "1px solid #000",
                      borderRadius: "2px",
                    }}
                    title={getTokenName(player.piece_token)}
                  />
                  <span>{player.name}</span>
                  {isCurrentPlayer && !isPlayerTurn && (
                    <span style={{ fontSize: "0.85em", color: "#F76902" }}>(you)</span>
                  )}
                  {isPlayerTurn && (
                    <span style={{ fontSize: "0.75em", color: "#F57F17", fontWeight: "bold" }}>PLAYING</span>
                  )}
                  <span
                    title={player.in_game ? "In Game" : "Offline"}
                    style={{
                      width: 10,
                      height: 10,
                      borderRadius: "999px",
                      backgroundColor: player.in_game ? "#00AA00" : "#D32F2F",
                      display: "inline-block",
                      flexShrink: 0,
                    }}
                  />
                </div>

                {/* Money */}
                <div
                  className="text-sm mb-2"
                  style={{
                    color: isPlayerTurn ? "#F57F17" : isCurrentPlayer ? "#000000" : "#7C878E",
                  }}
                >
                  Money: ₮{player.money.toLocaleString()}
                </div>

                {/* Properties placeholder */}
                <div
                  className="text-xs mt-2 p-2 border-t-2"
                  style={{
                    borderColor: isPlayerTurn ? "#FFD700" : isCurrentPlayer ? "#000000" : "#D0D3D4",
                    color: isPlayerTurn ? "#F57F17" : isCurrentPlayer ? "#000000" : "#7C878E",
                  }}
                >
                  {ownedProperties.length === 0 ? (
                    <div>Properties: None</div>
                  ) : (
                    <div className="flex flex-wrap gap-2">
                      {ownedProperties.map((property) => (
                        <div
                          key={property.id}
                          title={property.property_info.name}
                          style={{
                            width: 14,
                            height: 14,
                            border: "1px solid #000000",
                            backgroundColor: getPropertyColor(property),
                            flexShrink: 0,
                          }}
                        />
                      ))}
                    </div>
                  )}
                </div>
              </div>
            )
          })
        )}
      </div>
    </div>
  )
}
