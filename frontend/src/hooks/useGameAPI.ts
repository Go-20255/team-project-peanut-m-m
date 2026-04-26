"use client";

import { useEffect, useRef } from "react";
import { useMutation } from "@tanstack/react-query";
import { useRouter } from "next/navigation";

const API_URL = process.env.NEXT_PUBLIC_API_URL;

// Health check to verify backend is accessible
export async function checkBackendHealth(): Promise<boolean> {
  try {
    const res = await fetch(`${API_URL}/api/health`, {
      method: "GET",
      credentials: "include",
    });
    return res.ok;
  } catch (err) {
    console.error("Backend health check failed:", err);
    return false;
  }
}

/**
 * Create a new game session
 */
export function useCreateGame() {
  const router = useRouter();
  return useMutation({
    mutationFn: async (): Promise<{ session_id: string; code: number }> => {
      // Check if backend is accessible
      const isHealthy = await checkBackendHealth();
      if (!isHealthy) {
        throw new Error(
          `Backend not accessible at ${API_URL}. Make sure the backend server is running on port 9876.`
        );
      }

      const res = await fetch(`${API_URL}/api/game`, {
        method: "POST",
        credentials: "include",
      });
      if (!res.ok) {
        const errorText = await res.text();
        throw new Error(`Failed to create game: ${res.status} ${errorText}`);
      }
      return res.json();
    },
  });
}

/**
 * Join game with a code and get session ID
 */
export function useJoinGameByCode() {
  const router = useRouter();
  return useMutation({
    mutationFn: async (code: string): Promise<string> => {
      const res = await fetch(`${API_URL}/api/game/join?code=${code}`, {
        method: "POST",
        credentials: "include",
      });
      if (!res.ok) {
        throw new Error("Failed to join game");
      }
      return res.text();
    },
  });
}

/**
 * Create a player in a game session
 */
export function useCreatePlayer() {
  const router = useRouter();
  return useMutation({
    mutationFn: async ({
      playerName,
      sessionId,
    }: {
      playerName: string;
      sessionId: string;
    }): Promise<{ id: number; name: string; session_id: string }> => {
      console.log("Creating player:", playerName, "in session:", sessionId);
      
      const formData = new FormData();
      formData.append("player_name", playerName);
      formData.append("session_id", sessionId);

      const res = await fetch(`${API_URL}/api/player`, {
        method: "POST",
        credentials: "include",
        body: formData,
      });
      
      console.log("Create player response status:", res.status);
      
      if (!res.ok) {
        const errorText = await res.text();
        console.error("Create player error:", res.status, errorText);
        throw new Error(`Failed to create player: ${res.status} ${errorText}`);
      }
      
      const data = await res.json();
      console.log("Player created successfully:", data);
      return data;
    },
  });
}

/**
 * Fetch game state (including turn information)
 */
export async function fetchGameState(sessionId: string): Promise<any> {
  try {
    console.log("Fetching game state for session:", sessionId);
    const res = await fetch(`${API_URL}/api/game?session_id=${sessionId}`, {
      method: "GET",
      credentials: "include",
    });
    console.log("Game state fetch response status:", res.status);
    
    if (!res.ok) {
      const errorText = await res.text();
      console.error("Game state fetch error:", res.status, errorText);
      throw new Error(`Failed to fetch game state: ${res.status} ${errorText}`);
    }
    
    const data = await res.json();
    console.log("Game state fetched successfully:", data);
    return data;
  } catch (err) {
    console.error("Error fetching game state:", err);
    return null;
  }
}

/**
 * Fetch all players in a game session
 */
export async function fetchPlayersForSession(sessionId: string): Promise<any[]> {
  try {
    console.log("Fetching players for session:", sessionId);
    const res = await fetch(`${API_URL}/api/game/players?session_id=${sessionId}`, {
      method: "GET",
      credentials: "include",
    });
    console.log("Players fetch response status:", res.status);
    
    if (!res.ok) {
      const errorText = await res.text();
      console.error("Players fetch error:", res.status, errorText);
      throw new Error(`Failed to fetch players: ${res.status} ${errorText}`);
    }
    
    const data = await res.json();
    console.log("Players fetched successfully:", data);
    return data || [];
  } catch (err) {
    console.error("Error fetching players:", err);
    return [];
  }
}

/**
 * Connect to live game updates via SSE and set up event listeners
 */
