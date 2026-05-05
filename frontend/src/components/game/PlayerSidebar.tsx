"use client"

import { useState, useEffect, useMemo, useRef } from "react"
import { getTokenIcon, getTokenName } from "@/utils/tokens"
import { Player, GameState, OwnedProperty } from "@/types"
import { useEndTurn } from "@/hooks/playerHooks"
import {
  useMortgageProperty,
  usePurchaseHouse,
  usePurchaseHotel,
  useSellHouse,
  useSellHotel,
  useUnmortgageProperty,
} from "@/hooks/propertyHooks"
import { storage } from "@/utils"
import { emitToast } from "@/utils/toast"

interface PlayerSidebarProps {
  sessionId: string
  playerId: string
  playerName: string
  players: Player[]
  currentPlayerTurnId?: number | null
  gameState?: GameState | null
  activePropertyId?: number | null
  onSelectProperty: (propertyId: number | null) => void
  onOpenTrade: (tradePlayerId: number) => void
}

export default function PlayerSidebar({
  sessionId,
  playerId,
  playerName,
  players,
  currentPlayerTurnId,
  gameState,
  activePropertyId,
  onSelectProperty,
  onOpenTrade,
}: PlayerSidebarProps) {
  const [joinCode, setJoinCode] = useState<string>("")
  const popupRef = useRef<HTMLDivElement | null>(null)
  const endTurnMutation = useEndTurn()
  const purchaseHouseMutation = usePurchaseHouse()
  const purchaseHotelMutation = usePurchaseHotel()
  const sellHouseMutation = useSellHouse()
  const sellHotelMutation = useSellHotel()
  const mortgagePropertyMutation = useMortgageProperty()
  const unmortgagePropertyMutation = useUnmortgageProperty()
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
  const selectedProperty = useMemo(() => {
    if (!activePropertyId || !gameState) return null

    for (const playerInfo of gameState.players) {
      const ownedProperty = (playerInfo.owned_properties ?? []).find(
        (property) => property.property_info.id === activePropertyId,
      )
      if (ownedProperty) {
        return ownedProperty
      }
    }

    return null
  }, [activePropertyId, gameState])
  const getRankLabel = (rank: number) => {
    const value = Math.abs(rank)
    const mod100 = value % 100

    if (mod100 >= 11 && mod100 <= 13) {
      return `${rank}th place`
    }

    switch (value % 10) {
      case 1:
        return `${rank}st place`
      case 2:
        return `${rank}nd place`
      case 3:
        return `${rank}rd place`
      default:
        return `${rank}th place`
    }
  }

  useEffect(() => {
    const code = storage.getGameCode()
    if (code) setJoinCode(code)
  }, [])

  useEffect(() => {
    if (!activePropertyId) return
    if (selectedProperty) return
    onSelectProperty(null)
  }, [activePropertyId, onSelectProperty, selectedProperty])

  useEffect(() => {
    const handlePointerDown = (event: MouseEvent) => {
      if (popupRef.current && !popupRef.current.contains(event.target as Node)) {
        onSelectProperty(null)
      }
    }

    document.addEventListener("mousedown", handlePointerDown)
    return () => document.removeEventListener("mousedown", handlePointerDown)
  }, [onSelectProperty])

  const handleEndTurn = () => {
    if (!isCurrentPlayerTurn) return

    endTurnMutation.mutate(undefined, {
      onError: (error: Error) => {
        emitToast(error.message)
      },
    })
  }

  const handlePropertyAction = (
    propertyId: number,
    mutation:
      | typeof purchaseHouseMutation
      | typeof purchaseHotelMutation
      | typeof sellHouseMutation
      | typeof sellHotelMutation
      | typeof mortgagePropertyMutation
      | typeof unmortgagePropertyMutation,
  ) => {
    mutation.mutate(propertyId, {
      onSuccess: () => {
        onSelectProperty(null)
      },
      onError: (error: Error) => {
        emitToast(error.message)
      },
    })
  }

  const getRailroadRent = (properties: OwnedProperty[], property: OwnedProperty) => {
    if (property.is_mortgaged) {
      return 0
    }

    const railroadCount = properties.filter(
      (ownedProperty) =>
        ownedProperty.property_info.property_type === "RAILROAD" && !ownedProperty.is_mortgaged,
    ).length

    if (railroadCount < 1) {
      return 0
    }

    return 25 * 2 ** (railroadCount - 1)
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
              (gameState?.players
                .find((playerInfo) => playerInfo.player.id === player.id)
                ?.owned_properties ?? [])
                .slice()
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
                      objectFit: "cover",
                      objectPosition: "top",
                    }}
                    title={getTokenName(player.piece_token)}
                  />
                  <span>{player.name}</span>
                  {player.rank > 0 ? (
                    <span style={{ fontSize: "0.75em", color: "#7C878E", fontWeight: 700 }}>
                      {getRankLabel(player.rank)}
                    </span>
                  ) : null}
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

                {isCurrentPlayerTurn && isGameStarted && !isCurrentPlayer && !player.bankrupt ? (
                  <button
                    type="button"
                    onClick={() => onOpenTrade(player.id)}
                    disabled={!!gameState?.pending_trade || !!gameState?.pending_trade_draft}
                    style={{
                      width: "100%",
                      padding: "8px 10px",
                      marginBottom: 10,
                      backgroundColor: gameState?.pending_trade || gameState?.pending_trade_draft ? "#D0D3D4" : "#F76902",
                      color: "#FFFFFF",
                      fontWeight: 700,
                      cursor: gameState?.pending_trade || gameState?.pending_trade_draft ? "not-allowed" : "pointer",
                    }}
                  >
                    Trade
                  </button>
                ) : null}

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
                      {ownedProperties.map((property) => {
                        const propertyType = property.property_info.property_type
                        const isUtility = propertyType === "UTILITY"
                        const isRailroad = propertyType === "RAILROAD"
                        const isBuildable = !isUtility && !isRailroad
                        const propertyRent = property.is_mortgaged
                          ? 0
                          : isRailroad
                            ? getRailroadRent(ownedProperties, property)
                            : property.current_rent

                        return (
                        <div
                          key={property.id}
                          style={{
                            position: "relative",
                          }}
                        >
                              <button
                                type="button"
                                title={property.property_info.name}
                                onClick={() => {
                                  onSelectProperty(
                                    activePropertyId === property.property_info.id ? null : property.property_info.id,
                                  )
                                }}
                                style={{
                                  width: 14,
                                  height: 14,
                                  border: "1px solid #000000",
                                  backgroundColor: getPropertyColor(property),
                                  flexShrink: 0,
                                  display: "block",
                                  cursor: "pointer",
                                  boxShadow: activePropertyId === property.property_info.id ? "0 0 0 2px #F76902" : "none",
                                }}
                              />

                          {activePropertyId === property.property_info.id ? (
                            <div
                              ref={popupRef}
                              style={{
                                position: "absolute",
                                left: 24,
                                top: "50%",
                                transform: "translateY(-50%)",
                                zIndex: 20,
                                width: 180,
                                backgroundColor: "#FFFFFF",
                                border: "1px solid #000000",
                                padding: 10,
                                display: "flex",
                                flexDirection: "column",
                                gap: 8,
                              }}
                            >
                              <div
                                style={{
                                  position: "absolute",
                                  left: -8,
                                  top: "50%",
                                  width: 0,
                                  height: 0,
                                  borderTop: "8px solid transparent",
                                  borderBottom: "8px solid transparent",
                                  borderRight: "8px solid #000000",
                                  transform: "translateY(-50%)",
                                }}
                              />
                              <div
                                style={{
                                  position: "absolute",
                                  left: -7,
                                  top: "50%",
                                  width: 0,
                                  height: 0,
                                  borderTop: "7px solid transparent",
                                  borderBottom: "7px solid transparent",
                                  borderRight: "7px solid #FFFFFF",
                                  transform: "translateY(-50%)",
                                }}
                              />

                              <div
                                style={{
                                  color: "#000000",
                                  fontSize: 12,
                                  fontWeight: 700,
                                }}
                              >
                                {property.property_info.name}
                              </div>

                              <div
                                style={{
                                  color: "#7C878E",
                                  fontSize: 11,
                                  fontWeight: 700,
                                }}
                              >
                                Owner: {player.name}
                              </div>

                              <div
                                style={{
                                  color: "#7C878E",
                                  fontSize: 11,
                                  fontWeight: 700,
                                }}
                              >
                                Mortgaged: {property.is_mortgaged ? "Yes" : "No"}
                              </div>

                              {!isUtility ? (
                                <div
                                  style={{
                                    color: "#7C878E",
                                    fontSize: 11,
                                    fontWeight: 700,
                                  }}
                                >
                                  Rent: ₮{propertyRent.toLocaleString()}
                                </div>
                              ) : null}

                              {isBuildable ? (
                                <>
                                  <div
                                    style={{
                                      color: "#7C878E",
                                      fontSize: 11,
                                      fontWeight: 700,
                                    }}
                                  >
                                    Houses: {property.houses}
                                  </div>

                                  <div
                                    style={{
                                      color: "#7C878E",
                                      fontSize: 11,
                                      fontWeight: 700,
                                    }}
                                  >
                                    Hotel: {property.has_hotel ? "Yes" : "No"}
                                  </div>
                                </>
                              ) : null}

                              {isCurrentPlayer ? (
                                <>
                                  {isBuildable ? (
                                    <>
                                      <button
                                        type="button"
                                        onClick={() => handlePropertyAction(property.property_info.id, purchaseHouseMutation)}
                                        disabled={!isCurrentPlayerTurn || purchaseHouseMutation.isPending}
                                        style={{
                                          width: "100%",
                                          padding: "8px 10px",
                                          backgroundColor: !isCurrentPlayerTurn || purchaseHouseMutation.isPending ? "#D0D3D4" : "#F76902",
                                          color: "#FFFFFF",
                                          fontWeight: 700,
                                          cursor: !isCurrentPlayerTurn || purchaseHouseMutation.isPending ? "not-allowed" : "pointer",
                                        }}
                                      >
                                        Buy House
                                      </button>

                                      <button
                                        type="button"
                                        onClick={() => handlePropertyAction(property.property_info.id, sellHouseMutation)}
                                        disabled={!isCurrentPlayerTurn || sellHouseMutation.isPending}
                                        style={{
                                          width: "100%",
                                          padding: "8px 10px",
                                          backgroundColor: !isCurrentPlayerTurn || sellHouseMutation.isPending ? "#D0D3D4" : "#D0D3D4",
                                          color: "#000000",
                                          fontWeight: 700,
                                          cursor: !isCurrentPlayerTurn || sellHouseMutation.isPending ? "not-allowed" : "pointer",
                                          opacity: !isCurrentPlayerTurn || sellHouseMutation.isPending ? 0.7 : 1,
                                        }}
                                      >
                                        Sell House
                                      </button>

                                      <button
                                        type="button"
                                        onClick={() => handlePropertyAction(property.property_info.id, purchaseHotelMutation)}
                                        disabled={!isCurrentPlayerTurn || purchaseHotelMutation.isPending}
                                        style={{
                                          width: "100%",
                                          padding: "8px 10px",
                                          backgroundColor: !isCurrentPlayerTurn || purchaseHotelMutation.isPending ? "#D0D3D4" : "#F76902",
                                          color: "#FFFFFF",
                                          fontWeight: 700,
                                          cursor: !isCurrentPlayerTurn || purchaseHotelMutation.isPending ? "not-allowed" : "pointer",
                                        }}
                                      >
                                        Buy Hotel
                                      </button>

                                      <button
                                        type="button"
                                        onClick={() => handlePropertyAction(property.property_info.id, sellHotelMutation)}
                                        disabled={!isCurrentPlayerTurn || sellHotelMutation.isPending}
                                        style={{
                                          width: "100%",
                                          padding: "8px 10px",
                                          backgroundColor: !isCurrentPlayerTurn || sellHotelMutation.isPending ? "#D0D3D4" : "#D0D3D4",
                                          color: "#000000",
                                          fontWeight: 700,
                                          cursor: !isCurrentPlayerTurn || sellHotelMutation.isPending ? "not-allowed" : "pointer",
                                          opacity: !isCurrentPlayerTurn || sellHotelMutation.isPending ? 0.7 : 1,
                                        }}
                                      >
                                        Sell Hotel
                                      </button>
                                    </>
                                  ) : null}

                                  <button
                                    type="button"
                                    onClick={() =>
                                      handlePropertyAction(
                                        property.property_info.id,
                                        property.is_mortgaged ? unmortgagePropertyMutation : mortgagePropertyMutation,
                                      )
                                    }
                                    disabled={
                                      !isCurrentPlayerTurn ||
                                      mortgagePropertyMutation.isPending ||
                                      unmortgagePropertyMutation.isPending
                                    }
                                    style={{
                                      width: "100%",
                                      padding: "8px 10px",
                                      backgroundColor:
                                        !isCurrentPlayerTurn ||
                                        mortgagePropertyMutation.isPending ||
                                        unmortgagePropertyMutation.isPending
                                          ? "#D0D3D4"
                                          : "#D0D3D4",
                                      color: "#000000",
                                      fontWeight: 700,
                                      cursor:
                                        !isCurrentPlayerTurn ||
                                        mortgagePropertyMutation.isPending ||
                                        unmortgagePropertyMutation.isPending
                                          ? "not-allowed"
                                          : "pointer",
                                      opacity:
                                        !isCurrentPlayerTurn ||
                                        mortgagePropertyMutation.isPending ||
                                        unmortgagePropertyMutation.isPending
                                          ? 0.7
                                          : 1,
                                    }}
                                  >
                                    {property.is_mortgaged ? "Unmortgage" : "Mortgage"}
                                  </button>
                                </>
                              ) : null}
                            </div>
                          ) : null}
                        </div>
                        )
                      })}
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
