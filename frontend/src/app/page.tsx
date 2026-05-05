"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { useCreateGame, useJoinGameByCode } from "@/hooks/useGameAPI"
import { storage } from "@/utils/storage"
import { useCreatePlayer } from "@/hooks/playerHooks"

export default function Home() {
  const router = useRouter()
  const [gameCode, setGameCode] = useState("")
  const [error, setError] = useState("")

  const createGame = useCreateGame()
  const joinGame = useJoinGameByCode()
  const createPlayer = useCreatePlayer()

  const handleCreate = async () => {
    setError("")

    try {
      console.log("Creating game...")
      const gameData = await createGame.mutateAsync()
      console.log("Game created:", gameData)

      console.log("Storing to localStorage and navigating...")
      storage.setSessionId(gameData.session_id)
      storage.setGameCode(gameData.code.toString())

      router.push("/select-player")
    } catch (err) {
      setError("Failed to create game. Please try again.")
      console.error("Create game error:", err)
    }
  }

  const handleJoin = async () => {
    setError("")

    if (!gameCode.trim()) {
      setError("Please enter a game code")
      return
    }

    try {
      console.log("Joining game with code:", gameCode)
      const sessionId = await joinGame.mutateAsync(gameCode)
      console.log("Joined game, session ID:", sessionId)

      console.log("Storing to localStorage and navigating...")
      storage.setSessionId(sessionId)
      storage.setGameCode(gameCode)

      router.push("/select-player")
    } catch (err) {
      setError("Failed to join game. Check your code and try again.")
      console.error("Join game error:", err)
    }
  }

  const isLoading = createGame.isPending || joinGame.isPending || createPlayer.isPending

  return (
    <div className="h-screen flex items-center justify-center" style={{ backgroundColor: "#FFFFFF" }}>
      <div className="w-full max-w-xl px-4">
        <div className="text-center mb-15">
          <h1 className="text-4xl font-bold text-rit-orange">
            Monopoly
          </h1>
        </div>

        <div className="space-y-4">
          <div className="w-full border-t-2 pt-4" style={{ borderColor: "#D0D3D4" }}>
            <div>
              <input
                type="text"
                placeholder="Enter game code"
                value={gameCode}
                onChange={(e) => setGameCode(e.target.value)}
                disabled={isLoading}
                className="w-full px-4 py-2 border-2"
                style={{
                  borderColor: "#D0D3D4",
                  color: "#000000",
                }}
              />
            </div>
            <button
              onClick={handleJoin}
              disabled={isLoading}
              className="w-full px-4 py-2 font-bold text-white transition-colors mt-2 bg-black disabled:bg-monopoly-disabled hover:bg-gray-800 cursor-pointer disabled:cursor-not-allowed"

            >
              Join Game
            </button>
          </div>

          {error && (
            <div className="px-4 py-2 text-center" style={{ color: "#F76902" }}>
              {error}
            </div>
          )}
          <button
            onClick={handleCreate}
            disabled={isLoading}
            className="w-full px-4 py-2 font-bold text-white transition-colors"
            style={{
              backgroundColor: isLoading ? "#A2AAAD" : "#F76902",
              color: "#FFFFFF",
              cursor: isLoading ? "not-allowed" : "pointer",
            }}
          >
            Create Game
          </button>
        </div>
      </div>
    </div>
  )
}
