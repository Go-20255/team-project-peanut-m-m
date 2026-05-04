"use client"

import { type MouseEvent, type WheelEvent, useEffect, useMemo, useRef, useState } from "react"
import { getTokenIcon } from "@/utils/tokens"
import { DiceRoll, GameState, Player } from "@/types"
import { useReadyUp } from "@/hooks/playerHooks"
import { useEndTurn } from "@/hooks/playerHooks"
import { useJailRelease, useMovePlayer, useRollDice } from "@/hooks/useGameAPI"

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

type MoveAnimation = {
  playerId: number
  pieceToken: number
  path: { x: number; y: number; index: number }[]
  x: number
  y: number
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

function getContainSize(sourceWidth: number, sourceHeight: number, maxWidth: number, maxHeight: number) {
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

function getTileBounds(placement: TilePlacement) {
  return {
    x: (placement.colStart / BOARD_UNITS) * BOARD_SIZE,
    y: (placement.rowStart / BOARD_UNITS) * BOARD_SIZE,
    width: (placement.colSpan / BOARD_UNITS) * BOARD_SIZE,
    height: (placement.rowSpan / BOARD_UNITS) * BOARD_SIZE,
  }
}

function getTileCenter(index: number) {
  const placement = getTilePlacement(index)
  const bounds = getTileBounds(placement)

  return {
    x: bounds.x + bounds.width / 2,
    y: bounds.y + bounds.height / 2,
    index,
  }
}

function getMovementPath(oldPosition: number, newPosition: number) {
  const path = [getTileCenter(oldPosition)]
  let current = oldPosition

  while (current !== newPosition) {
    current = (current + 1) % 40
    path.push(getTileCenter(current))
  }

  return path
}

function getTokenSlots(count: number, centerX: number, centerY: number, tokenSize: number, gap: number) {
  const columns = count > 2 ? 2 : count
  const rows = Math.ceil(count / columns)
  const totalWidth = columns * tokenSize + (columns - 1) * gap
  const totalHeight = rows * tokenSize + (rows - 1) * gap
  const startX = centerX - totalWidth / 2
  const startY = centerY - totalHeight / 2

  return Array.from({ length: count }, (_, index) => {
    const column = index % columns
    const row = Math.floor(index / columns)

    return {
      x: startX + column * (tokenSize + gap),
      y: startY + row * (tokenSize + gap),
    }
  })
}

function getTileTokenSlots(
  placement: TilePlacement,
  playersOnTile: Player[],
  tokenSize: number,
  gap: number,
) {
  const bounds = getTileBounds(placement)

  if (placement.index !== 10) {
    return getTokenSlots(playersOnTile.length, bounds.x + bounds.width / 2, bounds.y + bounds.height / 2, tokenSize, gap)
  }

  const jailedPlayers = playersOnTile.filter((player) => player.jailed > 0)
  const visitingPlayers = playersOnTile.filter((player) => player.jailed === 0)
  const slotMap = new Map<number, { x: number; y: number }>()

  const visitingSlots = getTokenSlots(
    visitingPlayers.length,
    bounds.x + bounds.width * 0.28,
    bounds.y + bounds.height * 0.78,
    tokenSize,
    gap,
  )
  const jailedSlots = getTokenSlots(
    jailedPlayers.length,
    bounds.x + bounds.width * 0.68,
    bounds.y + bounds.height * 0.34,
    tokenSize,
    gap,
  )

  visitingPlayers.forEach((player, index) => {
    slotMap.set(player.id, visitingSlots[index])
  })
  jailedPlayers.forEach((player, index) => {
    slotMap.set(player.id, jailedSlots[index])
  })

  return playersOnTile.map((player) => slotMap.get(player.id) ?? { x: bounds.x, y: bounds.y })
}

export default function GameBoard({ sessionId, playerId, currentPlayerTurnId, gameState }: GameBoardProps) {
  const containerRef = useRef<HTMLDivElement | null>(null)
  const canvasRef = useRef<HTMLCanvasElement | null>(null)
  const tileImagesRef = useRef<Record<number, HTMLImageElement>>({})
  const tokenImagesRef = useRef<Record<number, HTMLImageElement>>({})
  const centerImagesRef = useRef<Record<string, HTMLImageElement>>({})
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
  const [readyError, setReadyError] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)
  const [rollDisplay, setRollDisplay] = useState<DiceRoll | null>(null)
  const [isRollingAnimation, setIsRollingAnimation] = useState(false)
  const [moveAnimation, setMoveAnimation] = useState<MoveAnimation | null>(null)

  const readyMutation = useReadyUp()
  const rollMutation = useRollDice()
  const moveMutation = useMovePlayer()
  const endTurnMutation = useEndTurn()
  const jailReleaseMutation = useJailRelease()

  const boardTiles = useMemo(() => Array.from({ length: 40 }, (_, index) => getTilePlacement(index)), [])

  const playerPositions = useMemo(() => {
    const positions: Record<number, Player[]> = {}
    gameState.players.forEach((playerInfo) => {
      const position = playerInfo.player.position
      if (!positions[position]) positions[position] = []
      positions[position].push(playerInfo.player)
    })
    return positions
  }, [gameState.players])

  const isGameStarted = gameState.current_turn >= 0
  const readyPlayers = useMemo(
    () => gameState.players.filter((playerInfo) => playerInfo.player.ready_up_status),
    [gameState.players],
  )
  const currentLobbyPlayer = useMemo(
    () => gameState.players.find((playerInfo) => playerInfo.player.id.toString() === playerId) ?? null,
    [gameState.players, playerId],
  )
  const isTurnOrderPhase = gameState.current_turn >= 0 && gameState.players.some((playerInfo) => playerInfo.player.player_order === -1)
  const activePlayer = useMemo(
    () => gameState.players.find((playerInfo) => playerInfo.player.id === currentPlayerTurnId)?.player ?? null,
    [currentPlayerTurnId, gameState.players],
  )
  const currentPlayer = useMemo(
    () => gameState.players.find((playerInfo) => playerInfo.player.id.toString() === playerId)?.player ?? null,
    [gameState.players, playerId],
  )
  const isCurrentPlayerTurn = currentPlayerTurnId?.toString() === playerId
  const currentRoll = gameState.current_roll ?? null
  const lastMove = gameState.last_move ?? null

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
    const targetCount = tileEntries.length + 6

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

    ;[
      ["cchest", "/assets/img/tiles/cchest.jpeg"],
      ["chance", "/assets/img/tiles/chance.jpeg"],
    ].forEach(([key, src]) => {
      const image = new Image()
      image.onload = finishLoad
      image.src = src
      centerImagesRef.current[key] = image
    })

    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    if (!currentRoll) {
      setRollDisplay(null)
      setIsRollingAnimation(false)
      return
    }

    setActionError(null)
    setIsRollingAnimation(true)

    const interval = window.setInterval(() => {
      setRollDisplay({
        ...currentRoll,
        die_one: Math.floor(Math.random() * 6) + 1,
        die_two: Math.floor(Math.random() * 6) + 1,
      })
    }, 90)

    const timeout = window.setTimeout(() => {
      window.clearInterval(interval)
      setRollDisplay(currentRoll)
      setIsRollingAnimation(false)
    }, 1100)

    return () => {
      window.clearInterval(interval)
      window.clearTimeout(timeout)
    }
  }, [currentRoll])

  useEffect(() => {
    if (!lastMove) {
      setMoveAnimation(null)
      return
    }

    const movingPlayer = gameState.players.find((playerInfo) => playerInfo.player.id === lastMove.player_id)?.player
    if (!movingPlayer) {
      setMoveAnimation(null)
      return
    }

    const path = getMovementPath(lastMove.old_position, lastMove.new_position)
    if (path.length < 2) {
      setMoveAnimation(null)
      return
    }

    let frame = 0
    let startTime = 0
    const duration = Math.max(320, (path.length - 1) * 220)

    setMoveAnimation({
      playerId: movingPlayer.id,
      pieceToken: movingPlayer.piece_token,
      path,
      x: path[0].x,
      y: path[0].y,
    })

    const tick = (time: number) => {
      if (!startTime) startTime = time

      const progress = Math.min((time - startTime) / duration, 1)
      const segmentProgress = progress * (path.length - 1)
      const segmentIndex = Math.min(Math.floor(segmentProgress), path.length - 2)
      const localProgress = segmentProgress - segmentIndex
      const from = path[segmentIndex]
      const to = path[segmentIndex + 1]

      setMoveAnimation({
        playerId: movingPlayer.id,
        pieceToken: movingPlayer.piece_token,
        path,
        x: from.x + (to.x - from.x) * localProgress,
        y: from.y + (to.y - from.y) * localProgress,
      })

      if (progress < 1) {
        frame = window.requestAnimationFrame(tick)
        return
      }

      setMoveAnimation(null)
    }

    frame = window.requestAnimationFrame(tick)

    return () => {
      window.cancelAnimationFrame(frame)
    }
  }, [gameState.players, lastMove])

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

      const bounds = getTileBounds(placement)
      const x = bounds.x
      const y = bounds.y
      const width = bounds.width
      const height = bounds.height
      const isCornerTile = placement.rowSpan === CORNER_UNITS && placement.colSpan === CORNER_UNITS
      const isSideTile = !isCornerTile && (placement.side === "left" || placement.side === "right")

      ctx.save()
      ctx.beginPath()
      ctx.rect(x, y, width, height)
      ctx.clip()
      ctx.translate(x + width / 2, y + height / 2)
      ctx.rotate((placement.rotation * Math.PI) / 180)

      if (isSideTile) {
        const drawSize = getContainSize(image.width, image.height, height, width)
        ctx.drawImage(image, -drawSize.width / 2, -drawSize.height / 2, drawSize.width, drawSize.height)
      } else {
        ctx.drawImage(image, -width / 2, -height / 2, width, height)
      }
      ctx.restore()

      const playersOnTile = (playerPositions[placement.index] || []).filter(
        (player) => player.id !== moveAnimation?.playerId,
      )
      if (playersOnTile.length === 0) return

      const tokenSize = isCornerTile ? 58 : 42
      const gap = 8
      const tokenSlots = getTileTokenSlots(placement, playersOnTile, tokenSize, gap)

      playersOnTile.forEach((player, index) => {
        const tokenImage = tokenImagesRef.current[player.piece_token]
        if (!tokenImage) return

        const slot = tokenSlots[index]
        const tokenX = slot.x
        const tokenY = slot.y
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

        ctx.drawImage(
          tokenImage,
          tokenX + tokenSize * 0.14,
          tokenY + tokenSize * 0.14,
          tokenSize * 0.72,
          tokenSize * 0.72,
        )
        ctx.restore()
      })
    })

    if (moveAnimation) {
      const tokenImage = tokenImagesRef.current[moveAnimation.pieceToken]
      if (tokenImage) {
        const tokenSize = 52
        const isTurn = moveAnimation.playerId === currentPlayerTurnId

        ctx.save()
        ctx.fillStyle = "#FFFFFF"
        ctx.beginPath()
        ctx.arc(moveAnimation.x, moveAnimation.y, tokenSize / 2, 0, Math.PI * 2)
        ctx.fill()

        if (isTurn) {
          ctx.strokeStyle = "#F76902"
          ctx.lineWidth = 4
          ctx.stroke()
        }

        ctx.drawImage(
          tokenImage,
          moveAnimation.x - tokenSize * 0.36,
          moveAnimation.y - tokenSize * 0.36,
          tokenSize * 0.72,
          tokenSize * 0.72,
        )
        ctx.restore()
      }
    }

    const drawCenterCard = (
      image: HTMLImageElement | undefined,
      centerX: number,
      centerY: number,
      width: number,
      rotation: number,
    ) => {
      if (!image) return

      const drawSize = getContainSize(image.width, image.height, width, width)

      ctx.save()
      ctx.translate(centerX, centerY)
      ctx.rotate((rotation * Math.PI) / 180)
      ctx.drawImage(
        image,
        -drawSize.width / 2,
        -drawSize.height / 2,
        drawSize.width,
        drawSize.height,
      )
      ctx.restore()
    }

    drawCenterCard(centerImagesRef.current.cchest, BOARD_SIZE * 0.32, BOARD_SIZE * 0.32, BOARD_SIZE * 0.22, 135)
    drawCenterCard(centerImagesRef.current.chance, BOARD_SIZE * 0.68, BOARD_SIZE * 0.68, BOARD_SIZE * 0.22, -45)

    ctx.restore()
  }, [boardTiles, currentPlayerTurnId, imageVersion, moveAnimation, playerPositions, size, viewport])

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

  const toggleReady = () => {
    if (!currentLobbyPlayer || isGameStarted) return

    setReadyError(null)
    readyMutation.mutate(!currentLobbyPlayer.player.ready_up_status, {
      onError: (error: Error) => {
        setReadyError(error.message)
      },
    })
  }

  const handleRoll = () => {
    if (!isCurrentPlayerTurn) return

    setActionError(null)
    rollMutation.mutate(
      {
        playerId,
        sessionId,
      },
      {
        onError: (error: Error) => {
          setActionError(error.message)
        },
      },
    )
  }

  const handleMove = () => {
    if (!isCurrentPlayerTurn) return

    setActionError(null)
    moveMutation.mutate(
      {
        playerId,
        sessionId,
      },
      {
        onError: (error: Error) => {
          setActionError(error.message)
        },
      },
    )
  }

  const handleEndTurn = () => {
    if (!isCurrentPlayerTurn) return

    setActionError(null)
    endTurnMutation.mutate(undefined, {
      onError: (error: Error) => {
        setActionError(error.message)
      },
    })
  }

  const handleJailRelease = (method: string) => {
    if (!isCurrentPlayerTurn) return

    setActionError(null)
    jailReleaseMutation.mutate(method, {
      onError: (error: Error) => {
        setActionError(error.message)
      },
    })
  }

  const rollLabel = isTurnOrderPhase
    ? "Roll for Order"
    : lastMove?.player_id === currentPlayerTurnId && lastMove?.roll_again
      ? "Roll Again"
      : "Roll Dice"

  const showMoveButton =
    !!currentRoll &&
    !isTurnOrderPhase &&
    !currentRoll.sent_to_jail &&
    (currentRoll.jailed === 0 || currentRoll.released_from_jail)

  const showJailEndTurn =
    !!currentRoll &&
    !isTurnOrderPhase &&
    currentRoll.jailed > 0 &&
    !currentRoll.released_from_jail &&
    currentRoll.jailed < 3

  const showJailReleaseButtons =
    !!currentRoll &&
    !isTurnOrderPhase &&
    currentRoll.jailed >= 3 &&
    !currentRoll.released_from_jail

  const showTurnOrderEndTurn = !!currentRoll && isTurnOrderPhase
  const showSentToJailEndTurn = !!currentRoll && !isTurnOrderPhase && currentRoll.sent_to_jail
  const showTurnPanel =
    !isGameStarted ||
    !!readyError ||
    !!actionError ||
    isRollingAnimation ||
    !!currentRoll ||
    (!moveAnimation && !!lastMove && isCurrentPlayerTurn && lastMove.roll_again) ||
    (!moveAnimation && !lastMove && isCurrentPlayerTurn)

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
        height: "100%",
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

      {!isGameStarted ? (
        <div
          style={{
            position: "absolute",
            left: "50%",
            top: "50%",
            transform: "translate(-50%, -50%)",
            zIndex: 15,
            width: "min(360px, 80%)",
            backgroundColor: "#FFFFFF",
            border: "2px solid #D0D3D4",
            padding: 20,
            display: "flex",
            flexDirection: "column",
            gap: 14,
          }}
        >
          <div
            style={{
              color: "#F76902",
              fontSize: 22,
              fontWeight: 700,
              textAlign: "center",
            }}
          >
            Lobby
          </div>

          <div
            style={{
              color: "#7C878E",
              fontSize: 13,
              textAlign: "center",
            }}
          >
            {readyPlayers.length} / {gameState.players.length} ready
          </div>

          <div
            style={{
              display: "flex",
              flexDirection: "column",
              gap: 8,
            }}
          >
            {gameState.players.map((playerInfo) => (
              <div
                key={playerInfo.player.id}
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                  padding: "10px 12px",
                  border: "1px solid #D0D3D4",
                  backgroundColor: playerInfo.player.id.toString() === playerId ? "#FFF3E0" : "#FFFFFF",
                }}
              >
                <div
                  style={{
                    color: "#000000",
                    fontSize: 14,
                    fontWeight: 600,
                  }}
                >
                  {playerInfo.player.name}
                </div>
                <div
                  style={{
                    color: playerInfo.player.ready_up_status ? "#F76902" : "#7C878E",
                    fontSize: 12,
                    fontWeight: 700,
                  }}
                >
                  {playerInfo.player.ready_up_status ? "READY" : "WAITING"}
                </div>
              </div>
            ))}
          </div>

          <button
            type="button"
            onClick={toggleReady}
            disabled={!currentLobbyPlayer || readyMutation.isPending}
            style={{
              width: "100%",
              padding: "12px 14px",
              backgroundColor: readyMutation.isPending ? "#D0D3D4" : "#F76902",
              color: "#FFFFFF",
              fontWeight: 700,
              cursor: readyMutation.isPending ? "not-allowed" : "pointer",
            }}
          >
            {currentLobbyPlayer?.player.ready_up_status ? "Unready" : "Ready Up"}
          </button>

          {readyError ? (
            <div
              style={{
                color: "#D32F2F",
                fontSize: 12,
                textAlign: "center",
              }}
            >
              {readyError}
            </div>
          ) : null}
        </div>
      ) : null}

      {isGameStarted && showTurnPanel ? (
        <div
          style={{
            position: "absolute",
            left: "50%",
            top: "50%",
            transform: "translate(-50%, -50%)",
            zIndex: 15,
            width: "min(360px, 80%)",
            backgroundColor: "#FFFFFF",
            border: "2px solid #D0D3D4",
            padding: 20,
            display: "flex",
            flexDirection: "column",
            gap: 14,
            alignItems: "center",
            textAlign: "center",
          }}
        >
          <div
            style={{
              color: "#F76902",
              fontSize: 20,
              fontWeight: 700,
            }}
          >
            {isTurnOrderPhase ? "Turn Order" : "Turn"}
          </div>

          <div
            style={{
              color: "#7C878E",
              fontSize: 13,
            }}
          >
            {activePlayer ? `${activePlayer.name}` : "Waiting"}
          </div>

          {rollDisplay ? (
            <div
              style={{
                display: "flex",
                gap: 12,
              }}
            >
              {[rollDisplay.die_one, rollDisplay.die_two].map((value, index) => (
                <div
                  key={`${value}-${index}`}
                  style={{
                    width: 64,
                    height: 64,
                    border: "2px solid #D0D3D4",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    fontSize: 28,
                    fontWeight: 700,
                    color: "#F76902",
                    backgroundColor: "#FFFFFF",
                  }}
                >
                  {value}
                </div>
              ))}
            </div>
          ) : null}

          <div
            style={{
              color: "#000000",
              fontSize: 14,
              fontWeight: 600,
            }}
          >
            {isRollingAnimation
              ? "Rolling..."
              : currentRoll
                ? isTurnOrderPhase
                  ? `Total ${currentRoll.total}`
                  : currentRoll.sent_to_jail
                    ? "Go to Jail"
                    : currentRoll.jailed > 0 && !currentRoll.released_from_jail
                      ? `Jail Turn ${currentRoll.jailed}`
                      : `Total ${currentRoll.total}`
                : isTurnOrderPhase
                  ? "Roll to set order"
                  : isCurrentPlayerTurn
                    ? "Start your turn"
                    : "Waiting"}
          </div>

          {isCurrentPlayerTurn && !currentRoll && !isRollingAnimation ? (
            <button
              type="button"
              onClick={handleRoll}
              disabled={rollMutation.isPending}
              style={{
                width: "100%",
                padding: "12px 14px",
                backgroundColor: rollMutation.isPending ? "#D0D3D4" : "#F76902",
                color: "#FFFFFF",
                fontWeight: 700,
                cursor: rollMutation.isPending ? "not-allowed" : "pointer",
              }}
            >
              {rollLabel}
            </button>
          ) : null}

          {isCurrentPlayerTurn && showMoveButton && !isRollingAnimation ? (
            <button
              type="button"
              onClick={handleMove}
              disabled={moveMutation.isPending}
              style={{
                width: "100%",
                padding: "12px 14px",
                backgroundColor: moveMutation.isPending ? "#D0D3D4" : "#F76902",
                color: "#FFFFFF",
                fontWeight: 700,
                cursor: moveMutation.isPending ? "not-allowed" : "pointer",
              }}
            >
              Move
            </button>
          ) : null}

          {isCurrentPlayerTurn && showTurnOrderEndTurn && !isRollingAnimation ? (
            <button
              type="button"
              onClick={handleEndTurn}
              disabled={endTurnMutation.isPending}
              style={{
                width: "100%",
                padding: "12px 14px",
                backgroundColor: endTurnMutation.isPending ? "#D0D3D4" : "#F76902",
                color: "#FFFFFF",
                fontWeight: 700,
                cursor: endTurnMutation.isPending ? "not-allowed" : "pointer",
              }}
            >
              End Turn
            </button>
          ) : null}

          {isCurrentPlayerTurn && showJailEndTurn && !isRollingAnimation ? (
            <button
              type="button"
              onClick={handleEndTurn}
              disabled={endTurnMutation.isPending}
              style={{
                width: "100%",
                padding: "12px 14px",
                backgroundColor: endTurnMutation.isPending ? "#D0D3D4" : "#F76902",
                color: "#FFFFFF",
                fontWeight: 700,
                cursor: endTurnMutation.isPending ? "not-allowed" : "pointer",
              }}
            >
              End Turn
            </button>
          ) : null}

          {isCurrentPlayerTurn && showSentToJailEndTurn && !isRollingAnimation ? (
            <button
              type="button"
              onClick={handleEndTurn}
              disabled={endTurnMutation.isPending}
              style={{
                width: "100%",
                padding: "12px 14px",
                backgroundColor: endTurnMutation.isPending ? "#D0D3D4" : "#F76902",
                color: "#FFFFFF",
                fontWeight: 700,
                cursor: endTurnMutation.isPending ? "not-allowed" : "pointer",
              }}
            >
              End Turn
            </button>
          ) : null}

          {isCurrentPlayerTurn && showJailReleaseButtons && !isRollingAnimation ? (
            <div
              style={{
                width: "100%",
                display: "flex",
                flexDirection: "column",
                gap: 10,
              }}
            >
              {currentPlayer && currentPlayer.get_out_of_jail_cards > 0 ? (
                <button
                  type="button"
                  onClick={() => handleJailRelease("card")}
                  disabled={jailReleaseMutation.isPending}
                  style={{
                    width: "100%",
                    padding: "12px 14px",
                    backgroundColor: jailReleaseMutation.isPending ? "#D0D3D4" : "#F76902",
                    color: "#FFFFFF",
                    fontWeight: 700,
                    cursor: jailReleaseMutation.isPending ? "not-allowed" : "pointer",
                  }}
                >
                  Use Card
                </button>
              ) : null}
              <button
                type="button"
                onClick={() => handleJailRelease("pay")}
                disabled={jailReleaseMutation.isPending}
                style={{
                  width: "100%",
                  padding: "12px 14px",
                  backgroundColor: jailReleaseMutation.isPending ? "#D0D3D4" : "#F76902",
                  color: "#FFFFFF",
                  fontWeight: 700,
                  cursor: jailReleaseMutation.isPending ? "not-allowed" : "pointer",
                }}
              >
                Pay ₮50
              </button>
            </div>
          ) : null}

          {isCurrentPlayerTurn && actionError ? (
            <div
              style={{
                color: "#D32F2F",
                fontSize: 12,
              }}
            >
              {actionError}
            </div>
          ) : null}
        </div>
      ) : null}

      <canvas ref={canvasRef} />
    </div>
  )
}
