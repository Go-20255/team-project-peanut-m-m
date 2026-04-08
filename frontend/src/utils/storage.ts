/**
 * Local storage utilities for persisting game state
 */

const prefix = 'monopoly_';

export const storage = {
  setSessionId(sessionId: string): void {
    if (typeof window !== 'undefined') {
      localStorage.setItem(`${prefix}session_id`, sessionId);
    }
  },

  getSessionId(): string | null {
    if (typeof window !== 'undefined') {
      return localStorage.getItem(`${prefix}session_id`);
    }
    return null;
  },

  setPlayerId(playerId: string): void {
    if (typeof window !== 'undefined') {
      localStorage.setItem(`${prefix}player_id`, playerId);
    }
  },

  getPlayerId(): string | null {
    if (typeof window !== 'undefined') {
      return localStorage.getItem(`${prefix}player_id`);
    }
    return null;
  },

  setPlayerName(playerName: string): void {
    if (typeof window !== 'undefined') {
      localStorage.setItem(`${prefix}player_name`, playerName);
    }
  },

  getPlayerName(): string | null {
    if (typeof window !== 'undefined') {
      return localStorage.getItem(`${prefix}player_name`);
    }
    return null;
  },

  clear(): void {
    if (typeof window !== 'undefined') {
      localStorage.removeItem(`${prefix}session_id`);
      localStorage.removeItem(`${prefix}player_id`);
      localStorage.removeItem(`${prefix}player_name`);
    }
  },
};

export default storage;
