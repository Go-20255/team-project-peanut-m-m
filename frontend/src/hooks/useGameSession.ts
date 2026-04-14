/**
 * @module Provides hooks for managing game session creation, joining, and player management
 */
'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { storage } from '@/utils/storage';
import { useRouter } from 'next/navigation';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:9876';

/**
 * @returns mutation function that can be executed to create a new game session
 */
export function useCreateGame() {
  const queryClient = useQueryClient();
  const router = useRouter();

  return useMutation({
    mutationFn: async (): Promise<{ session_id: string; join_code: string }> => {
      const response = await fetch(`${API_BASE_URL}/api/game`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
      });

      if (!response.ok) {
        if (response.status === 401) {
          router.push('/login');
        }
        const errorText = await response.text();
        throw new Error(`Failed to create game: ${errorText}`);
      }

      // Backend returns plain text with format: "session_id:join_code"
      const data = await response.text();
      const [session_id, join_code] = data.split(':');
      return { session_id, join_code };
    },
    onSuccess: (data) => {
      storage.setSessionId(data.session_id);
      queryClient.invalidateQueries({ queryKey: ['gameSession'] });
    },
  });
}

/**
 * @returns mutation function that can be executed to join an existing game session
 */
export function useJoinGame() {
  const queryClient = useQueryClient();
  const router = useRouter();

  return useMutation({
    mutationFn: async (joinCode: string): Promise<{ session_id: string }> => {
      const response = await fetch(`${API_BASE_URL}/api/game/join`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({ join_code: joinCode }),
      });

      if (!response.ok) {
        if (response.status === 401) {
          router.push('/login');
        }
        const errorText = await response.text();
        throw new Error(`Failed to resolve join code: ${errorText}`);
      }

      const session_id = await response.text();
      return { session_id };
    },
    onSuccess: (data) => {
      storage.setSessionId(data.session_id);
      queryClient.invalidateQueries({ queryKey: ['gameSession'] });
    },
  });
}

/**
 * @returns mutation function that can be executed to add a player to a game session
 */
export function useAddPlayer() {
  const queryClient = useQueryClient();
  const router = useRouter();

  return useMutation({
    mutationFn: async ({
      sessionId,
      playerName,
    }: {
      sessionId: string;
      playerName: string;
    }): Promise<{ player_id: string; name: string; session_id: string }> => {
      const response = await fetch(`${API_BASE_URL}/api/player`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({
          session_id: sessionId,
          name: playerName,
        }),
      });

      if (!response.ok) {
        if (response.status === 401) {
          router.push('/login');
        }
        const errorText = await response.text();
        throw new Error(`Failed to add player: ${errorText}`);
      }

      const player_id = await response.text();
      return {
        player_id,
        name: playerName,
        session_id: sessionId,
      };
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['gameSession', 'players'] });
    },
  });
}
