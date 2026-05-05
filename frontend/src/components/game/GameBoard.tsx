"use client"

import { type MouseEvent, type WheelEvent, useEffect, useMemo, useRef, useState } from "react"
import { getTokenIcon } from "@/utils/tokens"
import { DiceRoll, GameState, Player } from "@/types"
import { useReadyUp } from "@/hooks/playerHooks"
import { useEndTurn } from "@/hooks/playerHooks"
import {
  useDrawCard,
  useJailRelease,
  useMovePlayer,
  usePayBank,
  usePayRent,
  usePlayerBankrupt,
  usePlayerExchange,
  useReceiveBankPayout,
  useResolveCard,
  useRollDice,
} from "@/hooks/useGameAPI"
import { useIgnorePropertyPurchase, usePurchaseProperty } from "@/hooks/propertyHooks"
import { emitToast } from "@/utils/toast"

interface GameBoardProps {
  sessionId: string
  playerId: string
  playerName: string
  currentPlayerTurnId?: number | null
  gameState: GameState
  activePropertyId?: number | null
  onHoverProperty: (propertyId: number | null) => void
}

const BOARD_UNITS = 37
const CORNER_UNITS = 5
const EDGE_UNITS = 3
const BOARD_SIZE = 1850
const BOARD_MARGIN = 80
const CENTER_CARD_WIDTH = BOARD_SIZE * 0.22
const CENTER_CARD_SPECS = {
  COMMUNITY: { x: BOARD_SIZE * 0.32, y: BOARD_SIZE * 0.32, rotation: 135 },
  CHANCE: { x: BOARD_SIZE * 0.68, y: BOARD_SIZE * 0.68, rotation: -45 },
} as const

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

