'use client';

import React, { useEffect, useState } from 'react';
import { LoginScreen } from '@/components/auth/LoginScreen';
import { LobbyScreen } from '@/components/game/LobbyScreen';
import { GameScreen } from '@/components/game/GameScreen';
import { useAuth } from '@/hooks/useAuth';
import { storage } from '@/utils/storage';
import type { ScreenType } from '@/types';

type AppScreen = 'login' | 'lobby' | 'game';

interface SessionData {
  sessionId: string;
  playerId: string;
  playerName: string;
  joinCode?: string;
}

export default function Home() {
  const [screen, setScreen] = useState<AppScreen>('login');
  const [sessionData, setSessionData] = useState<SessionData | null>(null);
  const [isHydrated, setIsHydrated] = useState(false);
  const auth = useAuth();

  // Hydrate from storage on mount
  useEffect(() => {
    const sessionId = storage.getSessionId();
    const playerId = storage.getPlayerId();
    const playerName = storage.getPlayerName();

    if (sessionId && playerId && playerName) {
      setSessionData({
        sessionId,
        playerId,
        playerName,
      });
      setScreen('lobby');
    }

    setIsHydrated(true);
  }, []);

  const handleLoginSuccess = (sessionId: string) => {
    const playerName = storage.getPlayerName();
    const playerId = storage.getPlayerId();

    if (playerName && playerId) {
      setSessionData({
        sessionId,
        playerId,
        playerName,
        joinCode: extractJoinCode(sessionId),
      });
      setScreen('lobby');
    }
  };

  const handleStartGame = () => {
    // TODO: Implement game start logic with backend
    // - Validate player count
    // - Start game engine
    // - Initialize SSE connection
    if (sessionData) {
      setScreen('game');
    }
  };

  const handleLogout = () => {
    auth.logout();
    storage.clear();
    setSessionData(null);
    setScreen('login');
  };

  const handleLeaveGame = () => {
    // TODO: Implement proper game leave logic
    // - Notify backend
    // - Close SSE connection
    handleLogout();
  };

  // Helper function to extract or format join code
  // In a real app, this would come from the create game response
  const extractJoinCode = (sessionId: string): string => {
    // TODO: Store join code when creating game
    return sessionId.slice(0, 6).toUpperCase();
  };

  if (!isHydrated) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-rit-orange to-orange-600 flex items-center justify-center">
        <div className="animate-spin text-5xl drop-shadow-lg">🎲</div>
      </div>
    );
  }

  return (
    <main className="w-full">
      {screen === 'login' && (
        <LoginScreen onLoginSuccess={handleLoginSuccess} />
      )}

      {screen === 'lobby' && sessionData && (
        <LobbyScreen
          playerName={sessionData.playerName}
          sessionId={sessionData.sessionId}
          joinCode={sessionData.joinCode}
          onStartGame={handleStartGame}
          onLogout={handleLogout}
        />
      )}

      {screen === 'game' && sessionData && (
        <GameScreen
          sessionId={sessionData.sessionId}
          playerName={sessionData.playerName}
          onLeaveGame={handleLeaveGame}
        />
      )}
    </main>
  );
}
