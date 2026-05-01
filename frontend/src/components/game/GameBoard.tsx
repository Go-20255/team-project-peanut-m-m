"use client"

import { useEffect, useMemo, useState } from "react"
import { storage } from "@/utils/storage"
import { getTokenIcon, getTokenName } from "@/utils/tokens"
import { Player, GameState } from "@/types"

interface GameBoardProps {
  sessionId: string
  playerId: string
  playerName: string
  currentPlayerTurnId?: number | null
  gameState: GameState
}

const SPACE_SIZE = 50
const BOARD_DIM = 11

export default function GameBoard({
  currentPlayerTurnId,
  gameState,
}: GameBoardProps) {
  const [joinCode, setJoinCode] = useState<string>("")

  useEffect(() => {
    const code = storage.getGameCode()
    if (code) setJoinCode(code)
  }, [])

  // Tiles indexed by board position (0-39)
  const tilesByIndex = useMemo(() => {
    const map: Record<number, (typeof gameState.tiles)[number]> = {}
    gameState.tiles.forEach((t) => {
      map[t.id] = t
    })
    return map
  }, [gameState.tiles])

  // Position -> players on that tile
  const playerPositions = useMemo(() => {
    const positions: Record<number, Player[]> = {}
    gameState.players.forEach((pi) => {
      const pos = pi.player.position
      if (!positions[pos]) positions[pos] = []
      positions[pos].push(pi.player)
    })
    return positions
  }, [gameState.players])

  const getSpaceIdx = (row: number, col: number): number => {
    // Left column: bottom to top, 0 -> 10
    if (col === 0) return BOARD_DIM - 1 - row
    // Top row: left to right, 10 -> 20
    if (row === 0) return 10 + col
    // Right column: top to bottom, 20 -> 30
    if (col === BOARD_DIM - 1) return 20 + row
    // Bottom row: right to left, 30 -> 39
    if (row === BOARD_DIM - 1) return 30 + (BOARD_DIM - 1 - col)
    return -1
  }

  return (
    <div
      className="w-full h-full flex flex-col p-3"
      style={{ backgroundColor: "#FFFFFF" }}
    >
      <div className="flex justify-between items-center mb-3">
        <h2 className="text-2xl font-bold" style={{ color: "#F76902" }}>
          Board
        </h2>
        <div
          className="text-center border-2 px-3 py-1"
          style={{ borderColor: "#D0D3D4" }}
        >
          <div className="text-xs" style={{ color: "#7C878E" }}>
            Code
          </div>
          <div className="text-lg font-bold" style={{ color: "#F76902" }}>
            {joinCode || "..."}
          </div>
        </div>
      </div>

      <div className="flex-1 flex items-center justify-center overflow-auto">
        <div style={{ border: "3px solid #000000" }}>
          <table cellSpacing="0" cellPadding="0">
            <tbody>
              {Array.from({ length: BOARD_DIM }).map((_, row) => (
                <tr key={row}>
                  {Array.from({ length: BOARD_DIM }).map((_, col) => {
                    const spaceIdx = getSpaceIdx(row, col)

                    if (spaceIdx === -1) {
                      return (
                        <td
                          key={`${row}-${col}`}
                          style={{
                            width: SPACE_SIZE,
                            height: SPACE_SIZE,
                            backgroundColor: "#F0F0F0",
                          }}
                        />
                      )
                    }

                    const tile = tilesByIndex[spaceIdx]
                    const playersOnSpace = playerPositions[spaceIdx] || []

                    return (
                      <td
                        key={`${row}-${col}`}
                        style={{
                          width: SPACE_SIZE,
                          height: SPACE_SIZE,
                          border: "1px solid #000",
                          backgroundColor: "#FFFFFF",
                          color: "#000000",
                          fontSize: "10px",
                          fontWeight: "bold",
                          padding: "2px",
                          textAlign: "center",
                          verticalAlign: "middle",
                          cursor: "pointer",
                          overflow: "hidden",
                          position: "relative",
                        }}
                        title={tile?.name}
                      >
                        <div style={{ lineHeight: 1.2 }}>
                          <span>{tile?.name}</span>
                        </div>
                        <div
                          style={{
                            position: "absolute",
                            bottom: 1,
                            left: "50%",
                            transform: "translateX(-50%)",
                            display: "flex",
                            gap: 1,
                            flexWrap: "wrap",
                            justifyContent: "center",
                            width: "100%",
                          }}
                        >
                          {playersOnSpace.map((player) => {
                            const isTurn = player.id === currentPlayerTurnId
                            return (
                              <img
                                key={`${player.id}-token`}
                                src={getTokenIcon(player.piece_token)}
                                alt={getTokenName(player.piece_token)}
                                style={{
                                  width: 12,
                                  height: 12,
                                  border: isTurn
                                    ? "2px solid #FFD700"
                                    : "1px solid #000",
                                  cursor: "pointer",
                                  boxShadow: isTurn
                                    ? "0 0 4px #FFD700"
                                    : "none",
                                }}
                                title={`${player.name} (${getTokenName(
                                  player.piece_token,
                                )})${isTurn ? " - TURN" : ""}`}
                              />
                            )
                          })}
                        </div>
                      </td>
                    )
                  })}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
