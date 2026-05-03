"use client"

import { type CSSProperties, useMemo } from "react"
import { getTokenIcon, getTokenName } from "@/utils/tokens"
import { GameState, Player } from "@/types"

interface GameBoardProps {
  sessionId: string
  playerId: string
  playerName: string
  currentPlayerTurnId?: number | null
  gameState: GameState
}

const BOARD_UNITS = 37
const CORNER_UNITS = 5
const EDGE_UNITS = 3

type BoardSide = "bottom" | "left" | "top" | "right"

type TilePlacement = {
  index: number
  side: BoardSide
  rowStart: number
  colStart: number
  rowSpan: number
  colSpan: number
  rotation: number
}

function getTilePlacement(index: number): TilePlacement {
  if (index === 0) {
    return {
      index,
      side: "bottom",
      rowStart: BOARD_UNITS - CORNER_UNITS,
      colStart: BOARD_UNITS - CORNER_UNITS,
      rowSpan: CORNER_UNITS,
      colSpan: CORNER_UNITS,
      rotation: 0,
    }
  }

  if (index > 0 && index < 10) {
    return {
      index,
      side: "bottom",
      rowStart: BOARD_UNITS - CORNER_UNITS,
      colStart: BOARD_UNITS - CORNER_UNITS - EDGE_UNITS * index,
      rowSpan: CORNER_UNITS,
      colSpan: EDGE_UNITS,
      rotation: 0,
    }
  }

  if (index === 10) {
    return {
      index,
      side: "left",
      rowStart: BOARD_UNITS - CORNER_UNITS,
      colStart: 0,
      rowSpan: CORNER_UNITS,
      colSpan: CORNER_UNITS,
      rotation: 0,
    }
  }

  if (index > 10 && index < 20) {
    return {
      index,
      side: "left",
      rowStart: BOARD_UNITS - CORNER_UNITS - EDGE_UNITS * (index - 10),
      colStart: 0,
      rowSpan: EDGE_UNITS,
      colSpan: CORNER_UNITS,
      rotation: 90,
    }
  }

  if (index === 20) {
    return {
      index,
      side: "top",
      rowStart: 0,
      colStart: 0,
      rowSpan: CORNER_UNITS,
      colSpan: CORNER_UNITS,
      rotation: 90,
    }
  }

  if (index > 20 && index < 30) {
    return {
      index,
      side: "top",
      rowStart: 0,
      colStart: CORNER_UNITS + EDGE_UNITS * (index - 21),
      rowSpan: CORNER_UNITS,
      colSpan: EDGE_UNITS,
      rotation: 180,
    }
  }

  if (index === 30) {
    return {
      index,
      side: "right",
      rowStart: 0,
      colStart: BOARD_UNITS - CORNER_UNITS,
      rowSpan: CORNER_UNITS,
      colSpan: CORNER_UNITS,
      rotation: -180,
    }
  }

  return {
    index,
    side: "right",
    rowStart: CORNER_UNITS + EDGE_UNITS * (index - 31),
    colStart: BOARD_UNITS - CORNER_UNITS,
    rowSpan: EDGE_UNITS,
    colSpan: CORNER_UNITS,
    rotation: -90,
  }
}

function getTokenTrayStyle(side: BoardSide): CSSProperties {
  if (side === "bottom") {
    return {
      top: 8,
      left: 8,
      right: 8,
      justifyContent: "flex-start",
    }
  }

  if (side === "top") {
    return {
      bottom: 8,
      left: 8,
      right: 8,
      justifyContent: "flex-end",
    }
  }

  if (side === "left") {
    return {
      top: 8,
      bottom: 8,
      right: 8,
      justifyContent: "center",
      flexDirection: "column",
      alignItems: "flex-end",
    }
  }

  return {
    top: 8,
    bottom: 8,
    left: 8,
    justifyContent: "center",
    flexDirection: "column",
    alignItems: "flex-start",
  }
}

function getTileImageStyle(placement: TilePlacement): CSSProperties {
  const isCornerTile =
    placement.rowSpan === CORNER_UNITS && placement.colSpan === CORNER_UNITS

  if (!isCornerTile && (placement.side === "left" || placement.side === "right")) {
    return {
      position: "absolute",
      top: "50%",
      left: "50%",
      width: "60%",
      height: "166.6667%",
      transform: `translate(-50%, -50%) rotate(${placement.rotation}deg)`,
      transformOrigin: "center",
    }
  }

  return {
    position: "absolute",
    inset: 0,
    width: "100%",
    height: "100%",
    transform: `rotate(${placement.rotation}deg)`,
    transformOrigin: "center",
  }
}

