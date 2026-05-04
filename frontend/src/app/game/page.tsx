"use client"

import { useEffect, useMemo, useState } from "react"
import { useRouter } from "next/navigation"
import { storage } from "@/utils/storage"
import PlayerSidebar from "@/components/game/PlayerSidebar"
import GameBoard from "@/components/game/GameBoard"
import FinalRanksPage from "@/components/game/FinalRanksPage"
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

  const players = useMemo(() => gameState?.players.map((pi) => pi.player) ?? [], [gameState])
  const nonBankruptPlayers = useMemo(() => players.filter((player) => !player.bankrupt), [players])
  const showFinalRanksPage = useMemo(() => {
    if (!gameState) return false
    if (gameState.current_turn < 0) return false
    if (players.length === 0) return false

    return nonBankruptPlayers.length <= 1 && players.every((player) => player.rank > 0)
  }, [gameState, nonBankruptPlayers.length, players])

  const currentPlayerTurnId = useMemo(() => {
    if (!gameState) return null
    if (gameState.current_turn < 0 || gameState.players.length === 0) return null

    const currentPlayerIndex = gameState.current_turn % gameState.players.length
    for (let i = 0; i < gameState.players.length; i += 1) {
      const playerInfo = gameState.players[(currentPlayerIndex + i) % gameState.players.length]
      if (!playerInfo.player.bankrupt) {
        return playerInfo.player.id
      }
    }

    return null
  }, [gameState])

  if (!sessionId || !playerId || !playerName) return null

  if (showFinalRanksPage) {
    return <FinalRanksPage players={players} />
  }

  return (
    <div className="w-full h-screen flex overflow-hidden" style={{ backgroundColor: "#FFFFFF" }}>
      <div className="h-full overflow-hidden" style={{ flex: "7" }}>
        {gameState ? (
          <GameBoard
            sessionId={sessionId}
            playerId={playerId.toString()}
            playerName={playerName}
            currentPlayerTurnId={currentPlayerTurnId}
            gameState={gameState}
          />
        ) : (
          <div className="flex items-center justify-center h-full">Loading game state...</div>
        )}
      </div>

      <div className="h-full overflow-y-auto overflow-x-hidden" style={{ flex: "2", borderLeft: "2px solid #D0D3D4" }}>
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
