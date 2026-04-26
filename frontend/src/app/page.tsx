"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import {
  useCreateGame,
  useJoinGameByCode,
  useCreatePlayer,
} from "@/hooks/useGameAPI";
import { storage } from "@/utils/storage";

export default function Home() {
  const router = useRouter();
  const [playerName, setPlayerName] = useState("");
  const [gameCode, setGameCode] = useState("");
  const [error, setError] = useState("");
  const [createdCode, setCreatedCode] = useState<number | null>(null);

  const createGame = useCreateGame();
  const joinGame = useJoinGameByCode();
  const createPlayer = useCreatePlayer();

  const handleCreate = async () => {
    setError("");

    if (!playerName.trim()) {
      setError("Please enter a name");
      return;
    }

    try {
      console.log("Creating game...");
      // Step 1: Create game session
      const gameData = await createGame.mutateAsync();
      console.log("Game created:", gameData);

      console.log("Creating player:", playerName);
      // Step 2: Create player
      const playerData = await createPlayer.mutateAsync({
        playerName,
        sessionId: gameData.session_id,
      });
      console.log("Player created:", playerData);

      console.log("Storing to localStorage and navigating...");
      // Step 3: Store in localStorage
      storage.setSessionId(gameData.session_id);
      storage.setPlayerId(playerData.id.toString());
      storage.setPlayerName(playerName);
      storage.setGameCode(gameData.code.toString());

      // Step 4: Navigate to game
      router.push("/game");
    } catch (err) {
      setError("Failed to create game. Please try again.");
      console.error("Create game error:", err);
    }
  };

  const handleJoin = async () => {
    setError("");

    if (!playerName.trim()) {
      setError("Please enter a name");
      return;
    }

    if (!gameCode.trim()) {
      setError("Please enter a game code");
      return;
    }

    try {
      console.log("Joining game with code:", gameCode);
      // Step 1: Get session ID from code
      const sessionId = await joinGame.mutateAsync(gameCode);
      console.log("Joined game, session ID:", sessionId);

      console.log("Creating player:", playerName);
      // Step 2: Create player
      const playerData = await createPlayer.mutateAsync({
        playerName,
        sessionId,
      });
      console.log("Player created:", playerData);

      console.log("Storing to localStorage and navigating...");
      // Step 3: Store in localStorage
      storage.setSessionId(sessionId);
      storage.setPlayerId(playerData.id.toString());
      storage.setPlayerName(playerName);
      storage.setGameCode(gameCode);

      // Step 4: Navigate to game
      router.push("/game");
    } catch (err) {
      setError("Failed to join game. Check your code and try again.");
      console.error("Join game error:", err);
    }
  };

  const isLoading = createGame.isPending || joinGame.isPending || createPlayer.isPending;

  return (
    <div className="w-full h-screen flex items-center justify-center" style={{ backgroundColor: "#FFFFFF" }}>
      <div className="w-full max-w-md px-4">
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold" style={{ color: "#F76902" }}>
            Monopoly
          </h1>
        </div>

        <div className="space-y-4">
          {/* Player Name Input */}
          <div>
            <input
              type="text"
              placeholder="Enter your name"
              value={playerName}
              onChange={(e) => setPlayerName(e.target.value)}
              disabled={isLoading}
              className="w-full px-4 py-2 border-2"
              style={{
                borderColor: "#D0D3D4",
                color: "#000000",
              }}
            />
          </div>

          {/* Create Game Button */}
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

          {/* Join Game Section */}
          <div className="border-t-2 pt-4" style={{ borderColor: "#D0D3D4" }}>
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
              className="w-full px-4 py-2 font-bold text-white transition-colors mt-2"
              style={{
                backgroundColor: isLoading ? "#A2AAAD" : "#000000",
                color: "#FFFFFF",
                cursor: isLoading ? "not-allowed" : "pointer",
              }}
            >
              Join Game
            </button>
          </div>

          {/* Error Message */}
          {error && (
            <div className="px-4 py-2 text-center" style={{ color: "#F76902" }}>
              {error}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
