/**
 * Lobby screen - shown after login, allows join/create decisions
 * This is a duplicate screen pattern but shown after authentication
 * In a full implementation, you might consolidate with LoginScreen
 */

'use client';

import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';

interface LobbyScreenProps {
  playerName: string;
  sessionId: string;
  joinCode?: string;
  onStartGame: () => void;
  onLogout: () => void;
}

export function LobbyScreen({
  playerName,
  sessionId,
  joinCode,
  onStartGame,
  onLogout,
}: LobbyScreenProps) {
  const [copiedCode, setCopiedCode] = useState(false);

  const copyJoinCode = () => {
    if (joinCode) {
      navigator.clipboard.writeText(joinCode);
      setCopiedCode(true);
      setTimeout(() => setCopiedCode(false), 2000);
    }
  };

  // TODO: Implement actual game state fetching
  const isHosting = !!joinCode;
  const playerCount = 1; // Placeholder

  return (
    <div className="min-h-screen bg-gradient-to-br from-rit-orange via-orange-500 to-orange-600 flex flex-col items-center justify-center p-4">
      {/* Decorative background elements */}
      <div className="absolute top-10 left-10 w-32 h-32 bg-white rounded-full opacity-20 blur-3xl animate-float" />
      <div className="absolute bottom-10 right-10 w-40 h-40 bg-white rounded-full opacity-15 blur-3xl animate-float animation-delay-2000" />

      <div className="relative z-10 w-full max-w-md space-y-6">
        {/* Header */}
        <div className="text-center space-y-2">
          <h1 className="text-5xl font-bold text-white drop-shadow-lg">Game Lobby</h1>
          <p className="text-white text-xl font-semibold drop-shadow-md">Welcome, <span className="font-bold">{playerName}</span>!</p>
        </div>

        {/* Game Info Card */}
        <Card variant="elevated">
          <div className="space-y-4">
            <div>
              <p className="text-sm text-gray-700 font-bold uppercase">Your Status</p>
              <div className="mt-2 flex gap-2">
                <Badge variant="info">
                  {isHosting ? '🏠 Host' : '👤 Player'}
                </Badge>
                <Badge variant="success">
                  {playerCount} {playerCount === 1 ? 'Player' : 'Players'}
                </Badge>
              </div>
            </div>

            {isHosting && joinCode && (
              <div className="bg-orange-100 rounded-xl p-4 border-2 border-rit-gray-light">
                <p className="text-xs text-gray-700 font-bold uppercase mb-2">
                  Share this code with friends
                </p>
                <div className="flex items-center gap-3">
                  <code className="text-2xl font-bold text-rit-orange flex-1 text-center">
                    {joinCode}
                  </code>
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={copyJoinCode}
                  >
                    {copiedCode ? '✓' : '📋'}
                  </Button>
                </div>
              </div>
            )}

            <div className="border-t-2 border-rit-gray-light pt-4">
              <p className="text-sm text-gray-700 font-bold uppercase mb-3">
                Players in Game
              </p>
              <div className="space-y-2">
                <div className="flex items-center gap-2 p-3 bg-orange-100 rounded-lg">
                  <span className="text-2xl">👤</span>
                  <div>
                    <p className="font-semibold text-rit-charcoal">{playerName}</p>
                    <p className="text-xs text-gray-700">
                      {isHosting ? 'Host' : 'Player'}
                    </p>
                  </div>
                </div>

                {/* TODO: Display other players when they join */}
                {playerCount === 1 && (
                  <p className="text-center text-sm text-gray-600 italic">
                    Waiting for other players...
                  </p>
                )}
              </div>
            </div>
          </div>
        </Card>

        {/* Action Buttons */}
        <div className="space-y-3">
          <Button
            variant="accent"
            size="lg"
            onClick={onStartGame}
            className="w-full"
            disabled={playerCount < 2} // TODO: Update when players join
          >
            Start Game 🎮
          </Button>

          <Button
            variant="secondary"
            size="md"
            onClick={onLogout}
            className="w-full"
          >
            Leave Lobby
          </Button>
        </div>

        {/* Status message */}
        {playerCount < 2 && (
          <div className="bg-orange-100 border-2 border-rit-gray-light rounded-xl p-3 text-center">
            <p className="text-sm text-orange-800 font-medium">
              ⏳ Waiting for at least 2 players to start
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
