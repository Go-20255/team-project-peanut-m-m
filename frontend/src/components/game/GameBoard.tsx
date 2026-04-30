"use client"

import { useState, useEffect } from "react"
import { storage } from "@/utils/storage"
import { getTokenIcon, getTokenName } from "@/utils/tokens"
import { Player, GameState } from "@/types"

interface GameBoardProps {
  sessionId: string
  playerId: string
  playerName: string
  players: Player[]
  currentPlayerTurnId?: number | null
  gameState?: GameState | null
}

const BOARD_SPACES = [
  "Go",
  "Med Ave",
  "Chest",
  "Baltic",
  "Tax",
  "R.R.",
  "Oriental",
  "Chance",
  "Vermont",
  "Connect",
  "Jail",
  "St. Ch",
  "Elec",
  "States",
  "Virginia",
  "R.R.",
  "St. Ja",
  "Chest",
  "Tenn",
  "NY Ave",
  "Park",
  "Ky Ave",
  "Chance",
  "Indiana",
  "Illinois",
  "R.R.",
  "Atlantic",
  "Ventnor",
  "Water",
  "Marvin",
  "Go to J",
  "Pacific",
  "N.C.A",
  "Chest",
  "Penn",
  "R.R.",
  "Chance",
  "Park Pl",
  "Tax",
  "Boardwalk",
]

export default function GameBoard({
  sessionId,
  playerId,
  playerName,
  players,
  currentPlayerTurnId,
  gameState,
}: GameBoardProps) {
  const [joinCode, setJoinCode] = useState<string>("")

  // Find current player's turn info
  const currentPlayer = players.find((p) => p.id === currentPlayerTurnId)
  const isCurrentPlayerTurn = currentPlayerTurnId?.toString() === playerId

  useEffect(() => {
    const code = storage.getGameCode()
    if (code) {
      setJoinCode(code)
    }
  }, [])

  const getPlayerPositions = () => {
    const positions: { [key: number]: Player[] } = {}
    players.forEach((player) => {
      if (!positions[player.position]) {
        positions[player.position] = []
      }
      positions[player.position].push(player)
    })
    return positions
  }

  const playerPositions = getPlayerPositions()

  const SPACE_SIZE = 50
  const CORNER_SIZE = 50
  const BOARD_DIM = 11

  return (
    <div className="w-full h-full flex flex-col p-3" style={{ backgroundColor: "#FFFFFF" }}>
      {/* Top Bar with Game Code and Turn Info */}
      <div className="flex justify-between items-center mb-3">
        <h2 className="text-2xl font-bold" style={{ color: "#F76902" }}>
          Board
        </h2>

        {/* Game Code */}
        <div className="text-center border-2 px-3 py-1" style={{ borderColor: "#D0D3D4" }}>
          <div className="text-xs" style={{ color: "#7C878E" }}>
            Code
          </div>
          <div className="text-lg font-bold" style={{ color: "#F76902" }}>
            {joinCode || "..."}
          </div>
        </div>
      </div>

      {/* Game Board */}
      <div className="flex-1 flex items-center justify-center overflow-auto">
        <div style={{ border: "3px solid #000000" }}>
          <table cellSpacing="0" cellPadding="0">
            <tbody>
              {Array.from({ length: BOARD_DIM }).map((_, row) => (
                <tr key={row}>
                  {Array.from({ length: BOARD_DIM }).map((_, col) => {
                    let spaceIdx = -1

                    // Bottom row (left to right): 0-10
                    if (row === BOARD_DIM - 1) {
                      spaceIdx = col
                    }
                    // Top row (right to left): 30-20
                    else if (row === 0) {
                      spaceIdx = 30 - (BOARD_DIM - 1 - col)
                    }
                    // Left column (bottom to top): 39-31
                    else if (col === 0) {
                      spaceIdx = 40 - row
                    }
                    // Right column (top to bottom): 11-19
                    else if (col === BOARD_DIM - 1) {
                      spaceIdx = 10 + row
                    }
                    // Center (empty space)
                    else {
                      return (
                        <td
                          key={`${row}-${col}`}
                          style={{
                            width: `${SPACE_SIZE}px`,
                            height: `${SPACE_SIZE}px`,
                            backgroundColor: "#F0F0F0",
                          }}
                        />
                      )
                    }

                    const playersOnSpace = playerPositions[spaceIdx] || []
                    const isCorner = spaceIdx === 0 || spaceIdx === 10 || spaceIdx === 20 || spaceIdx === 30

                    return (
                      <td
                        key={`${row}-${col}`}
                        style={{
                          width: `${SPACE_SIZE}px`,
                          height: `${SPACE_SIZE}px`,
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
                        title={BOARD_SPACES[spaceIdx]}
                      >
                        <div style={{ lineHeight: 1.2 }}>
                          <span>{BOARD_SPACES[spaceIdx]}</span>
                        </div>
                        {/* Player tokens */}
                        <div
                          style={{
                            position: "absolute",
                            bottom: "1px",
                            left: "50%",
                            transform: "translateX(-50%)",
                            display: "flex",
                            gap: "1px",
                            flexWrap: "wrap",
                            justifyContent: "center",
                            width: "100%",
                          }}
                        >
                          {playersOnSpace.map((player, idx) => (
                            <img
                              key={`${player.id}-token`}
                              src={getTokenIcon(player.piece_token)}
                              alt={getTokenName(player.piece_token)}
                              style={{
                                width: "12px",
                                height: "12px",
                                border: player.id === currentPlayerTurnId ? "2px solid #FFD700" : "1px solid #000",
                                cursor: "pointer",
                                boxShadow: player.id === currentPlayerTurnId ? "0 0 4px #FFD700" : "none",
                              }}
                              title={`${player.name} (${getTokenName(player.piece_token)})${player.id === currentPlayerTurnId ? " - TURN" : ""}`}
                            />
                          ))}
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
