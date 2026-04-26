"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import { useRouter } from "next/navigation";
import { storage } from "@/utils/storage";
import { useLiveGameUpdates, fetchPlayersForSession } from "@/hooks/useGameAPI";
import PlayerSidebar from "@/components/game/PlayerSidebar";
import GameBoard from "@/components/game/GameBoard";
import { Player, GameState } from "@/types";

export default function GamePage() {
  const router = useRouter();
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [playerId, setPlayerId] = useState<string | null>(null);
  const [playerName, setPlayerName] = useState<string | null>(null);
  const [players, setPlayers] = useState<Player[]>([]);
  const [gameState, setGameState] = useState<GameState | null>(null);
  const [isLoadingPlayers, setIsLoadingPlayers] = useState(true);
  const [currentPlayerTurn, setCurrentPlayerTurn] = useState<number | null>(null);
  const initialLoadRef = useRef(false);
  const pollingIntervalRef = useRef<NodeJS.Timeout | null>(null);

  const calculateCurrentPlayerTurn = (players: Player[], turnNumber: number) => {
    const orderedPlayers = players.filter((p) => p.player_order !== -1).sort((a, b) => a.player_order - b.player_order);
    if (orderedPlayers.length === 0) return null;
    const currentIndex = turnNumber % orderedPlayers.length;
    return orderedPlayers[currentIndex]?.id || null;
  };

  const updateTurnNumber = (newTurnNumber: number) => {
    setGameState((prev) => (prev ? { ...prev, turn_number: newTurnNumber } : null));
    const newTurn = calculateCurrentPlayerTurn(players, newTurnNumber);
    setCurrentPlayerTurn(newTurn);
    console.log("Current player turn updated:", newTurn);
  };

  const refreshGameData = useCallback(async (sid: string) => {
    try {
      const playersData = await fetchPlayersForSession(sid);
      
      if (playersData && playersData.length > 0) {
        setPlayers(playersData);
        
        if (gameState && gameState.turn_number !== undefined) {
          const currentTurn = calculateCurrentPlayerTurn(playersData, gameState.turn_number);
          setCurrentPlayerTurn(currentTurn);
          console.log("Current player turn (from polling):", currentTurn);
        }
      }
    } catch (err) {
      console.error("Failed to refresh game data:", err);
    }
  }, [gameState]);

  useEffect(() => {
    const storedSessionId = storage.getSessionId();
    const storedPlayerId = storage.getPlayerId();
    const storedPlayerName = storage.getPlayerName();

    if (!storedSessionId || !storedPlayerId || !storedPlayerName) {
      router.push("/");
      return;
    }

    setSessionId(storedSessionId);
    setPlayerId(storedPlayerId);
    setPlayerName(storedPlayerName);

    if (!initialLoadRef.current) {
      initialLoadRef.current = true;
      
      fetchPlayersForSession(storedSessionId)
        .then(setPlayers)
        .catch((err) => console.error("Failed to fetch initial players:", err))
        .finally(() => setIsLoadingPlayers(false));

      pollingIntervalRef.current = setInterval(
        () => refreshGameData(storedSessionId).catch((err) => console.error("Polling error:", err)),
        3000
      );
    }

    return () => {
      if (pollingIntervalRef.current) clearInterval(pollingIntervalRef.current);
    };
  }, [router, refreshGameData]);

  const handleUpdate = useCallback((update: any) => {
    if (update.type === "playersUpdate" && Array.isArray(update.data)) {
      console.log("Players updated via SSE:", update.data);
      setPlayers(update.data);
    } else if (update.type === "gameStateUpdate" && update.data?.turn_number !== undefined) {
      console.log("Game state updated via SSE:", update.data);
      updateTurnNumber(update.data.turn_number);
    } else if (update.type === "playerMove" && update.data) {
      console.log("Player moved:", update.data);
      setPlayers((prev) =>
        prev.map((p) =>
          p.id === update.data.player_id
            ? { ...p, position: update.data.new_position }
            : p
        )
      );
      if (update.data.turn_number !== undefined) {
        updateTurnNumber(update.data.turn_number);
      }
    } else if (update.data) {
      console.log(`Event received: ${update.type}`, update.data);
    }
  }, [players, updateTurnNumber]);

  useLiveGameUpdates(sessionId, playerId, playerName, handleUpdate);

  if (!sessionId || !playerId || !playerName) {
    return null;
  }

  return (
    <div className="w-full h-screen flex" style={{ backgroundColor: "#FFFFFF" }}>
      <div className="flex-1" style={{ flex: "4" }}>
        <GameBoard
          sessionId={sessionId}
          playerId={playerId}
          playerName={playerName}
          players={players}
          currentPlayerTurnId={currentPlayerTurn}
          gameState={gameState}
        />
      </div>

      <div style={{ flex: "1", borderLeft: "2px solid #D0D3D4" }}>
        <PlayerSidebar
          sessionId={sessionId}
          playerId={playerId}
          playerName={playerName}
          players={players}
          currentPlayerTurnId={currentPlayerTurn}
          gameState={gameState}
        />
      </div>
    </div>
  );
}
