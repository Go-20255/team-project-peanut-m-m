"use client"

import { useState, useEffect } from "react"
import { getTokenIcon, getTokenName } from "@/utils/tokens"
import { Player, GameState } from "@/types"
import { useMovePlayer, useRollDice } from "@/hooks/useGameAPI"
import { storage } from "@/utils"
import { useEndTurn } from "@/hooks/playerHooks"

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
  const [error, setError] = useState<string | null>(null)
  const [joinCode, setJoinCode] = useState<string>("")

  const isCurrentPlayerTurn = currentPlayerTurnId?.toString() === playerId
  const isGameStarted = (gameState?.current_turn ?? -1) >= 0

  const currentPlayer = players.find((p) => p.id === currentPlayerTurnId)
  const rollMutation = useRollDice()
  const endTurnMutation = useEndTurn()
  const movePlayerMutation = useMovePlayer()

  useEffect(() => {
    const code = storage.getGameCode()
    if (code) setJoinCode(code)
  }, [])

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
    rollMutation.mutate(
      { playerId: playerId, sessionId: sessionId },
      {
        onSuccess: (res) => {
          setDiceRoll(res)
          console.log("Dice roll successful:", res)
        },
        onError: (err) => {
          var errorText = err.message
          setError(`Failed to roll: ${errorText}`)
          console.error("Roll dice error:", errorText)
        },
      },
    )
  }

  const handleMove = async () => {
    setError(null)
    movePlayerMutation.mutate(
      { playerId: playerId, sessionId: sessionId },
      {
        onSuccess: (res) => {
          setDiceRoll(null)
          console.log("Move successful")
        },
        onError: (err) => {
          var errorText = err.message
          setError(`Failed to move: ${errorText}`)
          console.error("Move error:", errorText)
        },
      },
    )
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
          {isGameStarted ? "CURRENT TURN" : "LOBBY"}
        </div>
        <div className="text-lg font-bold mb-2" style={{ color: isGameStarted && isCurrentPlayerTurn ? "#00AA00" : "#F76902" }}>
          {isGameStarted ? (currentPlayer ? currentPlayer.name : "Waiting...") : "Waiting"}
        </div>

        {isGameStarted && isCurrentPlayerTurn && (
          <div className="text-xs mb-3" style={{ color: "#00AA00" }}>
            Your Turn
          </div>
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
        ) : diceRoll ? (
          <>
            <div className="text-sm font-bold mb-3" style={{ color: "#F76902" }}>
              🎲 Last Roll: {diceRoll.die_one} + {diceRoll.die_two} = {diceRoll.total}
            </div>
            <button
              onClick={handleMove}
              disabled={movePlayerMutation.isPending || !isCurrentPlayerTurn}
              className="w-full py-2 px-3 rounded font-bold text-white"
              style={{
                backgroundColor: movePlayerMutation.isPending || !isCurrentPlayerTurn ? "#ccc" : "#F76902",
                cursor: movePlayerMutation.isPending || !isCurrentPlayerTurn ? "not-allowed" : "pointer",
              }}
              title={!isCurrentPlayerTurn ? "Not your turn" : "Move your piece"}
            >
              {movePlayerMutation.isPending ? "Moving..." : "Move"}
            </button>
          </>
        ) : (
          <button
            onClick={handleRollDice}
            disabled={rollMutation.isPending || !isCurrentPlayerTurn}
            className="w-full py-2 px-3 rounded font-bold text-white"
            style={{
              backgroundColor: rollMutation.isPending || !isCurrentPlayerTurn ? "#ccc" : "#F76902",
              cursor: rollMutation.isPending || !isCurrentPlayerTurn ? "not-allowed" : "pointer",
            }}
            title={!isCurrentPlayerTurn ? "Not your turn - wait for your turn to roll" : "Roll the dice"}
          >
            {rollMutation.isPending ? "Rolling..." : "Roll Dice"}
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

                <div
                  className="text-xs mb-2"
                  style={{
                    color: isPlayerTurn ? "#F57F17" : isCurrentPlayer ? "#000000" : "#7C878E",
                  }}
                >
                  {player.ready_up_status ? "Ready" : "Not Ready"}
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
