/**
 * Application-wide type definitions
 */

export interface GameSession {
  session_id: string;
  join_code: string;
  state: string;
  created_at: string;
  updated_at: string;
}

export interface Player {
  id: string;
  name: string;
  session_id: string;
  position: number;
  money: number;
  in_jail: boolean;
}

export interface LoginCredentials {
  playerName: string;
  gameCode?: string;
}

export interface AuthState {
  isLoggedIn: boolean;
  playerName: string | null;
  sessionId: string | null;
  playerId: string | null;
}

export interface GameAction {
  type: 'ROLL_DICE' | 'END_TURN' | 'BUY_PROPERTY' | 'PAY_RENT' | 'JOIN_GAME' | 'CREATE_GAME';
  payload?: any;
}

export type ScreenType = 'login' | 'lobby';
