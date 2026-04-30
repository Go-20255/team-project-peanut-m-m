"use client"

import { useEffect, useState, useCallback, useRef } from "react"
import { useRouter } from "next/navigation"
import { storage } from "@/utils/storage"
import PlayerSidebar from "@/components/game/PlayerSidebar"
import GameBoard from "@/components/game/GameBoard"
import { Player, GameState } from "@/types"
import { fetchPlayersForSession } from "@/hooks/playerHooks"
import { useLiveGameUpdates } from "@/hooks/liveUpdatesHooks"

export default function GamePage() {
  const router = useRouter()
  const [sessionId, setSessionId] = useState<string | null>(null)
  const [playerId, setPlayerId] = useState<number | null>(null)
  const [playerName, setPlayerName] = useState<string | null>(null)
  const [players, setPlayers] = useState<Player[]>([])
  const [gameState, setGameState] = useState<GameState>()
  const [isLoadingPlayers, setIsLoadingPlayers] = useState(true)
  const [currentPlayerTurn, setCurrentPlayerTurn] = useState<number | null>(null)

  useEffect(() => {
    const player = storage.getPlayer()
    const storedSessionId = storage.getSessionId()
    const storedPlayerId = player?.id
    const storedPlayerName = player?.name

    if (!storedSessionId || !storedPlayerId || !storedPlayerName) {
      router.push("/")
      return
    }

    setSessionId(storedSessionId)
    setPlayerId(storedPlayerId)
    setPlayerName(storedPlayerName)
  }, [router])

  const calculateCurrentPlayerTurn = (
    players: Player[],
    turnNumber: number,
  ) => {
    const orderedPlayers = players
      .filter((p) => p.player_order !== -1)
      .sort((a, b) => a.player_order - b.player_order)
    if (orderedPlayers.length === 0) return null
    const currentIndex = turnNumber % orderedPlayers.length
    return orderedPlayers[currentIndex]?.id || null
  }

  const handleUpdate = useCallback((update: any) => {
    if (update.type === "initialGameState" && update.data) {
      console.log("Initial game state received via SSE:", update.data)
      const { current_turn, tiles, players: playerInfos } = update.data

      // Update full game state (PlayerInfo[])
      setGameState({
        current_turn,
        tiles,
        players: playerInfos,
      })

      // Keep flat Player[] in sync for sidebar/turn calculation
      const flatPlayers: Player[] = playerInfos.map(
        (pi: any) => pi.player,
      )
      setPlayers(flatPlayers)
      setCurrentPlayerTurn(
        calculateCurrentPlayerTurn(flatPlayers, current_turn),
      )
      setIsLoadingPlayers(false)
    } else if (
      update.type === "playerMove" &&
      update.data
    ) {
      console.log("Player moved:", update.data)
      setPlayers((prev) =>
        prev.map((p) =>
          p.id === update.data.player_id
            ? { ...p, position: update.data.new_position }
            : p,
        ),
      )
      setGameState((prev) => {
        if (!prev) return prev
        return {
          ...prev,
          players: prev.players.map((pi) =>
            pi.player.id === update.data.player_id
              ? {
                  ...pi,
                  player: {
                    ...pi.player,
                    position: update.data.new_position,
                  },
                }
              : pi,
          ),
          current_turn:
            update.data.turn_number ?? prev.current_turn,
        }
      })
      if (update.data.turn_number !== undefined) {
        setCurrentPlayerTurn(
          calculateCurrentPlayerTurn(
            players,
            update.data.turn_number,
          ),
        )
      }
    } else if (update.data) {
      console.log(`Event received: ${update.type}`, update.data)
    }
  }, [players])

  useLiveGameUpdates(sessionId, playerId, playerName, handleUpdate)

  if (!sessionId || !playerId || !playerName) {
    return null
  }

  return (
    <div className="w-full h-screen flex" style={{ backgroundColor: "#FFFFFF" }}>
      <div className="flex-1" style={{ flex: "4" }}>
        {gameState ? (
          <GameBoard
            sessionId={sessionId}
            playerId={playerId.toString()}
            playerName={playerName}
            currentPlayerTurnId={currentPlayerTurn}
            gameState={gameState}
          />
        ) : (
          <div className="flex items-center justify-center h-full">
            Loading game state...
          </div>
        )}
      </div>

      <div style={{ flex: "1", borderLeft: "2px solid #D0D3D4" }}>
        <PlayerSidebar
          sessionId={sessionId}
          playerId={playerId.toString()}
          playerName={playerName}
          players={players}
          currentPlayerTurnId={currentPlayerTurn}
          gameState={gameState}
        />
      </div>
    </div>
  )
}
