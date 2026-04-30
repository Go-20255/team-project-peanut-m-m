"use client"

import { useEffect, useState } from "react"
import { getAvailableTokens, useFetchPlayersForSession } from "@/hooks/useGameAPI"
import { TOKEN_ICONS } from "@/utils/tokens"

interface TokenSelectorProps {
  sessionId: string
  playerId: number
  currentToken: number | null
  onTokenSelected: (token: number) => void
  isLoading?: boolean
}

export default function TokenSelector({
  sessionId,
  playerId,
  currentToken,
  onTokenSelected,
  isLoading = false,
}: TokenSelectorProps) {
  const [availableTokens, setAvailableTokens] = useState<number[]>([])
  const [loadingTokens, setLoadingTokens] = useState(true)

  const fetchPlayers = useFetchPlayersForSession()
  const players = fetchPlayers.data ?? []

  useEffect(() => {
    const fetchAvailable = async () => {
      try {
        setLoadingTokens(true)
        const available = await getAvailableTokens(players, sessionId)
        setAvailableTokens(available)
      } catch (err) {
        console.error("Failed to fetch available tokens:", err)
        setAvailableTokens([0, 1])
      } finally {
        setLoadingTokens(false)
      }
    }

    fetchAvailable()
  }, [sessionId])

  const isTokenAvailable = (token: number) => availableTokens.includes(token)

  return (
    <div className="w-full max-w-md mx-auto p-6">
      <div className="text-center mb-8">
        <h2 className="text-2xl font-bold mb-2" style={{ color: "#F76902" }}>
          Choose Your Icon
        </h2>
        <p className="text-gray-600">
          You were assigned: <strong>{TOKEN_ICONS[currentToken as keyof typeof TOKEN_ICONS].name}</strong>
        </p>
      </div>

      {loadingTokens ? (
        <div className="flex justify-center py-8">
          <p className="text-gray-500">Loading available icons...</p>
        </div>
      ) : (
        <div className="grid grid-cols-2 gap-4">
          {Object.entries(TOKEN_ICONS).map(([tokenId, tokenInfo]) => {
            const token = parseInt(tokenId)
            if (token > 1) return null

            const available = isTokenAvailable(token)
            const isSelected = currentToken === token

            return (
              <button
                key={tokenId}
                onClick={() => onTokenSelected(token)}
                disabled={!available || isLoading}
                className={`p-6 rounded-lg border-2 transition-all ${
                  isSelected
                    ? "border-green-500 bg-green-50"
                    : available
                      ? "border-gray-300 bg-white hover:border-orange-500"
                      : "border-gray-200 bg-gray-100 cursor-not-allowed opacity-50"
                }`}
              >
                <div className="flex flex-col items-center gap-2">
                  <img src={`/assets/img/icons/${tokenInfo.icon}`} alt={tokenInfo.name} className="w-16 h-16" />
                  <span className="font-semibold text-sm">{tokenInfo.name}</span>
                  {isSelected && <span className="text-xs text-green-600 font-bold">✓ Selected</span>}
                  {!available && <span className="text-xs text-gray-500 font-bold">Taken</span>}
                </div>
              </button>
            )
          })}
        </div>
      )}

      <div className="mt-6 text-center text-sm text-gray-600">
        {availableTokens.length === 1 ? (
          <p>Only one icon available. Select it to continue.</p>
        ) : (
          <p>Click an icon to select it for this game.</p>
        )}
      </div>
    </div>
  )
}
