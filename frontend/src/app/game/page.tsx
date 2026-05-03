"use client"

import { useEffect, useMemo, useState } from "react"
import { useRouter } from "next/navigation"
import { storage } from "@/utils/storage"
import PlayerSidebar from "@/components/game/PlayerSidebar"
import GameBoard from "@/components/game/GameBoard"
import { useLiveGameUpdates } from "@/hooks/liveUpdates"

export default function GamePage() {
  const router = useRouter()
  const [sessionId, setSessionId] = useState<string | null>(null)
  const [playerId, setPlayerId] = useState<number | null>(null)
  const [playerName, setPlayerName] = useState<string | null>(null)

  useEffect(() => {
    const player = storage.getPlayer()
    const storedSessionId = storage.getSessionId()

    if (!storedSessionId || !player?.id || !player?.name) {
      router.push("/")
      return
    }

    setSessionId(storedSessionId)
    setPlayerId(player.id)
    setPlayerName(player.name)
  }, [router])

  const gameState = useLiveGameUpdates(sessionId, playerId, playerName)

  const players = useMemo(
    () => gameState?.players.map((pi) => pi.player) ?? [],
    [gameState],
  )

  const currentPlayerTurnId = useMemo(() => {
    if (!gameState) return null
    const ordered = players
      .filter((p) => p.player_order !== -1)
      .sort((a, b) => a.player_order - b.player_order)
    if (ordered.length === 0) return null
    return ordered[gameState.current_turn % ordered.length]?.id ?? null
  }, [gameState, players])

  if (!sessionId || !playerId || !playerName) return null

  return (
    <div
      className="w-full h-screen flex"
      style={{ backgroundColor: "#FFFFFF" }}
    >
      <div style={{ flex: "6" }}>
        {gameState ? (
          <GameBoard
            sessionId={sessionId}
            playerId={playerId.toString()}
            playerName={playerName}
            currentPlayerTurnId={currentPlayerTurnId}
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
          currentPlayerTurnId={currentPlayerTurnId}
          gameState={gameState ?? undefined}
        />
      </div>
    </div>
  )
}