function isCardTile(index: number) {
  return index === 2 || index === 7 || index === 17 || index === 22 || index === 33 || index === 36
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

function getBaseZoom(width: number) {
  return width / (BOARD_SIZE + BOARD_MARGIN * 2)
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

function getCoverTopSize(sourceWidth: number, sourceHeight: number, boxWidth: number, boxHeight: number) {
  const scale = Math.max(boxWidth / sourceWidth, boxHeight / sourceHeight)

  return {
    width: sourceWidth * scale,
    height: sourceHeight * scale,
    x: (boxWidth - sourceWidth * scale) / 2,
    y: 0,
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

function getOutsideOwnerTokenPosition(placement: TilePlacement) {
  const bounds = getTileBounds(placement)
  const offset = 24

  switch (placement.side) {
    case "bottom":
      return {
        x: bounds.x + bounds.width / 2,
        y: bounds.y + bounds.height + offset,
      }
    case "left":
      return {
        x: bounds.x - offset,
        y: bounds.y + bounds.height / 2,
      }
    case "top":
      return {
        x: bounds.x + bounds.width / 2,
        y: bounds.y - offset,
      }
    case "right":
      return {
        x: bounds.x + bounds.width + offset,
        y: bounds.y + bounds.height / 2,
      }
  }
}

function getMovementPath(oldPosition: number, newPosition: number, passedGo: boolean, fromCard: boolean) {
  const path = [getTileCenter(oldPosition)]
  let current = oldPosition
  const moveBackward = fromCard && !passedGo && newPosition < oldPosition

  while (current !== newPosition) {
    current = moveBackward ? (current + 39) % 40 : (current + 1) % 40
    path.push(getTileCenter(current))
  }

  return path
}

function getTileIndexFromBoardPoint(boardTiles: TilePlacement[], boardX: number, boardY: number) {
  for (const placement of boardTiles) {
    const bounds = getTileBounds(placement)
    if (
      boardX >= bounds.x &&
      boardX <= bounds.x + bounds.width &&
      boardY >= bounds.y &&
      boardY <= bounds.y + bounds.height
    ) {
      return placement.index
    }
  }

  return null
}

function getScreenPointFromBoardPoint(
  boardX: number,
  boardY: number,
  size: { width: number; height: number },
  viewport: ViewportState,
) {
  const baseZoom = getBaseZoom(size.width)
  const scaledX = (boardX - BOARD_SIZE / 2) * baseZoom * viewport.zoom
  const scaledY = (boardY - BOARD_SIZE / 2) * baseZoom * viewport.zoom
  const rotated = rotatePoint(scaledX, scaledY, viewport.rotation)

  return {
    x: size.width / 2 + viewport.x + rotated.x,
    y: size.height / 2 + viewport.y + rotated.y,
  }
}

function getPropertyCalloutAnchor(placement: TilePlacement) {
  const bounds = getTileBounds(placement)
  const offset = 8

  switch (placement.side) {
    case "bottom":
      return {
        anchorX: bounds.x + bounds.width / 2,
        anchorY: bounds.y + bounds.height + offset,
      }
    case "left":
      return {
        anchorX: bounds.x - offset,
        anchorY: bounds.y + bounds.height / 2,
      }
    case "top":
      return {
        anchorX: bounds.x + bounds.width / 2,
        anchorY: bounds.y - offset,
      }
    case "right":
      return {
        anchorX: bounds.x + bounds.width + offset,
        anchorY: bounds.y + bounds.height / 2,
      }
  }
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

function rotatePoint(x: number, y: number, degrees: number) {
  const radians = (degrees * Math.PI) / 180
  const cos = Math.cos(radians)
  const sin = Math.sin(radians)

  return {
    x: x * cos - y * sin,
    y: x * sin + y * cos,
  }
}

function getBoardPointFromScreen(
  clientX: number,
  clientY: number,
  rect: DOMRect,
  viewport: ViewportState,
) {
  const centeredX = clientX - rect.left - rect.width / 2
  const centeredY = clientY - rect.top - rect.height / 2
  const unpannedX = centeredX - viewport.x
  const unpannedY = centeredY - viewport.y
  const baseZoom = getBaseZoom(rect.width)
  const unzoomedX = unpannedX / (baseZoom * viewport.zoom)
  const unzoomedY = unpannedY / (baseZoom * viewport.zoom)
  const unrotated = rotatePoint(unzoomedX, unzoomedY, -viewport.rotation)

  return {
    x: unrotated.x + BOARD_SIZE / 2,
    y: unrotated.y + BOARD_SIZE / 2,
  }
}

function isInsideCenterCard(
  boardX: number,
  boardY: number,
  cardType: string,
) {
  const spec = cardType === "COMMUNITY" ? CENTER_CARD_SPECS.COMMUNITY : CENTER_CARD_SPECS.CHANCE
  const local = rotatePoint(boardX - spec.x, boardY - spec.y, -spec.rotation)
  const halfSize = CENTER_CARD_WIDTH / 2

  return Math.abs(local.x) <= halfSize && Math.abs(local.y) <= halfSize
}

export default function GameBoard({
  sessionId,
  playerId,
  currentPlayerTurnId,
  gameState,
  activePropertyId,
  onHoverProperty,
}: GameBoardProps) {
  const containerRef = useRef<HTMLDivElement | null>(null)
  const canvasRef = useRef<HTMLCanvasElement | null>(null)
  const tileImagesRef = useRef<Record<number, HTMLImageElement>>({})
  const tokenImagesRef = useRef<Record<number, HTMLImageElement>>({})
  const centerImagesRef = useRef<Record<string, HTMLImageElement>>({})
  const rotationRef = useRef(0)
  const dragRef = useRef<{ active: boolean; moved: boolean; x: number; y: number }>({
    active: false,
    moved: false,
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
  const [rollDisplay, setRollDisplay] = useState<DiceRoll | null>(null)
  const [isRollingAnimation, setIsRollingAnimation] = useState(false)
  const [moveAnimation, setMoveAnimation] = useState<MoveAnimation | null>(null)
  const [showBankruptConfirm, setShowBankruptConfirm] = useState(false)

  const readyMutation = useReadyUp()
  const rollMutation = useRollDice()
  const moveMutation = useMovePlayer()
  const endTurnMutation = useEndTurn()
  const drawCardMutation = useDrawCard()
  const resolveCardMutation = useResolveCard()
  const jailReleaseMutation = useJailRelease()
  const payBankMutation = usePayBank()
  const receiveBankPayoutMutation = useReceiveBankPayout()
  const payRentMutation = usePayRent()
  const playerExchangeMutation = usePlayerExchange()
  const playerBankruptMutation = usePlayerBankrupt()
  const purchasePropertyMutation = usePurchaseProperty()
  const ignorePropertyMutation = useIgnorePropertyPurchase()

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
  const propertyOwnersByTile = useMemo(() => {
    const owners: Record<number, Player> = {}

    gameState.players.forEach((playerInfo) => {
      ;(playerInfo.owned_properties ?? []).forEach((property) => {
        const tile = gameState.tiles.find((tileInfo) => tileInfo.property_data?.id === property.property_info.id)
        if (!tile) return
        owners[tile.id] = playerInfo.player
      })
    })

    return owners
  }, [gameState.players, gameState.tiles])
  const activePropertyTile = useMemo(
    () => gameState.tiles.find((tile) => tile.property_data?.id === activePropertyId) ?? null,
    [activePropertyId, gameState.tiles],
  )
  const activePropertyPlacement = useMemo(
    () => (activePropertyTile ? getTilePlacement(activePropertyTile.id) : null),
    [activePropertyTile],
  )
  const activePropertyCallout = useMemo(() => {
    if (!activePropertyTile || !activePropertyPlacement || size.width === 0 || size.height === 0) return null

    const anchor = getPropertyCalloutAnchor(activePropertyPlacement)
    const screen = getScreenPointFromBoardPoint(anchor.anchorX, anchor.anchorY, size, viewport)

    return {
      screen,
      side: activePropertyPlacement.side,
      propertyId: activePropertyTile.property_data?.id ?? null,
      propertyName: activePropertyTile.property_data?.name ?? "",
    }
  }, [activePropertyPlacement, activePropertyTile, size, viewport])

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
  const extraRollPlayerId = gameState.extra_roll_player_id ?? null
  const pendingCardDraw = gameState.pending_card_draw ?? null
  const drawnCard = gameState.drawn_card ?? null
  const pendingRent = gameState.pending_rent ?? null
  const pendingBankPayment = gameState.pending_bank_payment ?? null
  const pendingBankPayout = gameState.pending_bank_payout ?? null
  const pendingExchange = gameState.pending_exchange ?? null
  const pendingPropertyPurchase = gameState.pending_property_purchase ?? null
  const pendingPropertyData = useMemo(
    () =>
      pendingPropertyPurchase
        ? gameState.tiles.find((tile) => tile.property_data?.id === pendingPropertyPurchase.property_id)?.property_data ?? null
        : null,
    [gameState.tiles, pendingPropertyPurchase],
  )
  const rentRecipient = useMemo(
    () =>
      pendingRent
        ? gameState.players.find((playerInfo) => playerInfo.player.id === pendingRent.to_player_id)?.player ?? null
        : null,
    [gameState.players, pendingRent],
  )

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

    const path = getMovementPath(lastMove.old_position, lastMove.new_position, lastMove.passed_go, lastMove.from_card)
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
  }, [lastMove])

  useEffect(() => {
    setShowBankruptConfirm(false)
  }, [pendingRent, pendingBankPayment, currentPlayerTurnId])

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

    const baseZoom = getBaseZoom(size.width)

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

      if (activePropertyTile?.id === placement.index) {
        ctx.save()
        ctx.strokeStyle = "#F76902"
        ctx.lineWidth = 8
        ctx.strokeRect(x + 2, y + 2, width - 4, height - 4)
        ctx.restore()
      }

      const playersOnTile = (playerPositions[placement.index] || []).filter(
        (player) => player.id !== moveAnimation?.playerId,
      )
      if (playersOnTile.length > 0) {
        const tokenSize = isCornerTile ? 64 : 46
        const gap = 8
        const tokenSlots = getTileTokenSlots(placement, playersOnTile, tokenSize, gap)

        playersOnTile.forEach((player, index) => {
          const tokenImage = tokenImagesRef.current[player.piece_token]
          if (!tokenImage) return

          const slot = tokenSlots[index]
          const tokenX = slot.x
          const tokenY = slot.y
          const isOwnPlayer = player.id.toString() === playerId

          ctx.save()
          ctx.fillStyle = "#FFFFFF"
          ctx.beginPath()
          ctx.arc(tokenX + tokenSize / 2, tokenY + tokenSize / 2, tokenSize / 2, 0, Math.PI * 2)
          ctx.fill()

          ctx.strokeStyle = isOwnPlayer ? "#F76902" : "#000000"
          ctx.lineWidth = isOwnPlayer ? 4 : 3
          ctx.stroke()

          ctx.save()
          ctx.beginPath()
          ctx.arc(tokenX + tokenSize / 2, tokenY + tokenSize / 2, tokenSize / 2 - 1, 0, Math.PI * 2)
          ctx.clip()
          const drawSize = getCoverTopSize(tokenImage.width, tokenImage.height, tokenSize * 0.72, tokenSize * 0.72)
          ctx.drawImage(
            tokenImage,
            tokenX + tokenSize * 0.14 + drawSize.x,
            tokenY + tokenSize * 0.14 + drawSize.y,
            drawSize.width,
            drawSize.height,
          )
          ctx.restore()
          ctx.restore()
        })
      }

      const propertyOwner = propertyOwnersByTile[placement.index]
      if (propertyOwner) {
        const ownerTokenImage = tokenImagesRef.current[propertyOwner.piece_token]
        if (ownerTokenImage) {
          const ownerToken = getOutsideOwnerTokenPosition(placement)
          const ownerTokenSize = 34
          const isOwnOwnerToken = propertyOwner.id.toString() === playerId

          ctx.save()
          ctx.fillStyle = "#FFFFFF"
          ctx.beginPath()
          ctx.arc(ownerToken.x, ownerToken.y, ownerTokenSize / 2, 0, Math.PI * 2)
          ctx.fill()
          ctx.strokeStyle = isOwnOwnerToken ? "#F76902" : "#000000"
          ctx.lineWidth = isOwnOwnerToken ? 3 : 2
          ctx.stroke()
          ctx.save()
          ctx.beginPath()
          ctx.arc(ownerToken.x, ownerToken.y, ownerTokenSize / 2 - 1, 0, Math.PI * 2)
          ctx.clip()
          const drawSize = getCoverTopSize(ownerTokenImage.width, ownerTokenImage.height, ownerTokenSize * 0.72, ownerTokenSize * 0.72)
          ctx.drawImage(
            ownerTokenImage,
            ownerToken.x - ownerTokenSize * 0.36 + drawSize.x,
            ownerToken.y - ownerTokenSize * 0.36 + drawSize.y,
            drawSize.width,
            drawSize.height,
          )
          ctx.restore()
          ctx.restore()
        }
      }
    })

    if (moveAnimation) {
      const tokenImage = tokenImagesRef.current[moveAnimation.pieceToken]
      if (tokenImage) {
        const tokenSize = 52
        const isOwnPlayer = moveAnimation.playerId.toString() === playerId

        ctx.save()
        ctx.fillStyle = "#FFFFFF"
        ctx.beginPath()
        ctx.arc(moveAnimation.x, moveAnimation.y, tokenSize / 2, 0, Math.PI * 2)
        ctx.fill()

        ctx.strokeStyle = isOwnPlayer ? "#F76902" : "#000000"
        ctx.lineWidth = isOwnPlayer ? 4 : 3
        ctx.stroke()

        ctx.save()
        ctx.beginPath()
        ctx.arc(moveAnimation.x, moveAnimation.y, tokenSize / 2 - 1, 0, Math.PI * 2)
        ctx.clip()
        const drawSize = getCoverTopSize(tokenImage.width, tokenImage.height, tokenSize * 0.72, tokenSize * 0.72)
        ctx.drawImage(
          tokenImage,
          moveAnimation.x - tokenSize * 0.36 + drawSize.x,
          moveAnimation.y - tokenSize * 0.36 + drawSize.y,
          drawSize.width,
          drawSize.height,
        )
        ctx.restore()
        ctx.restore()
      }
    }

    const drawCenterCard = (
      image: HTMLImageElement | undefined,
      centerX: number,
      centerY: number,
      width: number,
      rotation: number,
      isHighlighted: boolean,
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
      if (isHighlighted) {
        ctx.strokeStyle = "#F76902"
        ctx.lineWidth = 8
        ctx.strokeRect(
          -drawSize.width / 2,
          -drawSize.height / 2,
          drawSize.width,
          drawSize.height,
        )
      }
      ctx.restore()
    }

    drawCenterCard(
      centerImagesRef.current.cchest,
      CENTER_CARD_SPECS.COMMUNITY.x,
      CENTER_CARD_SPECS.COMMUNITY.y,
      CENTER_CARD_WIDTH,
      CENTER_CARD_SPECS.COMMUNITY.rotation,
      pendingCardDraw?.card_type === "COMMUNITY" || drawnCard?.card_type === "COMMUNITY",
    )
    drawCenterCard(
      centerImagesRef.current.chance,
      CENTER_CARD_SPECS.CHANCE.x,
      CENTER_CARD_SPECS.CHANCE.y,
      CENTER_CARD_WIDTH,
      CENTER_CARD_SPECS.CHANCE.rotation,
      pendingCardDraw?.card_type === "CHANCE" || drawnCard?.card_type === "CHANCE",
    )

    ctx.restore()
  }, [activePropertyTile, boardTiles, currentPlayerTurnId, drawnCard, imageVersion, moveAnimation, pendingCardDraw, playerPositions, propertyOwnersByTile, size, viewport])

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
      moved: false,
      x: event.clientX,
      y: event.clientY,
    }
    setIsDragging(true)
  }

  const onMouseMove = (event: MouseEvent<HTMLDivElement>) => {
    if (!dragRef.current.active) {
      const rect = containerRef.current?.getBoundingClientRect()
      if (!rect) return

      const boardPoint = getBoardPointFromScreen(event.clientX, event.clientY, rect, viewport)
      const tileIndex = getTileIndexFromBoardPoint(boardTiles, boardPoint.x, boardPoint.y)
      if (tileIndex == null) {
        onHoverProperty(null)
        return
      }

      const tile = gameState.tiles.find((tileInfo) => tileInfo.id === tileIndex) ?? null
      onHoverProperty(tile?.property_data?.id ?? null)
      return
    }

    const dx = event.clientX - dragRef.current.x
    const dy = event.clientY - dragRef.current.y
    const moved = dragRef.current.moved || Math.abs(dx) > 2 || Math.abs(dy) > 2

    dragRef.current = {
      active: true,
      moved,
      x: event.clientX,
      y: event.clientY,
    }

    setViewport((value) => ({
      ...value,
      x: value.x + dx,
      y: value.y + dy,
    }))
  }

  const onMouseUp = (event: MouseEvent<HTMLDivElement>) => {
    const wasDragging = dragRef.current.moved
    stopDrag()

    if (wasDragging || !pendingCardDraw || !isCurrentPlayerTurn) return

    const rect = containerRef.current?.getBoundingClientRect()
    if (!rect) return

    const boardPoint = getBoardPointFromScreen(event.clientX, event.clientY, rect, viewport)
    if (!isInsideCenterCard(boardPoint.x, boardPoint.y, pendingCardDraw.card_type)) return

    handleDrawCard()
  }

  const stopDrag = () => {
    dragRef.current.active = false
    dragRef.current.moved = false
    setIsDragging(false)
  }

  const toggleReady = () => {
    if (!currentLobbyPlayer || isGameStarted) return

    readyMutation.mutate(!currentLobbyPlayer.player.ready_up_status, {
      onError: (error: Error) => {
        emitToast(error.message)
      },
    })
  }

  const handleRoll = () => {
    if (!isCurrentPlayerTurn) return

    setShowBankruptConfirm(false)
    rollMutation.mutate(
      {
        playerId,
        sessionId,
      },
      {
        onError: (error: Error) => {
          emitToast(error.message)
        },
      },
    )
  }

  const handleMove = () => {
    if (!isCurrentPlayerTurn) return

    setShowBankruptConfirm(false)
    moveMutation.mutate(
      {
        playerId,
        sessionId,
      },
      {
        onError: (error: Error) => {
          emitToast(error.message)
        },
      },
    )
  }

  const handleEndTurn = () => {
    if (!isCurrentPlayerTurn) return

    setShowBankruptConfirm(false)
    endTurnMutation.mutate(undefined, {
      onError: (error: Error) => {
        emitToast(error.message)
      },
    })
  }

  const handleJailRelease = (method: string) => {
    if (!isCurrentPlayerTurn) return

    setShowBankruptConfirm(false)
    jailReleaseMutation.mutate(method, {
      onError: (error: Error) => {
        emitToast(error.message)
      },
    })
  }

  const handlePurchaseProperty = () => {
    if (!isCurrentPlayerTurn || !pendingPropertyPurchase) return

    setShowBankruptConfirm(false)
    purchasePropertyMutation.mutate(undefined, {
      onError: (error: Error) => {
        emitToast(error.message)
      },
    })
  }

  const handleIgnoreProperty = () => {
    if (!isCurrentPlayerTurn) return

    setShowBankruptConfirm(false)
    ignorePropertyMutation.mutate(undefined, {
      onError: (error: Error) => {
        emitToast(error.message)
      },
    })
  }

  const handlePayBank = () => {
    if (!isCurrentPlayerTurn) return

    setShowBankruptConfirm(false)
    payBankMutation.mutate(undefined, {
      onError: (error: Error) => {
        emitToast(error.message)
      },
    })
  }

  const handleReceiveBankPayout = () => {
    if (!isCurrentPlayerTurn) return

    setShowBankruptConfirm(false)
    receiveBankPayoutMutation.mutate(undefined, {
      onError: (error: Error) => {
        emitToast(error.message)
      },
    })
  }

  const handlePayRent = () => {
    if (!isCurrentPlayerTurn || !pendingRent) return

    setShowBankruptConfirm(false)
    payRentMutation.mutate(
      {
        dst_player: pendingRent.to_player_id.toString(),
        amount: pendingRent.amount.toString(),
      },
      {
        onError: (error: Error) => {
          emitToast(error.message)
        },
      },
    )
  }

  const handlePlayerExchange = () => {
    if (!isCurrentPlayerTurn) return

    setShowBankruptConfirm(false)
    playerExchangeMutation.mutate(undefined, {
      onError: (error: Error) => {
        emitToast(error.message)
      },
    })
  }

  const handleBankrupt = () => {
    if (!isCurrentPlayerTurn) return

    setShowBankruptConfirm(false)
    playerBankruptMutation.mutate(undefined, {
      onError: (error: Error) => {
        emitToast(error.message)
      },
    })
  }

  const handleDrawCard = () => {
    if (!isCurrentPlayerTurn || !pendingCardDraw) return

    setShowBankruptConfirm(false)
    drawCardMutation.mutate(undefined, {
      onError: (error: Error) => {
        emitToast(error.message)
      },
    })
  }

  const handleResolveCard = () => {
    if (!isCurrentPlayerTurn || !drawnCard) return

    setShowBankruptConfirm(false)
    resolveCardMutation.mutate(undefined, {
      onError: (error: Error) => {
        emitToast(error.message)
      },
    })
  }

  const rollLabel = isTurnOrderPhase
    ? "Roll for Order"
    : extraRollPlayerId === currentPlayerTurnId
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
    currentRoll.jailed < 4

  const showJailReleaseButtons =
    !!currentRoll &&
    !isTurnOrderPhase &&
    currentRoll.jailed > 0 &&
    !currentRoll.released_from_jail &&
    !currentRoll.is_double

  const promptsReady = !isRollingAnimation && !moveAnimation
  const showTurnOrderEndTurn = !!currentRoll && isTurnOrderPhase
  const showSentToJailEndTurn = !!currentRoll && !isTurnOrderPhase && currentRoll.sent_to_jail
  const showCardDrawPanel = !!pendingCardDraw && promptsReady
  const showDrawnCardPanel = !!drawnCard && promptsReady
  const showRentPanel = !!pendingRent && promptsReady
  const showBankPaymentPanel = !!pendingBankPayment && promptsReady
  const showBankPayoutPanel = !!pendingBankPayout && promptsReady
  const showPlayerExchangePanel = !!pendingExchange && promptsReady
  const showPropertyPurchasePanel = !!pendingPropertyPurchase && promptsReady
  const wasJustSentToJail =
    !!lastMove &&
    activePlayer?.id === lastMove.player_id &&
    activePlayer.jailed > 0 &&
    lastMove.new_position === 10
  const showSentToJailPanel = wasJustSentToJail && promptsReady
  const activePlayerNeedsJailRoll = !!activePlayer && activePlayer.jailed > 0 && !currentRoll && !wasJustSentToJail
  const currentPlayerHasExtraRoll = promptsReady && extraRollPlayerId === currentPlayerTurnId
  const activePrompt =
    showBankPayoutPanel ? "bankPayout"
    : showCardDrawPanel ? "cardDraw"
    : showDrawnCardPanel ? "drawnCard"
    : showRentPanel ? "rent"
    : showBankPaymentPanel ? "bankPayment"
    : showPlayerExchangePanel ? "playerExchange"
    : showPropertyPurchasePanel ? "propertyPurchase"
    : showSentToJailPanel ? "sentToJail"
    : "turn"
  const showTurnPanel =
    !isGameStarted ||
    showSentToJailPanel ||
    showCardDrawPanel ||
    showDrawnCardPanel ||
    showRentPanel ||
    showBankPaymentPanel ||
    showBankPayoutPanel ||
    showPlayerExchangePanel ||
    showPropertyPurchasePanel ||
    isRollingAnimation ||
    !!currentRoll ||
    activePlayerNeedsJailRoll ||
    (isCurrentPlayerTurn && currentPlayerHasExtraRoll) ||
    (promptsReady && !lastMove && isCurrentPlayerTurn)
  const showDoublesNotice =
    !isTurnOrderPhase &&
    !!currentRoll &&
    promptsReady &&
    currentRoll.is_double &&
    currentRoll.roll_again &&
    !currentRoll.sent_to_jail &&
    currentRoll.jailed === 0

  return (
    <div
      ref={containerRef}
      className="w-full h-full"
      onWheel={onWheel}
      onMouseDown={onMouseDown}
      onMouseMove={onMouseMove}
      onMouseUp={onMouseUp}
      onMouseLeave={() => {
        stopDrag()
        onHoverProperty(null)
      }}
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
          {activePrompt === "bankPayout" ? (
            <>
              <div
                style={{
                  color: "#F76902",
                  fontSize: 20,
                  fontWeight: 700,
                }}
              >
                Bank Payout
              </div>

              <div
                style={{
                  color: "#7C878E",
                  fontSize: 13,
                }}
              >
                {activePlayer ? `${activePlayer.name}` : "Waiting"}
              </div>

              <div
                style={{
                  color: "#000000",
                  fontSize: 14,
                  fontWeight: 600,
                }}
              >
                {isCurrentPlayerTurn
                  ? pendingBankPayout!.reason
                  : activePlayer
                    ? `Waiting for ${activePlayer.name} to receive`
                    : "Waiting"}
              </div>

              <div
                style={{
                  color: "#7C878E",
                  fontSize: 13,
                }}
              >
                Amount: ₮{pendingBankPayout!.amount.toLocaleString()}
              </div>
            </>
          ) : activePrompt === "cardDraw" ? (
            <>
              <div
                style={{
                  color: "#F76902",
                  fontSize: 20,
                  fontWeight: 700,
                }}
              >
                {pendingCardDraw!.tile_name}
              </div>

              <div
                style={{
                  color: "#7C878E",
                  fontSize: 13,
                }}
              >
                {activePlayer ? `${activePlayer.name}` : "Waiting"}
              </div>

              <div
                style={{
                  color: "#000000",
                  fontSize: 14,
                  fontWeight: 600,
                }}
              >
                {isCurrentPlayerTurn
                  ? `Click the ${pendingCardDraw!.tile_name} card to draw`
                  : activePlayer
                    ? `Waiting for ${activePlayer.name} to draw`
                    : "Waiting"}
              </div>
            </>
          ) : activePrompt === "drawnCard" ? (
            <>
              <div
                style={{
                  color: "#F76902",
                  fontSize: 20,
                  fontWeight: 700,
                }}
              >
                {drawnCard!.card_type === "CHANCE" ? "Chance" : "Community Chest"}
              </div>

              <div
                style={{
                  color: "#7C878E",
                  fontSize: 13,
                }}
              >
                {activePlayer ? `${activePlayer.name}` : "Waiting"}
              </div>

              <div
                style={{
                  color: "#000000",
                  fontSize: 16,
                  fontWeight: 700,
                }}
              >
                {drawnCard!.name}
              </div>

              <div
                style={{
                  color: "#000000",
                  fontSize: 14,
                  fontWeight: 600,
                }}
              >
                {drawnCard!.description}
              </div>
            </>
          ) : activePrompt === "rent" ? (
            <>
              <div
                style={{
                  color: "#F76902",
                  fontSize: 20,
                  fontWeight: 700,
                }}
              >
                Rent
              </div>

              <div
                style={{
                  color: "#7C878E",
                  fontSize: 13,
                }}
              >
                {activePlayer ? `${activePlayer.name}` : "Waiting"}
              </div>

              <div
                style={{
                  color: "#000000",
                  fontSize: 14,
                  fontWeight: 600,
                }}
              >
                {isCurrentPlayerTurn
                  ? rentRecipient
                    ? `Pay ${rentRecipient.name}`
                    : "Pay rent"
                  : activePlayer
                    ? `Waiting for ${activePlayer.name} to pay rent`
                    : "Waiting"}
              </div>

              <div
                style={{
                  color: "#7C878E",
                  fontSize: 13,
                }}
              >
                Amount: ₮{pendingRent!.amount.toLocaleString()}
              </div>
            </>
          ) : activePrompt === "bankPayment" ? (
            <>
              <div
                style={{
                  color: "#F76902",
                  fontSize: 20,
                  fontWeight: 700,
                }}
              >
                {pendingBankPayment!.reason === "jail release" ? "Leave Jail" : "Bank Payment"}
              </div>

              <div
                style={{
                  color: "#7C878E",
                  fontSize: 13,
                }}
              >
                {activePlayer ? `${activePlayer.name}` : "Waiting"}
              </div>

              <div
                style={{
                  color: "#000000",
                  fontSize: 14,
                  fontWeight: 600,
                }}
              >
                {isCurrentPlayerTurn
                  ? pendingBankPayment!.reason === "jail release"
                    ? "Pay to leave jail"
                    : pendingBankPayment!.reason
                  : activePlayer
                    ? pendingBankPayment!.reason === "jail release"
                      ? `Waiting for ${activePlayer.name} to leave jail`
                      : `Waiting for ${activePlayer.name} to pay`
                    : "Waiting"}
              </div>

              <div
                style={{
                  color: "#7C878E",
                  fontSize: 13,
                }}
              >
                Amount: ₮{pendingBankPayment!.amount.toLocaleString()}
              </div>
            </>
          ) : activePrompt === "playerExchange" ? (
            <>
              <div
                style={{
                  color: "#F76902",
                  fontSize: 20,
                  fontWeight: 700,
                }}
              >
                Player Exchange
              </div>

              <div
                style={{
                  color: "#7C878E",
                  fontSize: 13,
                }}
              >
                {activePlayer ? `${activePlayer.name}` : "Waiting"}
              </div>

              <div
                style={{
                  color: "#000000",
                  fontSize: 14,
                  fontWeight: 600,
                }}
              >
                {pendingExchange!.is_paying_all
                  ? `Pay each player ₮${pendingExchange!.amount.toLocaleString()}`
                  : `Collect ₮${pendingExchange!.amount.toLocaleString()} from each player`}
              </div>

              <div
                style={{
                  color: "#7C878E",
                  fontSize: 13,
                }}
              >
                {isCurrentPlayerTurn
                  ? "Continue"
                  : activePlayer
                    ? `Waiting for ${activePlayer.name}`
                    : "Waiting"}
              </div>
            </>
          ) : activePrompt === "propertyPurchase" ? (
            <>
              <div
                style={{
                  color: "#F76902",
                  fontSize: 20,
                  fontWeight: 700,
                }}
              >
                Property
              </div>

              <div
                style={{
                  color: "#7C878E",
                  fontSize: 13,
                }}
              >
                {activePlayer ? `${activePlayer.name}` : "Waiting"}
              </div>

              <div
                style={{
                  color: "#000000",
                  fontSize: 16,
                  fontWeight: 700,
                }}
              >
                {pendingPropertyData?.name ?? "Unowned Property"}
              </div>

              <div
                style={{
                  color: "#7C878E",
                  fontSize: 13,
                }}
              >
                Cost: ₮{pendingPropertyPurchase!.purchase_cost.toLocaleString()}
              </div>

              <div
                style={{
                  color: "#000000",
                  fontSize: 14,
                  fontWeight: 600,
                }}
              >
                {isCurrentPlayerTurn
                  ? pendingPropertyPurchase!.can_afford
                    ? "Buy or ignore"
                    : "Cannot afford this property"
                  : activePlayer
                    ? `Waiting for ${activePlayer.name} to decide`
                    : "Waiting"}
              </div>
            </>
          ) : activePrompt === "sentToJail" ? (
            <>
              <div
                style={{
                  color: "#F76902",
                  fontSize: 20,
                  fontWeight: 700,
                }}
              >
                Jail
              </div>

              <div
                style={{
                  color: "#7C878E",
                  fontSize: 13,
                }}
              >
                {activePlayer ? `${activePlayer.name}` : "Waiting"}
              </div>

              <div
                style={{
                  color: "#000000",
                  fontSize: 14,
                  fontWeight: 600,
                }}
              >
                {activePlayer ? `${activePlayer.name} was sent to jail` : "Sent to jail"}
              </div>
            </>
          ) : (
            <>
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
                        ? activePlayer
                          ? `${activePlayer.name} was sent to jail`
                          : "Sent to jail"
                        : currentRoll.jailed > 0 && !currentRoll.released_from_jail
                          ? currentRoll.jailed >= 4
                            ? "No doubles. Leave jail."
                            : "No doubles. Stay in jail or leave."
                          : `Total ${currentRoll.total}`
                    : activePlayerNeedsJailRoll
                      ? isCurrentPlayerTurn
                        ? `Roll for doubles to leave jail`
                        : activePlayer
                          ? `Waiting for ${activePlayer.name} to roll for doubles`
                          : "Waiting"
                    : isTurnOrderPhase
                      ? "Roll to set order"
                    : isCurrentPlayerTurn
                      ? "Start your turn"
                      : "Waiting"}
              </div>

              {showDoublesNotice ? (
                <div
                  style={{
                    width: "100%",
                    border: "2px solid #F76902",
                    backgroundColor: "#FFF3E0",
                    padding: "10px 12px",
                    display: "flex",
                    flexDirection: "column",
                    gap: 4,
                  }}
                >
                  <div
                    style={{
                      color: "#F76902",
                      fontSize: 15,
                      fontWeight: 700,
                    }}
                  >
                    Doubles
                  </div>

                  <div
                    style={{
                      color: "#000000",
                      fontSize: 13,
                      fontWeight: 600,
                    }}
                  >
                    {isCurrentPlayerTurn
                      ? "Move now. You will roll again after this move."
                      : activePlayer
                        ? `${activePlayer.name} will roll again after this move.`
                        : "Another roll is coming after this move."}
                  </div>
                </div>
              ) : null}
            </>
          )}

          {isCurrentPlayerTurn && activePrompt === "drawnCard" ? (
            <button
              type="button"
              onClick={handleResolveCard}
              disabled={resolveCardMutation.isPending}
              style={{
                width: "100%",
                padding: "12px 14px",
                backgroundColor: resolveCardMutation.isPending ? "#D0D3D4" : "#F76902",
                color: "#FFFFFF",
                fontWeight: 700,
                cursor: resolveCardMutation.isPending ? "not-allowed" : "pointer",
              }}
            >
              Continue
            </button>
          ) : null}

          {isCurrentPlayerTurn && activePrompt === "rent" ? (
            <div
              style={{
                width: "100%",
                display: "flex",
                flexDirection: "column",
                gap: 10,
              }}
            >
              <div
                style={{
                  width: "100%",
                  display: "flex",
                  gap: 10,
                }}
              >
                <button
                  type="button"
                  onClick={handlePayRent}
                  disabled={payRentMutation.isPending || playerBankruptMutation.isPending}
                  style={{
                    flex: 1,
                    padding: "12px 14px",
                    backgroundColor: payRentMutation.isPending || playerBankruptMutation.isPending ? "#D0D3D4" : "#F76902",
                    color: "#FFFFFF",
                    fontWeight: 700,
                    cursor: payRentMutation.isPending || playerBankruptMutation.isPending ? "not-allowed" : "pointer",
                  }}
                >
                  Pay Rent
                </button>

                <button
                  type="button"
                  onClick={() => setShowBankruptConfirm((value) => !value)}
                  disabled={payRentMutation.isPending || playerBankruptMutation.isPending}
                  style={{
                    flex: 1,
                    padding: "12px 14px",
                    backgroundColor: "#D0D3D4",
                    color: "#000000",
                    fontWeight: 700,
                    cursor: payRentMutation.isPending || playerBankruptMutation.isPending ? "not-allowed" : "pointer",
                    opacity: payRentMutation.isPending || playerBankruptMutation.isPending ? 0.7 : 1,
                  }}
                >
                  {showBankruptConfirm ? "Cancel" : "Bankrupt"}
                </button>
              </div>

              {showBankruptConfirm ? (
                <button
                  type="button"
                  onClick={handleBankrupt}
                  disabled={payRentMutation.isPending || playerBankruptMutation.isPending}
                  style={{
                    width: "100%",
                    padding: "12px 14px",
                    backgroundColor: "#D0D3D4",
                    color: "#000000",
                    fontWeight: 700,
                    cursor: payRentMutation.isPending || playerBankruptMutation.isPending ? "not-allowed" : "pointer",
                    opacity: payRentMutation.isPending || playerBankruptMutation.isPending ? 0.7 : 1,
                  }}
                >
                  Confirm Bankrupt
                </button>
              ) : null}
            </div>
          ) : null}

          {isCurrentPlayerTurn && activePrompt === "bankPayment" ? (
            <div
              style={{
                width: "100%",
                display: "flex",
                flexDirection: "column",
                gap: 10,
              }}
            >
              <div
                style={{
                  width: "100%",
                  display: "flex",
                  gap: 10,
                }}
              >
                <button
                  type="button"
                  onClick={handlePayBank}
                  disabled={payBankMutation.isPending || playerBankruptMutation.isPending}
                  style={{
                    flex: 1,
                    padding: "12px 14px",
                    backgroundColor: payBankMutation.isPending || playerBankruptMutation.isPending ? "#D0D3D4" : "#F76902",
                    color: "#FFFFFF",
                    fontWeight: 700,
                    cursor: payBankMutation.isPending || playerBankruptMutation.isPending ? "not-allowed" : "pointer",
                  }}
                >
                  {pendingBankPayment!.reason === "jail release"
                    ? `Pay ₮${pendingBankPayment!.amount.toLocaleString()}`
                    : "Pay Bank"}
                </button>
                <button
                  type="button"
                  onClick={() => setShowBankruptConfirm((value) => !value)}
                  disabled={payBankMutation.isPending || playerBankruptMutation.isPending}
                  style={{
                    flex: 1,
                    padding: "12px 14px",
                    backgroundColor: "#D0D3D4",
                    color: "#000000",
                    fontWeight: 700,
                    cursor: payBankMutation.isPending || playerBankruptMutation.isPending ? "not-allowed" : "pointer",
                    opacity: payBankMutation.isPending || playerBankruptMutation.isPending ? 0.7 : 1,
                  }}
                >
                  {showBankruptConfirm ? "Cancel" : "Bankrupt"}
                </button>
              </div>

              {showBankruptConfirm ? (
                <button
                  type="button"
                  onClick={handleBankrupt}
                  disabled={payBankMutation.isPending || playerBankruptMutation.isPending}
                  style={{
                    width: "100%",
                    padding: "12px 14px",
                    backgroundColor: "#D0D3D4",
                    color: "#000000",
                    fontWeight: 700,
                    cursor: payBankMutation.isPending || playerBankruptMutation.isPending ? "not-allowed" : "pointer",
                    opacity: payBankMutation.isPending || playerBankruptMutation.isPending ? 0.7 : 1,
                  }}
                >
                  Confirm Bankrupt
                </button>
              ) : null}
            </div>
          ) : null}

          {isCurrentPlayerTurn && activePrompt === "bankPayout" ? (
            <button
              type="button"
              onClick={handleReceiveBankPayout}
              disabled={receiveBankPayoutMutation.isPending}
              style={{
                width: "100%",
                padding: "12px 14px",
                backgroundColor: receiveBankPayoutMutation.isPending ? "#D0D3D4" : "#F76902",
                color: "#FFFFFF",
                fontWeight: 700,
                cursor: receiveBankPayoutMutation.isPending ? "not-allowed" : "pointer",
              }}
            >
              Receive
            </button>
          ) : null}

          {isCurrentPlayerTurn && activePrompt === "playerExchange" ? (
            <button
              type="button"
              onClick={handlePlayerExchange}
              disabled={playerExchangeMutation.isPending}
              style={{
                width: "100%",
                padding: "12px 14px",
                backgroundColor: playerExchangeMutation.isPending ? "#D0D3D4" : "#F76902",
                color: "#FFFFFF",
                fontWeight: 700,
                cursor: playerExchangeMutation.isPending ? "not-allowed" : "pointer",
              }}
            >
              Continue
            </button>
          ) : null}

          {isCurrentPlayerTurn && activePrompt === "propertyPurchase" ? (
            <div
              style={{
                width: "100%",
                display: "flex",
                gap: 10,
              }}
            >
              {pendingPropertyPurchase!.can_afford ? (
                <button
                  type="button"
                  onClick={handlePurchaseProperty}
                  disabled={purchasePropertyMutation.isPending || ignorePropertyMutation.isPending}
                  style={{
                    flex: 1,
                    padding: "12px 14px",
                    backgroundColor: purchasePropertyMutation.isPending || ignorePropertyMutation.isPending ? "#D0D3D4" : "#F76902",
                    color: "#FFFFFF",
                    fontWeight: 700,
                    cursor: purchasePropertyMutation.isPending || ignorePropertyMutation.isPending ? "not-allowed" : "pointer",
                  }}
                >
                  Buy
                </button>
              ) : null}

                <button
                  type="button"
                  onClick={handleIgnoreProperty}
                  disabled={purchasePropertyMutation.isPending || ignorePropertyMutation.isPending}
                  style={{
                    flex: 1,
                    padding: "12px 14px",
                    backgroundColor: "#D0D3D4",
                    color: "#000000",
                    fontWeight: 700,
                    cursor: purchasePropertyMutation.isPending || ignorePropertyMutation.isPending ? "not-allowed" : "pointer",
                    opacity: purchasePropertyMutation.isPending || ignorePropertyMutation.isPending ? 0.7 : 1,
                  }}
                >
                  Ignore
              </button>
            </div>
          ) : null}

          {isCurrentPlayerTurn && activePrompt === "turn" && !currentRoll && !isRollingAnimation ? (
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

          {isCurrentPlayerTurn && activePrompt === "turn" && showMoveButton && !isRollingAnimation ? (
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

          {isCurrentPlayerTurn && activePrompt === "turn" && showTurnOrderEndTurn && !isRollingAnimation ? (
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

          {isCurrentPlayerTurn && activePrompt === "turn" && showJailEndTurn && !isRollingAnimation ? (
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

          {isCurrentPlayerTurn && activePrompt === "turn" && showSentToJailEndTurn && !isRollingAnimation ? (
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

          {isCurrentPlayerTurn && activePrompt === "turn" && showJailReleaseButtons && !isRollingAnimation ? (
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

        </div>
      ) : null}

      <canvas ref={canvasRef} />

      {activePropertyCallout && activePropertyCallout.propertyId ? (
        <div
          style={{
            position: "absolute",
            left: activePropertyCallout.screen.x,
            top: activePropertyCallout.screen.y,
            transform:
              activePropertyCallout.side === "top"
                ? "translate(-50%, -100%)"
                : activePropertyCallout.side === "bottom"
                  ? "translate(-50%, 0)"
                  : activePropertyCallout.side === "left"
                    ? "translate(-100%, -50%)"
                    : "translate(0, -50%)",
            zIndex: 18,
            pointerEvents: "none",
          }}
        >
          <div
            style={{
              position: "relative",
              backgroundColor: "#FFFFFF",
              border: "1px solid #000000",
              padding: 8,
            }}
          >
            <img
              src={`/assets/img/deeds/${activePropertyCallout.propertyId}.png`}
              alt={activePropertyCallout.propertyName}
              style={{
                display: "block",
                width: 180,
                height: "auto",
              }}
            />
          </div>
        </div>
      ) : null}
    </div>
  )
}
