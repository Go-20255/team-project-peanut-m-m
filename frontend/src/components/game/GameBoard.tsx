"use client"

import { type MouseEvent, type WheelEvent, useEffect, useMemo, useRef, useState } from "react"
import { getTokenIcon } from "@/utils/tokens"
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
const BOARD_SIZE = 1850

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

type ViewportState = {
  zoom: number
  x: number
  y: number
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

function clamp(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value))
}

function getContainSize(
  sourceWidth: number,
  sourceHeight: number,
  maxWidth: number,
  maxHeight: number,
) {
  const sourceRatio = sourceWidth / sourceHeight
  const boxRatio = maxWidth / maxHeight

  if (sourceRatio > boxRatio) {
    return {
      width: maxWidth,
      height: maxWidth / sourceRatio,
    }
  }

  return {
    width: maxHeight * sourceRatio,
    height: maxHeight,
  }
}

export default function GameBoard({
  currentPlayerTurnId,
  gameState,
}: GameBoardProps) {
  const containerRef = useRef<HTMLDivElement | null>(null)
  const canvasRef = useRef<HTMLCanvasElement | null>(null)
  const tileImagesRef = useRef<Record<number, HTMLImageElement>>({})
  const tokenImagesRef = useRef<Record<number, HTMLImageElement>>({})
  const rotationRef = useRef(0)
  const dragRef = useRef<{ active: boolean; x: number; y: number }>({
    active: false,
    x: 0,
    y: 0,
  })

  const [viewport, setViewport] = useState<ViewportState>({
    zoom: 1,
    x: 0,
    y: 0,
    rotation: 0,
  })
  const [rotationTarget, setRotationTarget] = useState(0)
  const [isDragging, setIsDragging] = useState(false)
  const [size, setSize] = useState({ width: 0, height: 0 })
  const [imageVersion, setImageVersion] = useState(0)

  const boardTiles = useMemo(
    () => Array.from({ length: 40 }, (_, index) => getTilePlacement(index)),
    [],
  )

  const playerPositions = useMemo(() => {
    const positions: Record<number, Player[]> = {}
    gameState.players.forEach((playerInfo) => {
      const position = playerInfo.player.position
      if (!positions[position]) positions[position] = []
      positions[position].push(playerInfo.player)
    })
    return positions
  }, [gameState.players])

  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    const updateSize = () => {
      setSize({
        width: container.clientWidth,
        height: container.clientHeight,
      })
    }

    updateSize()

    const observer = new ResizeObserver(() => updateSize())
    observer.observe(container)

    return () => observer.disconnect()
  }, [])

  useEffect(() => {
    rotationRef.current = viewport.rotation
  }, [viewport.rotation])

  useEffect(() => {
    if (rotationRef.current === rotationTarget) return

    const startRotation = rotationRef.current
    const delta = rotationTarget - startRotation
    const duration = 220
    let animationFrame = 0
    let startTime = 0

    const tick = (time: number) => {
      if (!startTime) startTime = time

      const progress = Math.min((time - startTime) / duration, 1)
      const eased = 1 - Math.pow(1 - progress, 3)
      const nextRotation = startRotation + delta * eased

      rotationRef.current = nextRotation
      setViewport((value) => ({
        ...value,
        rotation: nextRotation,
      }))

      if (progress < 1) {
        animationFrame = window.requestAnimationFrame(tick)
      }
    }

    animationFrame = window.requestAnimationFrame(tick)

    return () => window.cancelAnimationFrame(animationFrame)
  }, [rotationTarget])

  useEffect(() => {
    let cancelled = false

    const tileEntries = Array.from({ length: 40 }, (_, index) => index)
    let loadedCount = 0
    const targetCount = tileEntries.length + 4

    const finishLoad = () => {
      loadedCount += 1
      if (!cancelled && loadedCount === targetCount) {
        setImageVersion((value) => value + 1)
      }
    }

    tileEntries.forEach((index) => {
      const image = new Image()
      image.onload = finishLoad
      image.src = `/assets/img/tiles/${index}.png`
      tileImagesRef.current[index] = image
    })

    ;[0, 1, 2, 3].forEach((token) => {
      const image = new Image()
      image.onload = finishLoad
      image.src = getTokenIcon(token)
      tokenImagesRef.current[token] = image
    })

    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas || size.width === 0 || size.height === 0) return

    const ratio = window.devicePixelRatio || 1
    canvas.width = Math.floor(size.width * ratio)
    canvas.height = Math.floor(size.height * ratio)
    canvas.style.width = `${size.width}px`
    canvas.style.height = `${size.height}px`

    const ctx = canvas.getContext("2d")
    if (!ctx) return

    ctx.setTransform(1, 0, 0, 1, 0, 0)
    ctx.scale(ratio, ratio)
    ctx.clearRect(0, 0, size.width, size.height)
    ctx.fillStyle = "#FFFFFF"
    ctx.fillRect(0, 0, size.width, size.height)

    const baseZoom = size.width / BOARD_SIZE

    ctx.save()
    ctx.translate(size.width / 2 + viewport.x, size.height / 2 + viewport.y)
    ctx.rotate((viewport.rotation * Math.PI) / 180)
    ctx.scale(baseZoom * viewport.zoom, baseZoom * viewport.zoom)
    ctx.translate(-BOARD_SIZE / 2, -BOARD_SIZE / 2)

    boardTiles.forEach((placement) => {
      const image = tileImagesRef.current[placement.index]
      if (!image) return

      const x = (placement.colStart / BOARD_UNITS) * BOARD_SIZE
      const y = (placement.rowStart / BOARD_UNITS) * BOARD_SIZE
      const width = (placement.colSpan / BOARD_UNITS) * BOARD_SIZE
      const height = (placement.rowSpan / BOARD_UNITS) * BOARD_SIZE
      const isCornerTile =
        placement.rowSpan === CORNER_UNITS && placement.colSpan === CORNER_UNITS
      const isSideTile = !isCornerTile && (placement.side === "left" || placement.side === "right")

      ctx.save()
      ctx.beginPath()
      ctx.rect(x, y, width, height)
      ctx.clip()
      ctx.translate(x + width / 2, y + height / 2)
      ctx.rotate((placement.rotation * Math.PI) / 180)

      if (isSideTile) {
        const drawSize = getContainSize(image.width, image.height, height, width)
        ctx.drawImage(
          image,
          -drawSize.width / 2,
          -drawSize.height / 2,
          drawSize.width,
          drawSize.height,
        )
      } else {
        ctx.drawImage(image, -width / 2, -height / 2, width, height)
      }
      ctx.restore()

      const playersOnTile = playerPositions[placement.index] || []
      if (playersOnTile.length === 0) return

      const columns = playersOnTile.length > 2 ? 2 : playersOnTile.length
      const rows = Math.ceil(playersOnTile.length / columns)
      const tokenSize = isCornerTile ? 58 : 42
      const gap = 8
      const totalWidth = columns * tokenSize + (columns - 1) * gap
      const totalHeight = rows * tokenSize + (rows - 1) * gap
      const startX = x + width / 2 - totalWidth / 2
      const startY = y + height / 2 - totalHeight / 2

      playersOnTile.forEach((player, index) => {
        const tokenImage = tokenImagesRef.current[player.piece_token]
        if (!tokenImage) return

        const column = index % columns
        const row = Math.floor(index / columns)
        const tokenX = startX + column * (tokenSize + gap)
        const tokenY = startY + row * (tokenSize + gap)
        const isTurn = player.id === currentPlayerTurnId

        ctx.save()
        ctx.fillStyle = "#FFFFFF"
        ctx.beginPath()
        ctx.arc(tokenX + tokenSize / 2, tokenY + tokenSize / 2, tokenSize / 2, 0, Math.PI * 2)
        ctx.fill()

        if (isTurn) {
          ctx.strokeStyle = "#F76902"
          ctx.lineWidth = 4
          ctx.stroke()
        }

        ctx.drawImage(tokenImage, tokenX + tokenSize * 0.14, tokenY + tokenSize * 0.14, tokenSize * 0.72, tokenSize * 0.72)
        ctx.restore()
      })
    })

    ctx.restore()
  }, [boardTiles, currentPlayerTurnId, imageVersion, playerPositions, size, viewport])

  const onWheel = (event: WheelEvent<HTMLDivElement>) => {
    event.preventDefault()
    const rect = containerRef.current?.getBoundingClientRect()
    if (!rect) return

    const cursorX = event.clientX - rect.left - rect.width / 2
    const cursorY = event.clientY - rect.top - rect.height / 2

    setViewport((value) => {
      const newZoom = clamp(Number((value.zoom - event.deltaY * 0.001).toFixed(3)), 0.5, 4)
      const zoomRatio = newZoom / value.zoom

      return {
        ...value,
        zoom: newZoom,
        x: cursorX - zoomRatio * (cursorX - value.x),
        y: cursorY - zoomRatio * (cursorY - value.y),
      }
    })
  }

  const onMouseDown = (event: MouseEvent<HTMLDivElement>) => {
    dragRef.current = {
      active: true,
      x: event.clientX,
      y: event.clientY,
    }
    setIsDragging(true)
  }

  const onMouseMove = (event: MouseEvent<HTMLDivElement>) => {
    if (!dragRef.current.active) return

    const dx = event.clientX - dragRef.current.x
    const dy = event.clientY - dragRef.current.y

    dragRef.current = {
      active: true,
      x: event.clientX,
      y: event.clientY,
    }

    setViewport((value) => ({
      ...value,
      x: value.x + dx,
      y: value.y + dy,
    }))
  }

  const stopDrag = () => {
    dragRef.current.active = false
    setIsDragging(false)
  }

  return (
    <div
      ref={containerRef}
      className="w-full h-full"
      onWheel={onWheel}
      onMouseDown={onMouseDown}
      onMouseMove={onMouseMove}
      onMouseUp={stopDrag}
      onMouseLeave={stopDrag}
      style={{
        position: "relative",
        overflow: "hidden",
        backgroundColor: "#FFFFFF",
        minHeight: "100vh",
        cursor: isDragging ? "grabbing" : "grab",
      }}
    >
      <div
        style={{
          position: "absolute",
          left: 16,
          top: 16,
          zIndex: 20,
          padding: 8,
          backgroundColor: "#FFFFFF",
          border: "1px solid #000000",
        }}
      >
        <button
          type="button"
          onMouseDown={(event) => event.stopPropagation()}
          onClick={() => setRotationTarget((value) => value + 90)}
          style={{
            width: 40,
            height: 40,
            backgroundColor: "#FFFFFF",
            border: "1px solid #000000",
            fontSize: 20,
          }}
        >
          ↻
        </button>
      </div>

      <canvas ref={canvasRef} />
    </div>
  )
}
