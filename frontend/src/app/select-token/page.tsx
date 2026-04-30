"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import { storage } from "@/utils/storage"
import { useUpdatePlayerToken } from "@/hooks/useGameAPI"
import TokenSelector from "@/components/game/TokenSelector"

export default function SelectTokenPage() {
  const router = useRouter()
  const [sessionId, setSessionId] = useState<string | null>(null)
  const [playerId, setPlayerId] = useState<number | null>(null)
  const [playerName, setPlayerName] = useState<string | null>(null)
  const [currentToken, setCurrentToken] = useState<number | null>(null)
  const [selectedToken, setSelectedToken] = useState<number | null>(null)
  const updateToken = useUpdatePlayerToken()

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
    setCurrentToken(player?.piece_token) // Default based on players data
  }, [router])

  const handleTokenSelected = async (token: number) => {
    if (!sessionId || !playerId) return

    setSelectedToken(token)

    try {
      await updateToken.mutateAsync({
        playerId,
        sessionId,
        pieceToken: token,
      })

      router.push("/game")
    } catch (err) {
      console.error("Failed to update token:", err)
      setSelectedToken(null)
    }
  }

  if (!sessionId || !playerId || !playerName) {
    return <div>Loading...</div>
  }

  return (
    <div className="w-full min-h-screen flex items-center justify-center" style={{ backgroundColor: "#FFFFFF" }}>
      <TokenSelector
        sessionId={sessionId}
        playerId={playerId}
        currentToken={currentToken}
        onTokenSelected={handleTokenSelected}
        isLoading={updateToken.isPending}
      />
    </div>
  )
}
