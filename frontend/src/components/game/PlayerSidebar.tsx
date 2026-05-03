"use client"

import { useState, useEffect } from "react"
import { getTokenIcon, getTokenName } from "@/utils/tokens"
import { Player, GameState } from "@/types"
import { useEndTurn, useMovePlayer, useRollDice } from "@/hooks/useGameAPI"

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
  const [diceRoll, setDiceRoll] = useState<any>(null)
  const [isRolling, setIsRolling] = useState(false)
  const [isMoving, setIsMoving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const isCurrentPlayerTurn = currentPlayerTurnId?.toString() === playerId

  const currentPlayer = players.find((p) => p.id === currentPlayerTurnId)
  const rollMutation = useRollDice()
  const endTurnMutation = useEndTurn()
  const movePlayerMutation = useMovePlayer()

  const handleEndTurn = () => {
    if (!isCurrentPlayerTurn) {
      setError("It's not your turn!")
      return
    }
    setError(null)
    endTurnMutation.mutate()
  }

  const handleRollDice = async () => {
    if (!isCurrentPlayerTurn) {
      setError("It's not your turn!")
      return
    }

    setError(null)
    setIsRolling(true)
    rollMutation.mutate({playerId: playerId, sessionId: sessionId}, {
      onSuccess: (res) => {
        setDiceRoll(res)
        console.log("Dice roll successful:", res)
      },
      onError: (err) => {
        var errorText = err.message
        setError(`Failed to roll: ${errorText}`)
        console.error("Roll dice error:", errorText)
      }
    })
    setIsRolling(false)
  }

  const handleMove = async () => {
    setError(null)
    setIsMoving(true)
    movePlayerMutation.mutate({playerId: playerId, sessionId: sessionId}, {
      onSuccess: (res) => {
      setDiceRoll(null)
      console.log("Move successful")
      },
      onError: (err) => {
        var errorText = err.message
        setError(`Failed to move: ${errorText}`)
        console.error("Move error:", errorText)
      }
    })
    setIsMoving(false)
  }

  return (
    <div className="w-full h-full flex flex-col p-4 overflow-y-auto" style={{ backgroundColor: "#FFFFFF" }}>
      <h3 className="text-xl font-bold mb-4" style={{ color: "#F76902" }}>
        Players
      </h3>

      {/* Current Turn Info */}
      <div
        className="mb-4 p-3 border-2"
        style={{
          borderColor: isCurrentPlayerTurn ? "#00AA00" : "#D0D3D4",
          backgroundColor: isCurrentPlayerTurn ? "#E8F5E9" : "#F9F9F9",
        }}
      >
        <div className="text-xs font-bold mb-2" style={{ color: "#7C878E" }}>
          CURRENT TURN
        </div>
        <div className="text-lg font-bold mb-2" style={{ color: isCurrentPlayerTurn ? "#00AA00" : "#F76902" }}>
          {currentPlayer ? currentPlayer.name : "Waiting..."}
        </div>

        {isCurrentPlayerTurn && (
          <div className="text-xs mb-3" style={{ color: "#00AA00" }}>
            ✅ It's your turn to play!
          </div>
        )}

        {!isCurrentPlayerTurn && currentPlayer && (
          <div className="text-xs mb-3" style={{ color: "#7C878E" }}>
            Waiting for {currentPlayer.name} to play...
          </div>
        )}
      </div>

      {/* Actions Section */}
      <div className="mb-4 p-3 border-2 flex flex-col gap-4" style={{ borderColor: "#D0D3D4" }}>
        {diceRoll ? (
          <>
            <div className="text-sm font-bold mb-3" style={{ color: "#F76902" }}>
              🎲 Last Roll: {diceRoll.die_one} + {diceRoll.die_two} = {diceRoll.total}
            </div>
            <button
              onClick={handleMove}
              disabled={isMoving || !isCurrentPlayerTurn}
              className="w-full py-2 px-3 rounded font-bold text-white"
              style={{
                backgroundColor: isMoving || !isCurrentPlayerTurn ? "#ccc" : "#F76902",
                cursor: isMoving || !isCurrentPlayerTurn ? "not-allowed" : "pointer",
              }}
              title={!isCurrentPlayerTurn ? "Not your turn" : "Move your piece"}
            >
              {isMoving ? "Moving..." : "Move"}
            </button>
          </>
        ) : (
          <button
            onClick={handleRollDice}
            disabled={isRolling || !isCurrentPlayerTurn}
            className="w-full py-2 px-3 rounded font-bold text-white"
            style={{
              backgroundColor: isRolling || !isCurrentPlayerTurn ? "#ccc" : "#F76902",
              cursor: isRolling || !isCurrentPlayerTurn ? "not-allowed" : "pointer",
            }}
            title={!isCurrentPlayerTurn ? "Not your turn - wait for your turn to roll" : "Roll the dice"}
          >
            {isRolling ? "Rolling..." : !isCurrentPlayerTurn ? "Waiting..." : "Roll Dice"}
          </button>
        )}
        <>
          <button
            onClick={handleEndTurn}
            disabled={!isCurrentPlayerTurn}
            className="w-full py-2 px-3 rounded font-bold text-white"
            style={{
              backgroundColor: !isCurrentPlayerTurn ? "#ccc" : "#F76902",
              cursor: !isCurrentPlayerTurn ? "not-allowed" : "pointer",
            }}
            title={!isCurrentPlayerTurn ? "Not your turn" : "End Turn"}
          >
            End Turn
          </button>
        </>

        {/* Error message */}
        {error && (
          <div className="text-xs mt-2 p-2 rounded" style={{ color: "#D32F2F", backgroundColor: "#FFEBEE" }}>
            {error}
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
                </div>

                {/* Position */}
                <div
                  className="text-sm mb-2"
                  style={{
                    color: isPlayerTurn ? "#F57F17" : isCurrentPlayer ? "#000000" : "#7C878E",
                  }}
                >
                  Position: {player.position}
                </div>

                {/* Money */}
                <div
                  className="text-sm mb-2"
                  style={{
                    color: isPlayerTurn ? "#F57F17" : isCurrentPlayer ? "#000000" : "#7C878E",
                  }}
                >
                  Money: ${player.money.toLocaleString()}
                </div>

                {/* Properties placeholder */}
                <div
                  className="text-xs mt-2 p-2 border-t-2"
                  style={{
                    borderColor: isPlayerTurn ? "#FFD700" : isCurrentPlayer ? "#000000" : "#D0D3D4",
                    color: isPlayerTurn ? "#F57F17" : isCurrentPlayer ? "#000000" : "#7C878E",
                  }}
                >
                  Properties: None
                </div>
              </div>
            )
          })
        )}
      </div>
    </div>
  )
}
