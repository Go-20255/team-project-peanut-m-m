"use client"

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import {
    fetchPlayersForSession,
  useCreatePlayer,
  useFetchPlayersForSession,
} from "@/hooks/useGameAPI";
import {Player} from "@/types"
import { storage } from "@/utils/storage";


export default function SelectPlayer() {
  const router = useRouter();
  const [playerName, setPlayerName] = useState("");
  const [gameCode, setGameCode] = useState("");
  const [error, setError] = useState("");
  const [createdCode, setCreatedCode] = useState<number | null>(null);
  const [players, setPlayers] = useState<Player[]>([])

  const createPlayer = useCreatePlayer();
  const fetchPlayers = useFetchPlayersForSession()
  useEffect(() => {
    if (fetchPlayers.data != undefined) {
      setPlayers(fetchPlayers.data)
    }
  })


  return (
  <div>
    {players?.length > 0 ? (
      players.map((p) => (
        <div className="flex flex-col text-center ">
          <button>{p.name}</button>
        </div>
      ))
    ) : (
    <div>
          <p>no players exist for session</p>
    </div>
    )
    }
  </div>
  )
}
