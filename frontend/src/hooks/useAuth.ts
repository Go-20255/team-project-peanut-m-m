/**
 * Custom hook for authentication logic
 */

'use client';

import { useCallback } from 'react';
import { api } from '@/utils/api';
import { storage } from '@/utils/storage';

export interface UseAuthReturn {
  login: (playerName: string) => Promise<{ sessionId: string; playerId: string }>;
  logout: () => void;
  isLoggedIn: boolean;
  playerName: string | null;
  sessionId: string | null;
  playerId: string | null;
}

export function useAuth(): UseAuthReturn {
  const playerName = storage.getPlayerName();
  const sessionId = storage.getSessionId();
  const playerId = storage.getPlayerId();
  const isLoggedIn = !!(playerName && sessionId && playerId);

  const login = useCallback(
    async (playerName: string): Promise<{ sessionId: string; playerId: string }> => {
      if (!sessionId) {
        throw new Error('No active session');
      }

      const result = await api.addPlayer(sessionId, playerName);
      
      // Store authentication data
      storage.setPlayerName(playerName);
      storage.setPlayerId(result.player_id);

      return {
        sessionId: result.session_id,
        playerId: result.player_id,
      };
    },
    [sessionId]
  );

  const logout = useCallback(() => {
    storage.clear();
  }, []);

  return {
    login,
    logout,
    isLoggedIn,
    playerName,
    sessionId,
    playerId,
  };
}
