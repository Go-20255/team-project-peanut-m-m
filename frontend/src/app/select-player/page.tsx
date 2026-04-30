"use client"

import { useMemo, useState } from "react"
import { useRouter } from "next/navigation"
import { useCreatePlayer, useFetchPlayersForSession } from "@/hooks/useGameAPI"
import { Player } from "@/types"
import { storage } from "@/utils/storage"
import { TOKEN_ICONS, getTokenIcon, getTokenName } from "@/utils/tokens"

const MAX_PLAYERS = 4

export default function SelectPlayer() {
  const router = useRouter()
  const [playerName, setPlayerName] = useState("")
  const [selectedToken, setSelectedToken] = useState<number | null>(null)
  const [error, setError] = useState("")

  const createPlayer = useCreatePlayer()
  const { data, isLoading } = useFetchPlayersForSession()

  // Normalize: API may return null when there are no players yet.
  const players: Player[] = data ?? []

  const canCreatePlayer = players.length < MAX_PLAYERS

  const takenTokens = useMemo(
    () => new Set(players.map((p) => p.piece_token)),
    [players],
  )

  const availableTokens = useMemo(
    () =>
      Object.keys(TOKEN_ICONS)
        .map(Number)
        .filter((id) => !takenTokens.has(id)),
    [takenTokens],
  )

  const hasSelectedToken = (p: Player) =>
    p.piece_token !== null &&
    p.piece_token !== undefined &&
    p.piece_token in TOKEN_ICONS

  const joinAsPlayer = (p: Player) => {
    storage.setPlayer(p)
    router.push("/game")
  }

  const handleCreatePlayer = async (e: React.FormEvent) => {
    e.preventDefault()
    setError("")

    const trimmed = playerName.trim()
    if (!trimmed) {
      setError("Please enter a name.")
      return
    }
    if (selectedToken === null) {
      setError("Please select an icon.")
      return
    }

    const sessionId = storage.getSessionId()
    if (!sessionId) {
      setError("No active session found.")
      return
    }

    try {
      const newPlayer = await createPlayer.mutateAsync({
        playerName: trimmed,
        sessionId,
        pieceToken: selectedToken,
      })
      storage.setPlayer(newPlayer)
      router.push("/select-token")
    } catch (err) {
      console.error(err)
      setError(err instanceof Error ? err.message : "Failed to create player.")
    }
  }

  return (
    <div className="w-full max-w-md mx-auto p-6">
      <div className="text-center mb-8">
        <h2 className="text-2xl font-bold mb-2" style={{ color: "#F76902" }}>
          Select Your Player
        </h2>
        <p className="text-gray-600">
          Rejoin as an existing player or create a new one.
        </p>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-8">
          <p className="text-gray-500">Loading players...</p>
        </div>
      ) : players.length > 0 ? (
        <div className="grid grid-cols-2 gap-4 mb-6">
          {players.map((p) => {
            const selected = hasSelectedToken(p)
            return (
              <button
                key={p.id}
                type="button"
                onClick={() => joinAsPlayer(p)}
                disabled={p.in_game}
                className="p-6 rounded-lg border-2 border-gray-300 bg-white hover:border-orange-500 hover:cursor-pointer transition-all disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <div className="flex flex-col items-center gap-2">
                  {selected ? (
                    <img
                      src={getTokenIcon(p.piece_token)}
                      alt={getTokenName(p.piece_token)}
                      className="w-16 h-16"
                    />
                  ) : (
                    <div className="w-16 h-16 flex items-center justify-center rounded-full bg-gray-100 border border-dashed border-gray-300">
                      <span className="text-xs text-gray-400">?</span>
                    </div>
                  )}
                  <span className="font-semibold text-sm">{p.name}</span>
                  <span className="text-xs text-gray-500">
                    {selected ? getTokenName(p.piece_token) : "No icon yet"}
                  </span>
                </div>
              </button>
            )
          })}
        </div>
      ) : (
        <div className="text-center py-8 mb-6">
          <p className="text-gray-500">No players exist for this session.</p>
        </div>
      )}

      {canCreatePlayer ? (
        <form
          onSubmit={handleCreatePlayer}
          className="border-t border-gray-200 pt-6 flex flex-col gap-4"
        >
          <div className="flex flex-col gap-2">
            <label
              htmlFor="player-name"
              className="text-sm font-semibold text-gray-700"
            >
              Create a new player
            </label>
            <input
              id="player-name"
              type="text"
              value={playerName}
              onChange={(e) => setPlayerName(e.target.value)}
              placeholder="Enter player name"
              disabled={createPlayer.isPending}
              className="px-3 py-2 rounded-lg border-2 border-gray-300 focus:border-orange-500 focus:outline-none disabled:opacity-50"
            />
          </div>

          <div className="flex flex-col gap-2">
            <span className="text-sm font-semibold text-gray-700">
              Choose an icon
            </span>
            {availableTokens.length === 0 ? (
              <p className="text-sm text-gray-500">
                No icons available. Wait for a slot to free up.
              </p>
            ) : (
              <div className="grid grid-cols-2 gap-3">
                {Object.entries(TOKEN_ICONS).map(([tokenId, tokenInfo]) => {
                  const token = Number(tokenId)
                  const available = !takenTokens.has(token)
                  const isSelected = selectedToken === token
                  const disabled =
                    !available || createPlayer.isPending

                  return (
                    <button
                      key={tokenId}
                      type="button"
                      onClick={() => setSelectedToken(token)}
                      disabled={disabled}
                      aria-pressed={isSelected}
                      className={`p-4 rounded-lg border-2 transition-all ${
                        isSelected
                          ? "border-green-500 bg-green-50"
                          : available
                            ? "border-gray-300 bg-white hover:border-orange-500 hover:cursor-pointer"
                            : "border-gray-200 bg-gray-100 cursor-not-allowed opacity-50"
                      }`}
                    >
                      <div className="flex flex-col items-center gap-1">
                        <img
                          src={`/assets/img/icons/${tokenInfo.icon}`}
                          alt={tokenInfo.name}
                          className="w-12 h-12"
                        />
                        <span className="font-semibold text-xs">
                          {tokenInfo.name}
                        </span>
                        {!available && (
                          <span className="text-[10px] text-gray-500 font-bold">
                            Taken
                          </span>
                        )}
                      </div>
                    </button>
                  )
                })}
              </div>
            )}
          </div>

          {error && <p className="text-sm text-red-600">{error}</p>}

          <button
            type="submit"
            disabled={
              createPlayer.isPending ||
              !playerName.trim() ||
              selectedToken === null
            }
            className="px-4 py-2 rounded-lg bg-orange-500 text-white font-semibold hover:bg-orange-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {createPlayer.isPending ? "Creating..." : "Create Player"}
          </button>
        </form>
      ) : (
        <div className="border-t border-gray-200 pt-6 text-center text-sm text-gray-600">
          <p>This game is full ({MAX_PLAYERS} players max).</p>
        </div>
      )}
    </div>
  )
}
