/**
 * Custom hook for managing game session creation and joining
 */

'use client';

import { useCallback } from 'react';
import { api } from '@/utils/api';
import { storage } from '@/utils/storage';

export interface UseGameSessionReturn {
  createGame: () => Promise<{ sessionId: string; joinCode: string }>;
  joinGame: (joinCode: string) => Promise<string>;
}

export function useGameSession(): UseGameSessionReturn {
  const createGame = useCallback(async (): Promise<{ sessionId: string; joinCode: string }> => {
    const result = await api.createGame();
    storage.setSessionId(result.session_id);
    return {
      sessionId: result.session_id,
      joinCode: result.join_code,
    };
  }, []);

  const joinGame = useCallback(async (joinCode: string): Promise<string> => {
    const result = await api.resolveJoinCode(joinCode);
    storage.setSessionId(result.session_id);
    return result.session_id;
  }, []);

  return {
    createGame,
    joinGame,
  };
}
