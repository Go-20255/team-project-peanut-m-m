"use client"

import { useEffect, useMemo, useState } from "react"
import { GameState, OwnedProperty } from "@/types"
import {
  useAcceptTrade,
  useCancelTrade,
  useCloseTradeDraft,
  useOpenTradeDraft,
  useProposeTrade,
  useRejectTrade,
} from "@/hooks/useGameAPI"
import { emitToast } from "@/utils/toast"

interface TradeOverlayProps {
  playerId: string
  currentPlayerTurnId?: number | null
  gameState: GameState
  openTradePlayerId: number | null
  onCloseTrade: () => void
}

function getPropertyColor(property: OwnedProperty) {
  switch (property.property_info.property_type) {
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

function TradePropertySquare({
  property,
  selected,
  onClick,
}: {
  property: OwnedProperty
  selected: boolean
  onClick?: () => void
}) {
  return (
    <button
      type="button"
      title={property.is_mortgaged ? `${property.property_info.name} (Mortgaged)` : property.property_info.name}
      onClick={onClick}
      style={{
        width: 18,
        height: 18,
        border: "1px solid #000000",
        backgroundColor: getPropertyColor(property),
        boxShadow: selected ? "0 0 0 2px #F76902" : "none",
        cursor: onClick ? "pointer" : "default",
        position: "relative",
        flexShrink: 0,
      }}
    >
      {property.is_mortgaged ? (
        <span
          style={{
            position: "absolute",
            right: 1,
            top: 8,
            width: 20,
            height: 2,
            backgroundColor: "#D32F2F",
            transform: "rotate(-45deg)",
            transformOrigin: "right center",
            pointerEvents: "none",
          }}
        />
      ) : null}
    </button>
  )
}

export default function TradeOverlay({
  playerId,
  currentPlayerTurnId,
  gameState,
  openTradePlayerId,
  onCloseTrade,
}: TradeOverlayProps) {
  const openTradeDraftMutation = useOpenTradeDraft()
  const closeTradeDraftMutation = useCloseTradeDraft()
  const proposeTradeMutation = useProposeTrade()
  const acceptTradeMutation = useAcceptTrade()
  const rejectTradeMutation = useRejectTrade()
  const cancelTradeMutation = useCancelTrade()
  const [openingTradePlayerId, setOpeningTradePlayerId] = useState<number | null>(null)
  const [offeredMoney, setOfferedMoney] = useState("0")
  const [requestedMoney, setRequestedMoney] = useState("0")
  const [offeredPropertyIds, setOfferedPropertyIds] = useState<number[]>([])
  const [requestedPropertyIds, setRequestedPropertyIds] = useState<number[]>([])

  const pendingTradeDraft = gameState.pending_trade_draft ?? null
  const pendingTrade = gameState.pending_trade ?? null
  const currentPlayerId = Number(playerId)
  const isCurrentPlayerTurn = currentPlayerTurnId === currentPlayerId

  const currentPlayerInfo = useMemo(
    () => gameState.players.find((playerInfo) => playerInfo.player.id === currentPlayerId) ?? null,
    [currentPlayerId, gameState.players],
  )

  const draftTargetInfo = useMemo(
    () => gameState.players.find((playerInfo) => playerInfo.player.id === openTradePlayerId) ?? null,
    [gameState.players, openTradePlayerId],
  )

  const draftFromInfo = useMemo(
    () => gameState.players.find((playerInfo) => playerInfo.player.id === pendingTradeDraft?.from_player_id) ?? null,
    [gameState.players, pendingTradeDraft?.from_player_id],
  )

  const draftToInfo = useMemo(
    () => gameState.players.find((playerInfo) => playerInfo.player.id === pendingTradeDraft?.to_player_id) ?? null,
    [gameState.players, pendingTradeDraft?.to_player_id],
  )

  const pendingFromInfo = useMemo(
    () => gameState.players.find((playerInfo) => playerInfo.player.id === pendingTrade?.from_player_id) ?? null,
    [gameState.players, pendingTrade?.from_player_id],
  )

  const pendingToInfo = useMemo(
    () => gameState.players.find((playerInfo) => playerInfo.player.id === pendingTrade?.to_player_id) ?? null,
    [gameState.players, pendingTrade?.to_player_id],
  )

  const isDraftVisible = !pendingTrade && (!!pendingTradeDraft || !!openTradePlayerId)
  const isDraftComposer =
    isDraftVisible &&
    (pendingTradeDraft ? pendingTradeDraft.from_player_id === currentPlayerId : !!openTradePlayerId) &&
    isCurrentPlayerTurn
  const isTradeTarget = pendingTrade?.to_player_id === currentPlayerId
  const isTradeProposer = pendingTrade?.from_player_id === currentPlayerId
  const visible = !!pendingTrade || isDraftVisible

  useEffect(() => {
    if (!isDraftComposer) {
      setOfferedMoney("0")
      setRequestedMoney("0")
      setOfferedPropertyIds([])
      setRequestedPropertyIds([])
    }
  }, [isDraftComposer])

  useEffect(() => {
    if (!openTradePlayerId || pendingTrade) {
      setOpeningTradePlayerId(null)
      return
    }

    if (
      pendingTradeDraft &&
      pendingTradeDraft.from_player_id === currentPlayerId &&
      pendingTradeDraft.to_player_id === openTradePlayerId
    ) {
      setOpeningTradePlayerId(null)
      return
    }

    if (openingTradePlayerId === openTradePlayerId || openTradeDraftMutation.isPending) {
      return
    }

    setOpeningTradePlayerId(openTradePlayerId)
    openTradeDraftMutation.mutate(
      { withPlayerId: openTradePlayerId },
      {
        onSuccess: () => {
          setOpeningTradePlayerId(null)
        },
        onError: (error: Error) => {
          setOpeningTradePlayerId(null)
          emitToast(error.message)
          onCloseTrade()
        },
      },
    )
  }, [
    currentPlayerId,
    openTradeDraftMutation,
    openTradePlayerId,
    onCloseTrade,
    openingTradePlayerId,
    pendingTrade,
    pendingTradeDraft,
  ])

  const toggleProperty = (propertyId: number, side: "offered" | "requested") => {
    if (side === "offered") {
      setOfferedPropertyIds((value) =>
        value.includes(propertyId) ? value.filter((id) => id !== propertyId) : [...value, propertyId],
      )
      return
    }

    setRequestedPropertyIds((value) =>
      value.includes(propertyId) ? value.filter((id) => id !== propertyId) : [...value, propertyId],
    )
  }

  const handlePropose = () => {
    const targetPlayerInfo = pendingTradeDraft ? draftToInfo : draftTargetInfo
    if (!targetPlayerInfo) return

    proposeTradeMutation.mutate(
      {
        withPlayerId: targetPlayerInfo.player.id,
        offeredMoney: Number(offeredMoney || "0"),
        requestedMoney: Number(requestedMoney || "0"),
        offeredPropertyIds,
        requestedPropertyIds,
      },
      {
        onSuccess: () => {
          onCloseTrade()
        },
        onError: (error: Error) => {
          emitToast(error.message)
        },
      },
    )
  }

  const handleAccept = () => {
    acceptTradeMutation.mutate(undefined, {
      onSuccess: () => {
        onCloseTrade()
      },
      onError: (error: Error) => {
        emitToast(error.message)
      },
    })
  }

  const handleReject = () => {
    rejectTradeMutation.mutate(undefined, {
      onSuccess: () => {
        onCloseTrade()
      },
      onError: (error: Error) => {
        emitToast(error.message)
      },
    })
  }

  const handleCancel = () => {
    if (pendingTrade) {
      cancelTradeMutation.mutate(undefined, {
        onSuccess: () => {
          onCloseTrade()
        },
        onError: (error: Error) => {
          emitToast(error.message)
        },
      })
      return
    }

    if (pendingTradeDraft && pendingTradeDraft.from_player_id === currentPlayerId) {
      closeTradeDraftMutation.mutate(undefined, {
        onSuccess: () => {
          onCloseTrade()
        },
        onError: (error: Error) => {
          emitToast(error.message)
        },
      })
      return
    }

    onCloseTrade()
  }

  if (!visible) {
    return null
  }

  const leftInfo = pendingTrade
    ? pendingFromInfo
    : pendingTradeDraft
      ? draftFromInfo
      : currentPlayerInfo
  const rightInfo = pendingTrade
    ? pendingToInfo
    : pendingTradeDraft
      ? draftToInfo
      : draftTargetInfo

  if (!leftInfo || !rightInfo) {
    return null
  }

  const leftProperties = ((leftInfo.owned_properties ?? []).slice()).sort(
    (a, b) => a.property_info.purchase_cost - b.property_info.purchase_cost,
  )
  const rightProperties = ((rightInfo.owned_properties ?? []).slice()).sort(
    (a, b) => a.property_info.purchase_cost - b.property_info.purchase_cost,
  )

  const leftSelectedIds = pendingTrade
    ? new Set(pendingTrade.offered_properties.map((property) => property.property_id))
    : new Set(offeredPropertyIds)
  const rightSelectedIds = pendingTrade
    ? new Set(pendingTrade.requested_properties.map((property) => property.property_id))
    : new Set(requestedPropertyIds)

  const tradeTitle = pendingTrade
    ? pendingFromInfo && pendingToInfo
      ? `${pendingFromInfo.player.name} proposed a trade`
      : "Trade"
    : rightInfo
      ? `Trade with ${rightInfo.player.name}`
      : "Trade"

  const showDraftNoticeOnly = !!pendingTradeDraft && !isDraftComposer
  const showTradeDetails = pendingTrade || isDraftComposer
  const closeLabel = pendingTrade && isTradeProposer ? "Cancel Trade" : "Close"

  return (
    <div
      style={{
        position: "fixed",
        inset: 0,
        zIndex: 40,
        pointerEvents: "none",
      }}
    >
      <div
        style={{
          position: "absolute",
          inset: 0,
          backgroundColor: "rgba(0, 0, 0, 0.12)",
        }}
      />

      <div
        style={{
          position: "absolute",
          left: "50%",
          top: "50%",
          transform: "translate(-50%, -50%)",
          width: "min(760px, calc(100vw - 48px))",
          backgroundColor: "#FFFFFF",
          border: "2px solid #000000",
          padding: 20,
          display: "flex",
          flexDirection: "column",
          gap: 18,
          pointerEvents: "auto",
        }}
      >
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            gap: 12,
          }}
        >
          <div>
            <div
              style={{
                color: "#F76902",
                fontSize: 22,
                fontWeight: 700,
              }}
            >
              Trade
            </div>
            <div
              style={{
                color: "#7C878E",
                fontSize: 13,
                fontWeight: 600,
              }}
            >
              {tradeTitle}
            </div>
          </div>

          {!showDraftNoticeOnly ? (
            <button
              type="button"
              onClick={handleCancel}
              disabled={closeTradeDraftMutation.isPending || cancelTradeMutation.isPending}
              style={{
                padding: "8px 12px",
                backgroundColor: "#D0D3D4",
                color: "#000000",
                fontWeight: 700,
                cursor: closeTradeDraftMutation.isPending || cancelTradeMutation.isPending ? "not-allowed" : "pointer",
              }}
            >
              {closeLabel}
            </button>
          ) : null}
        </div>

        {showTradeDetails ? (
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr 1fr",
              gap: 16,
            }}
          >
            {[leftInfo, rightInfo].map((playerInfo, index) => {
              const isLeft = index === 0
              const properties = isLeft ? leftProperties : rightProperties
              const selectedIds = isLeft ? leftSelectedIds : rightSelectedIds
              const moneyValue = pendingTrade
                ? isLeft
                  ? pendingTrade.offered_money
                  : pendingTrade.requested_money
                : isLeft
                  ? offeredMoney
                  : requestedMoney

              return (
                <div
                  key={playerInfo.player.id}
                  style={{
                    border: "1px solid #D0D3D4",
                    padding: 14,
                    display: "flex",
                    flexDirection: "column",
                    gap: 10,
                  }}
                >
                  <div
                    style={{
                      color: "#000000",
                      fontSize: 16,
                      fontWeight: 700,
                    }}
                  >
                    {playerInfo.player.name}
                  </div>

                  <div
                    style={{
                      color: "#7C878E",
                      fontSize: 12,
                      fontWeight: 700,
                    }}
                  >
                    Money in trade
                  </div>

                  {pendingTrade ? (
                    <div
                      style={{
                        border: "1px solid #D0D3D4",
                        padding: "10px 12px",
                        color: "#000000",
                        fontSize: 14,
                        fontWeight: 700,
                      }}
                    >
                      ₮{Number(moneyValue).toLocaleString()}
                    </div>
                  ) : (
                    <input
                      type="number"
                      min="0"
                      value={moneyValue}
                      onChange={(event) =>
                        isLeft ? setOfferedMoney(event.target.value) : setRequestedMoney(event.target.value)
                      }
                      style={{
                        border: "1px solid #D0D3D4",
                        padding: "10px 12px",
                        color: "#000000",
                      }}
                    />
                  )}

                  <div
                    style={{
                      color: "#7C878E",
                      fontSize: 12,
                      fontWeight: 700,
                    }}
                  >
                    Properties
                  </div>

                  {properties.length === 0 ? (
                    <div
                      style={{
                        color: "#7C878E",
                        fontSize: 12,
                      }}
                    >
                      None
                    </div>
                  ) : (
                    <div
                      style={{
                        display: "flex",
                        flexWrap: "wrap",
                        gap: 8,
                      }}
                    >
                      {properties.map((property) => (
                        <TradePropertySquare
                          key={property.id}
                          property={property}
                          selected={selectedIds.has(property.property_info.id)}
                          onClick={
                            pendingTrade
                              ? undefined
                              : () => toggleProperty(property.property_info.id, isLeft ? "offered" : "requested")
                          }
                        />
                      ))}
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        ) : (
          <div
            style={{
              border: "1px solid #D0D3D4",
              padding: 18,
              display: "flex",
              flexDirection: "column",
              gap: 10,
            }}
          >
            <div
              style={{
                color: "#000000",
                fontSize: 16,
                fontWeight: 700,
                textAlign: "center",
              }}
            >
              {draftFromInfo?.player.name} is proposing a trade...
            </div>
            <div
              style={{
                color: "#7C878E",
                fontSize: 13,
                fontWeight: 600,
                textAlign: "center",
              }}
            >
              With {draftToInfo?.player.name}
            </div>
          </div>
        )}

        {pendingTrade ? (
          <div
            style={{
              color: "#7C878E",
              fontSize: 13,
              fontWeight: 600,
              textAlign: "center",
            }}
          >
            {isTradeTarget
              ? "Choose whether to accept or reject this trade."
              : pendingTrade.from_player_id === currentPlayerId
                ? "Waiting for a response."
                : `${pendingFromInfo?.player.name ?? "A player"} is proposing a trade...`}
          </div>
        ) : showDraftNoticeOnly ? (
          <div
            style={{
              color: "#7C878E",
              fontSize: 13,
              fontWeight: 600,
              textAlign: "center",
            }}
          >
            {draftFromInfo?.player.name ?? "A player"} is proposing a trade...
          </div>
        ) : (
          <div
            style={{
              color: "#7C878E",
              fontSize: 13,
              fontWeight: 600,
              textAlign: "center",
            }}
          >
            Select properties and money for each side of the trade.
          </div>
        )}

        <div
          style={{
            display: "flex",
            justifyContent: "flex-end",
            gap: 10,
            flexWrap: "wrap",
          }}
        >
          {pendingTrade ? (
            isTradeTarget ? (
              <>
                <button
                  type="button"
                  onClick={handleReject}
                  disabled={rejectTradeMutation.isPending}
                  style={{
                    padding: "10px 14px",
                    backgroundColor: "#D0D3D4",
                    color: "#000000",
                    fontWeight: 700,
                    cursor: rejectTradeMutation.isPending ? "not-allowed" : "pointer",
                  }}
                >
                  Reject
                </button>
                <button
                  type="button"
                  onClick={handleAccept}
                  disabled={acceptTradeMutation.isPending}
                  style={{
                    padding: "10px 14px",
                    backgroundColor: "#F76902",
                    color: "#FFFFFF",
                    fontWeight: 700,
                    cursor: acceptTradeMutation.isPending ? "not-allowed" : "pointer",
                  }}
                >
                  Accept
                </button>
              </>
            ) : null
          ) : isDraftComposer ? (
            <>
              <button
                type="button"
                onClick={handlePropose}
                disabled={!isCurrentPlayerTurn || proposeTradeMutation.isPending}
                style={{
                  padding: "10px 14px",
                  backgroundColor: !isCurrentPlayerTurn || proposeTradeMutation.isPending ? "#D0D3D4" : "#F76902",
                  color: "#FFFFFF",
                  fontWeight: 700,
                  cursor: !isCurrentPlayerTurn || proposeTradeMutation.isPending ? "not-allowed" : "pointer",
                }}
              >
                Propose
              </button>
            </>
          ) : null}
        </div>
      </div>
    </div>
  )
}