export function useLiveGameUpdates(
  sessionId: string | null,
  playerId: string | null,
  playerName: string | null,
  onUpdate: (data: any) => void,
  enabled = true
) {
  const onUpdateRef = useRef(onUpdate);

  // Update the ref when onUpdate changes, but don't re-establish the connection
  useEffect(() => {
    onUpdateRef.current = onUpdate;
  }, [onUpdate]);

  useEffect(() => {
    if (!enabled || !sessionId || !playerId || !playerName) {
      return;
    }

    const params = new URLSearchParams({
      session_id: sessionId,
      player_id: playerId,
      player_name: playerName,
    });

    const eventSource = new EventSource(
      `${API_URL}/api/game/join/live?${params}`
    );

    console.log("SSE connection established for session:", sessionId, "player:", playerId);

    eventSource.addEventListener("gameStateUpdate", (event: any) => {
      try {
        const data = JSON.parse(event.data);
        onUpdateRef.current({ type: "gameStateUpdate", data });
      } catch (e) {
        console.error("Failed to parse gameStateUpdate:", e);
      }
    });

    eventSource.addEventListener("playerJoined", (event: any) => {
      try {
        const data = JSON.parse(event.data);
        console.log("Received playerJoined event:", data);
        onUpdateRef.current({ type: "playerJoined", data });
      } catch (e) {
        console.error("Failed to parse playerJoined:", e);
      }
    });

    eventSource.addEventListener("playersUpdate", (event: any) => {
      try {
        const data = JSON.parse(event.data);
        console.log("Received playersUpdate event:", data);
        onUpdateRef.current({ type: "playersUpdate", data });
      } catch (e) {
        console.error("Failed to parse playersUpdate:", e);
      }
    });

    eventSource.addEventListener("diceRoll", (event: any) => {
      try {
        const data = JSON.parse(event.data);
        onUpdateRef.current({ type: "diceRoll", data });
      } catch (e) {
        console.error("Failed to parse diceRoll:", e);
      }
    });

    eventSource.addEventListener("playerMove", (event: any) => {
      try {
        const data = JSON.parse(event.data);
        onUpdateRef.current({ type: "playerMove", data });
      } catch (e) {
        console.error("Failed to parse playerMove:", e);
      }
    });

    eventSource.onerror = (error) => {
      console.error("SSE connection error:", error);
      eventSource.close();
    };

    return () => {
      console.log("Closing SSE connection for session:", sessionId);
      eventSource.close();
    };
  }, [sessionId, playerId, playerName, enabled]);
}

/**
 * Roll dice
 */
export function useRollDice() {
  return useMutation({
    mutationFn: async ({
      playerId,
      sessionId,
    }: {
      playerId: number;
      sessionId: string;
    }) => {
      const formData = new FormData();
      formData.append("player_id", playerId.toString());
      formData.append("session_id", sessionId);

      const res = await fetch(`${API_URL}/api/game/roll`, {
        method: "POST",
        credentials: "include",
        body: formData,
      });
      if (!res.ok) {
        throw new Error("Failed to roll dice");
      }
      return res.json();
    },
  });
}

/**
 * Move player
 */
export function useMovePlayer() {
  return useMutation({
    mutationFn: async ({
      playerId,
      sessionId,
    }: {
      playerId: number;
      sessionId: string;
    }) => {
      const formData = new FormData();
      formData.append("player_id", playerId.toString());
      formData.append("session_id", sessionId);

      const res = await fetch(`${API_URL}/api/game/move`, {
        method: "POST",
        credentials: "include",
        body: formData,
      });
      if (!res.ok) {
        throw new Error("Failed to move player");
      }
      return res.json();
    },
  });
}

/**
 * Purchase property
 */
export function usePurchaseProperty() {
  return useMutation({
    mutationFn: async ({
      playerId,
      sessionId,
      propertyId,
    }: {
      playerId: number;
      sessionId: string;
      propertyId: number;
    }) => {
      const formData = new FormData();
      formData.append("player_id", playerId.toString());
      formData.append("session_id", sessionId);
      formData.append("property_id", propertyId.toString());

      const res = await fetch(`${API_URL}/api/game/property`, {
        method: "POST",
        credentials: "include",
        body: formData,
      });
      if (!res.ok) {
        throw new Error("Failed to purchase property");
      }
      return res.json();
    },
  });
}
