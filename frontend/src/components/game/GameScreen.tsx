/**
 * Main game screen - placeholder for actual game implementation
 */

'use client';

import React from 'react';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';

interface GameScreenProps {
  sessionId: string;
  playerName: string;
  onLeaveGame: () => void;
}

export function GameScreen({
  sessionId,
  playerName,
  onLeaveGame,
}: GameScreenProps) {
  return (
    <div className="min-h-screen bg-gradient-to-br from-rit-orange via-orange-500 to-orange-600 flex flex-col items-center justify-center p-4">
      <div className="relative z-10 w-full max-w-3xl">
        {/* Header */}
        <Card variant="elevated" className="mb-6">
          <div className="flex justify-between items-center">
            <div>
              <h1 className="text-3xl font-bold text-rit-orange">Game in Progress</h1>
              <p className="text-gray-700 font-semibold">Playing as {playerName}</p>
            </div>
            <Button
              variant="secondary"
              size="md"
              onClick={onLeaveGame}
            >
              Leave Game
            </Button>
          </div>
        </Card>

        {/* Game placeholder */}
        <Card variant="elevated" className="space-y-6">
          <div className="text-center py-12 space-y-4">
            <div className="text-6xl">🎲</div>
            <h2 className="text-2xl font-bold text-rit-orange">Game Board</h2>
            <p className="text-gray-700 text-lg">
              Game implementation coming soon!
            </p>
            <div className="bg-orange-100 border-2 border-rit-gray-light rounded-xl p-6 mt-6">
              <p className="text-sm text-rit-charcoal mb-4">
                This is a placeholder for the game board and game mechanics.
              </p>
              <p className="text-xs text-gray-700">
                Session ID: <code className="font-mono text-rit-charcoal">{sessionId}</code>
              </p>
            </div>
          </div>

          {/* TODO: Implement actual game UI */}
          {/* - Game board visualization */}
          {/* - Player positions and money */}
          {/* - Dice rolling */}
          {/* - Property management */}
          {/* - SSE event handling for real-time updates */}
        </Card>
      </div>
    </div>
  );
}
