"use client"

import { useMemo } from "react"
import { useFetchPlayersForSession } from "@/hooks/useGameAPI"
import { TOKEN_ICONS, getTokenName } from "@/utils/tokens"

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
  const { data: players = [], isLoading: isLoadingPlayers } =
    useFetchPlayersForSession()

  // Tokens taken by *other* players (the current player's token shouldn't
  // disqualify itself from being shown as selectable).
  const takenTokens = useMemo(
    () =>
      new Set(
        players
          .filter((p) => p.id !== playerId)
          .map((p) => p.piece_token),
      ),
    [players, playerId],
  )

  const availableTokenIds = useMemo(
    () =>
      Object.keys(TOKEN_ICONS)
        .map(Number)
        .filter((id) => !takenTokens.has(id)),
    [takenTokens],
  )

  const isTokenAvailable = (token: number) => !takenTokens.has(token)

  return (
    <div className="w-full max-w-md mx-auto p-6">
      <div className="text-center mb-8">
        <h2 className="text-2xl font-bold mb-2" style={{ color: "#F76902" }}>
          Choose Your Icon
        </h2>
        {currentToken !== null && (
          <p className="text-gray-600">
            You were assigned: <strong>{getTokenName(currentToken)}</strong>
          </p>
        )}
      </div>

      {isLoadingPlayers ? (
        <div className="flex justify-center py-8">
          <p className="text-gray-500">Loading available icons...</p>
        </div>
      ) : (
        <div className="grid grid-cols-2 gap-4">
          {Object.entries(TOKEN_ICONS).map(([tokenId, tokenInfo]) => {
            const token = Number(tokenId)
            const available = isTokenAvailable(token)
            const isSelected = currentToken === token
            const disabled = !available || isLoading

            return (
              <button
                key={tokenId}
                type="button"
                onClick={() => onTokenSelected(token)}
                disabled={disabled}
                aria-pressed={isSelected}
                className={`p-6 rounded-lg border-2 transition-all ${
                  isSelected
                    ? "border-green-500 bg-green-50 hover:cursor-pointer"
                    : available
                      ? "border-gray-300 bg-white hover:border-orange-500 hover:cursor-pointer"
                      : "border-gray-200 bg-gray-100 cursor-not-allowed opacity-50"
                }`}
              >
                <div className="flex flex-col items-center gap-2">
                  <img
                    src={`/assets/img/icons/${tokenInfo.icon}`}
                    alt={tokenInfo.name}
                    className="w-16 h-16"
                  />
                  <span className="font-semibold text-sm">
                    {tokenInfo.name}
                  </span>
                  {isSelected && (
                    <span className="text-xs text-green-600 font-bold">
                      ✓ Selected
                    </span>
                  )}
                  {!available && (
                    <span className="text-xs text-gray-500 font-bold">
                      Taken
                    </span>
                  )}
                </div>
              </button>
            )
          })}
        </div>
      )}

      <div className="mt-6 text-center text-sm text-gray-600">
        {availableTokenIds.length === 1 ? (
          <p>Only one icon available. Select it to continue.</p>
        ) : (
          <p>Click an icon to select it for this game.</p>
        )}
      </div>
    </div>
  )
}
