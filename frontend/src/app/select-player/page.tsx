"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import { fetchPlayersForSession, useCreatePlayer, useFetchPlayersForSession } from "@/hooks/useGameAPI"
import { Player } from "@/types"
import { storage } from "@/utils/storage"

export default function SelectPlayer() {
  const router = useRouter()
  const [playerName, setPlayerName] = useState("")
  const [gameCode, setGameCode] = useState("")
  const [error, setError] = useState("")
  const [createdCode, setCreatedCode] = useState<number | null>(null)

  const createPlayer = useCreatePlayer()
  const fetchPlayers = useFetchPlayersForSession()
  const players = fetchPlayers.data ?? []

  const joinGame = (p: Player) => {
    storage.setPlayerName(p.name)
    storage.setPlayerId(p.id.toString())
    router.push("/select-token")
  }

  return (
    <div className="flex flex-col gap-3">
      {players?.length > 0 ? (
        players.map((p) => (
            <button
            key={p.id}
            className="bg-gray-100 px-2 hover:cursor-pointer hover:bg-gray-300 "
            onClick={() => joinGame(p)}
          >{p.name}</button>
        ))
      ) : (
        <div>
          <p>no players exist for session</p>
        </div>
      )}
      {players?.length < 4 ? (
        <form>
          <button>Create Player</button>
        </form>
      ) : (
      <></>
      )}
    </div>
  )
}