export default function GameBoard({
  currentPlayerTurnId,
  gameState,
}: GameBoardProps) {
  const tilesByIndex = useMemo(() => {
    const map: Record<number, (typeof gameState.tiles)[number]> = {}
    gameState.tiles.forEach((tile) => {
      map[tile.id] = tile
    })
    return map
  }, [gameState.tiles])

  const playerPositions = useMemo(() => {
    const positions: Record<number, Player[]> = {}
    gameState.players.forEach((playerInfo) => {
      const position = playerInfo.player.position
      if (!positions[position]) positions[position] = []
      positions[position].push(playerInfo.player)
    })
    return positions
  }, [gameState.players])

  const boardTiles = useMemo(
    () => Array.from({ length: 40 }, (_, index) => getTilePlacement(index)),
    [],
  )

  return (
    <div
      className="w-full h-full overflow-y-auto overflow-x-hidden"
      style={{
        backgroundColor: "#FFFFFF",
      }}
    >
      <div className="p-6">
        <div
          className="relative"
          style={{
            width: "100%",
            aspectRatio: "1 / 1",
          }}
        >
          <div
            className="grid w-full h-full"
            style={{
              gridTemplateColumns: `repeat(${BOARD_UNITS}, minmax(0, 1fr))`,
              gridTemplateRows: `repeat(${BOARD_UNITS}, minmax(0, 1fr))`,
              backgroundColor: "#FFFFFF",
            }}
          >
            {boardTiles.map((placement) => {
              const tile = tilesByIndex[placement.index]
              const playersOnTile = playerPositions[placement.index] || []

              return (
                <div
                  key={placement.index}
                  title={tile?.name || `Tile ${placement.index}`}
                  style={{
                    gridColumn: `${placement.colStart + 1} / span ${placement.colSpan}`,
                    gridRow: `${placement.rowStart + 1} / span ${placement.rowSpan}`,
                    position: "relative",
                    overflow: "hidden",
                    backgroundColor: "#FFFFFF",
                  }}
                >
                  <img
                    src={`/assets/img/tiles/${placement.index}.png`}
                    alt={tile?.name || `Tile ${placement.index}`}
                    style={getTileImageStyle(placement)}
                  />

                  {playersOnTile.length > 0 ? (
                    <div
                      style={{
                        position: "absolute",
                        display: "flex",
                        gap: 6,
                        flexWrap: "wrap",
                        ...getTokenTrayStyle(placement.side),
                      }}
                    >
                      {playersOnTile.map((player) => {
                        const isTurn = player.id === currentPlayerTurnId
                        return (
                          <div
                            key={`${placement.index}-${player.id}`}
                            title={`${player.name} (${getTokenName(player.piece_token)})${isTurn ? " - TURN" : ""}`}
                            style={{
                              width: placement.rowSpan === CORNER_UNITS && placement.colSpan === CORNER_UNITS ? 28 : 22,
                              height: placement.rowSpan === CORNER_UNITS && placement.colSpan === CORNER_UNITS ? 28 : 22,
                              display: "flex",
                              alignItems: "center",
                              justifyContent: "center",
                              opacity: isTurn ? 1 : 0.9,
                            }}
                          >
                            <img
                              src={getTokenIcon(player.piece_token)}
                              alt={getTokenName(player.piece_token)}
                              style={{
                                width: "72%",
                                height: "72%",
                                objectFit: "contain",
                              }}
                            />
                          </div>
                        )
                      })}
                    </div>
                  ) : null}
                </div>
              )
            })}

            <div
              className="flex flex-col items-center justify-center text-center px-6"
              style={{
                gridColumn: `${CORNER_UNITS + 1} / span ${BOARD_UNITS - CORNER_UNITS * 2}`,
                gridRow: `${CORNER_UNITS + 1} / span ${BOARD_UNITS - CORNER_UNITS * 2}`,
                backgroundColor: "#FFFFFF",
              }}
            />
          </div>
        </div>
      </div>
    </div>
  )
}
