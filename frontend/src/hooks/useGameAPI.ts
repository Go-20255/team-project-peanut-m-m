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
    }): Promise<{ id: number; name: string; session_id: string; piece_token: number }> => {
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
    return flattenPlayerInfo(data || []);
  } catch (err) {
    console.error("Error fetching players:", err);
    return [];
  }
}

function flattenPlayerInfo(playerInfoList: any[]): any[] {
  return playerInfoList.map((playerInfo) => {
    if (playerInfo.player) {
      return playerInfo.player;
    }
    return playerInfo;
  });
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

    eventSource.addEventListener("InitialGameBoardDataEvent", (event: any) => {
      try {
        const data = JSON.parse(event.data);
        console.log("Received InitialGameBoardDataEvent:", data);
        const flattenedPlayers = flattenPlayerInfo(data.Players);
        onUpdateRef.current({ type: "playersUpdate", data: flattenedPlayers });
        onUpdateRef.current({ type: "gameStateUpdate", data: { turn_number: data.CurrentTurn } });
      } catch (e) {
        console.error("Failed to parse InitialGameBoardDataEvent:", e);
      }
    });

    eventSource.addEventListener("RollDiceEvent", (event: any) => {
      try {
        const data = JSON.parse(event.data);
        console.log("Received RollDiceEvent:", data);
        onUpdateRef.current({ type: "diceRoll", data });
      } catch (e) {
        console.error("Failed to parse RollDiceEvent:", e);
      }
    });

    eventSource.addEventListener("MovePlayerEvent", (event: any) => {
      try {
        const data = JSON.parse(event.data);
        console.log("Received MovePlayerEvent:", data);
        onUpdateRef.current({ type: "playerMove", data });
      } catch (e) {
        console.error("Failed to parse MovePlayerEvent:", e);
      }
    });

    eventSource.addEventListener("RentPaidEvent", (event: any) => {
      try {
        const data = JSON.parse(event.data);
        console.log("Received RentPaidEvent:", data);
        onUpdateRef.current({ type: "rentPaid", data });
      } catch (e) {
        console.error("Failed to parse RentPaidEvent:", e);
      }
    });

    eventSource.addEventListener("PropertyPurchased", (event: any) => {
      try {
        const data = JSON.parse(event.data);
        console.log("Received PropertyPurchased:", data);
        onUpdateRef.current({ type: "propertyPurchased", data });
      } catch (e) {
        console.error("Failed to parse PropertyPurchased:", e);
      }
    });

    eventSource.addEventListener("HousePurchased", (event: any) => {
      try {
        const data = JSON.parse(event.data);
        console.log("Received HousePurchased:", data);
        onUpdateRef.current({ type: "housePurchased", data });
      } catch (e) {
        console.error("Failed to parse HousePurchased:", e);
      }
    });

    eventSource.addEventListener("HotelPurchased", (event: any) => {
      try {
        const data = JSON.parse(event.data);
        console.log("Received HotelPurchased:", data);
        onUpdateRef.current({ type: "hotelPurchased", data });
      } catch (e) {
        console.error("Failed to parse HotelPurchased:", e);
      }
    });

    // HouseSold: Sent when a house is sold
    eventSource.addEventListener("HouseSold", (event: any) => {
      try {
        const data = JSON.parse(event.data);
        console.log("Received HouseSold:", data);
        onUpdateRef.current({ type: "houseSold", data });
      } catch (e) {
        console.error("Failed to parse HouseSold:", e);
      }
    });

    // HotelSold: Sent when a hotel is sold
    eventSource.addEventListener("HotelSold", (event: any) => {
      try {
        const data = JSON.parse(event.data);
        console.log("Received HotelSold:", data);
        onUpdateRef.current({ type: "hotelSold", data });
      } catch (e) {
        console.error("Failed to parse HotelSold:", e);
      }
    });

    // PropertyMortgaged: Sent when a property is mortgaged
    eventSource.addEventListener("PropertyMortgaged", (event: any) => {
      try {
        const data = JSON.parse(event.data);
        console.log("Received PropertyMortgaged:", data);
        onUpdateRef.current({ type: "propertyMortgaged", data });
      } catch (e) {
        console.error("Failed to parse PropertyMortgaged:", e);
      }
    });

    // PropertyUnmortgaged: Sent when a property is unmortgaged
    eventSource.addEventListener("PropertyUnmortgaged", (event: any) => {
      try {
        const data = JSON.parse(event.data);
        console.log("Received PropertyUnmortgaged:", data);
        onUpdateRef.current({ type: "propertyUnmortgaged", data });
      } catch (e) {
        console.error("Failed to parse PropertyUnmortgaged:", e);
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
