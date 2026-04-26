"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import { useRouter } from "next/navigation";
import { storage } from "@/utils/storage";
import { useLiveGameUpdates, fetchPlayersForSession, fetchGameState } from "@/hooks/useGameAPI";
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

  // Calculate who's turn it is based on turn_number and player_order
  const calculateCurrentPlayerTurn = (players: Player[], turnNumber: number) => {
    const orderedPlayers = players.filter((p) => p.player_order !== -1).sort((a, b) => a.player_order - b.player_order);
    if (orderedPlayers.length === 0) return null;
    const currentIndex = turnNumber % orderedPlayers.length;
    return orderedPlayers[currentIndex]?.id || null;
  };

  // Fetch game state and players data
  const refreshGameData = useCallback(async (sid: string) => {
    try {
      const gameStateData = await fetchGameState(sid);
      const playersData = await fetchPlayersForSession(sid);
      
      if (gameStateData) {
        setGameState(gameStateData);
        console.log("Game state updated - turn_number:", gameStateData.turn_number);
      }
      
      if (playersData && playersData.length > 0) {
        setPlayers(playersData);
        
        // Calculate whose turn it is
        if (gameStateData && gameStateData.turn_number !== undefined) {
          const currentTurn = calculateCurrentPlayerTurn(playersData, gameStateData.turn_number);
          setCurrentPlayerTurn(currentTurn);
          console.log("Current player turn:", currentTurn);
        }
      }
    } catch (err) {
      console.error("Failed to refresh game data:", err);
    }
  }, []);

  useEffect(() => {
    // Get stored data
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

    // Fetch initial data
    if (!initialLoadRef.current) {
      initialLoadRef.current = true;
      refreshGameData(storedSessionId)
        .then(() => {
          setIsLoadingPlayers(false);
        })
        .catch((err) => {
          console.error("Failed to fetch initial game data:", err);
          setIsLoadingPlayers(false);
        });

      // Set up periodic polling every 3 seconds as a fallback to SSE
      pollingIntervalRef.current = setInterval(() => {
        refreshGameData(storedSessionId).catch((err) => {
          console.error("Polling error:", err);
        });
      }, 3000);
    }

    return () => {
      if (pollingIntervalRef.current) {
        clearInterval(pollingIntervalRef.current);
      }
    };
  }, [router, refreshGameData]);

  // Memoize the update handler to prevent SSE reconnection on every render
  const handleUpdate = useCallback((update: any) => {
    if (update.type === "playersUpdate" && Array.isArray(update.data)) {
      console.log("Players updated via SSE:", update.data);
      setPlayers(update.data);
    } else if (update.type === "playerJoined" && update.data) {
      // Refresh full player list when someone joins
      console.log("Player joined:", update.data);
      if (sessionId) {
        refreshGameData(sessionId);
      }
    } else if (update.type === "playerMove" && update.data) {
      // Update player position
      console.log("Player moved:", update.data);
      setPlayers((prev) =>
        prev.map((p) =>
          p.id === update.data.player_id
            ? { ...p, position: update.data.new_position }
            : p
        )
      );
      
      // Update turn number if provided
      if (update.data.turn_number !== undefined) {
        setGameState((prev) =>
          prev ? { ...prev, turn_number: update.data.turn_number } : null
        );
        const currentTurn = calculateCurrentPlayerTurn(players, update.data.turn_number);
        setCurrentPlayerTurn(currentTurn);
      }
    } else if (update.type === "diceRoll" && update.data) {
      // Dice roll happened - trigger a refresh
      console.log("Dice roll received:", update.data);
      if (sessionId) {
        refreshGameData(sessionId);
      }
    } else if (update.type === "gameStateUpdate" && update.data) {
      // Game state updated - refresh all data
      console.log("Game state updated via SSE:", update.data);
      if (sessionId) {
        refreshGameData(sessionId);
      }
    }
  }, [sessionId, players, refreshGameData]);

  // Listen to live game updates
  useLiveGameUpdates(
    sessionId,
    playerId,
    playerName,
    handleUpdate,
    true
  );

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
