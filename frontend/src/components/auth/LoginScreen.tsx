/**
 * Login/Splash screen component
 */

'use client';

import React, { useState } from 'react';
import { Button } from '@/components/ui/Button';
import { TextInput } from '@/components/ui/TextInput';
import { Card } from '@/components/ui/Card';
import { useCreateGame, useJoinGame, useAddPlayer } from '@/hooks/useGameSession';

interface LoginScreenProps {
  onLoginSuccess: (sessionId: string) => void;
}

export function LoginScreen({ onLoginSuccess }: LoginScreenProps) {
  const [loginMode, setLoginMode] = useState<'splash' | 'create' | 'join' | 'success'>('splash');
  const [playerName, setPlayerName] = useState('');
  const [joinCode, setJoinCode] = useState('');
  const [successJoinCode, setSuccessJoinCode] = useState('');
  const [successSessionId, setSuccessSessionId] = useState('');
  const [error, setError] = useState('');

  const createGameMutation = useCreateGame();
  const joinGameMutation = useJoinGame();
  const addPlayerMutation = useAddPlayer();

  const isLoading =
    createGameMutation.isPending || joinGameMutation.isPending || addPlayerMutation.isPending;

  const handleCreateGame = async () => {
    if (!playerName.trim()) {
      setError('Please enter your name');
      return;
    }

    setError('');

    try {
      const gameData = await createGameMutation.mutateAsync();
      const playerData = await addPlayerMutation.mutateAsync({
        sessionId: gameData.session_id,
        playerName: playerName.trim(),
      });
      // Show success screen with join code
      setSuccessJoinCode(gameData.join_code);
      setSuccessSessionId(playerData.session_id);
      setLoginMode('success');
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'Failed to create game. Please try again.'
      );
    }
  };

  const handleJoinGame = async () => {
    if (!playerName.trim()) {
      setError('Please enter your name');
      return;
    }

    if (!joinCode.trim() || joinCode.length !== 6) {
      setError('Please enter a valid 6-digit code');
      return;
    }

    setError('');

    try {
      const gameData = await joinGameMutation.mutateAsync(joinCode);
      const playerData = await addPlayerMutation.mutateAsync({
        sessionId: gameData.session_id,
        playerName: playerName.trim(),
      });
      onLoginSuccess(playerData.session_id);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'Failed to join game. Please check the code.'
      );
    }
  };

  const handleContinueFromSuccess = () => {
    onLoginSuccess(successSessionId);
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-rit-orange via-orange-500 to-orange-600 flex items-center justify-center p-4">
      {/* Decorative background elements */}
      <div className="absolute top-10 left-10 w-32 h-32 bg-white rounded-full opacity-20 blur-3xl animate-float" />
      <div className="absolute bottom-10 right-10 w-40 h-40 bg-white rounded-full opacity-15 blur-3xl animate-float animation-delay-2000" />
      <div className="absolute top-1/2 right-1/4 w-24 h-24 bg-white rounded-full opacity-10 blur-3xl animate-float animation-delay-4000" />

      <div className="relative z-10 w-full max-w-md">
        {loginMode === 'splash' && (
          <div className="text-center space-y-6">
            {/* Logo/Title */}
            <div className="space-y-3 mb-8">
              <div className="text-6xl font-bold">
                🎲
              </div>
              <h1 className="text-5xl font-bold text-white drop-shadow-lg">Monopoly</h1>
              <p className="text-white text-xl font-semibold drop-shadow-md">Play online with friends!</p>
            </div>

            {/* Action Buttons */}
            <Card variant="elevated" className="space-y-4 bg-white bg-opacity-95 backdrop-blur-sm">
              <Button
                variant="primary"
                size="lg"
                onClick={() => setLoginMode('create')}
                className="w-full"
              >
                Create Game 🏠
              </Button>

              <div className="flex items-center gap-3">
                <div className="flex-1 h-px bg-gray-300" />
                <span className="text-gray-800 font-bold">or</span>
                <div className="flex-1 h-px bg-gray-300" />
              </div>

              <Button
                variant="secondary"
                size="lg"
                onClick={() => setLoginMode('join')}
                className="w-full"
              >
                Join Game 🔗
              </Button>
            </Card>
          </div>
        )}

        {loginMode === 'create' && (
          <Card variant="elevated" className="space-y-6 bg-white bg-opacity-95 backdrop-blur-sm">
            <div>
              <h2 className="text-3xl font-bold text-rit-orange mb-2">Create Game</h2>
              <p className="text-gray-700 font-semibold">You'll be the host!</p>
            </div>

            <TextInput
              label="Your Name"
              placeholder="Enter your name"
              value={playerName}
              onChange={(e) => {
                setPlayerName(e.target.value);
                setError('');
              }}
              disabled={isLoading}
              maxLength={30}
            />

            {error && (
              <div className="bg-red-100 border-2 border-red-300 text-red-800 px-4 py-3 rounded-xl">
                {error}
              </div>
            )}

            <div className="space-y-3">
              <Button
                variant="accent"
                size="lg"
                onClick={handleCreateGame}
                isLoading={isLoading}
                className="w-full"
              >
                Create & Continue 🎉
              </Button>

              <Button
                variant="secondary"
                size="md"
                onClick={() => {
                  setLoginMode('splash');
                  setPlayerName('');
                  setError('');
                }}
                disabled={isLoading}
                className="w-full"
              >
                Back
              </Button>
            </div>
          </Card>
        )}

        {loginMode === 'join' && (
          <Card variant="elevated" className="space-y-6 bg-white bg-opacity-95 backdrop-blur-sm">
            <div>
              <h2 className="text-3xl font-bold text-rit-orange mb-2">Join Game</h2>
              <p className="text-gray-700 font-semibold">Ask the host for the game code</p>
            </div>

            <TextInput
              label="Your Name"
              placeholder="Enter your name"
              value={playerName}
              onChange={(e) => {
                setPlayerName(e.target.value);
                setError('');
              }}
              disabled={isLoading}
              maxLength={30}
            />

            <TextInput
              label="Game Code"
              placeholder="000000"
              value={joinCode}
              onChange={(e) => {
                setJoinCode(e.target.value.replace(/\D/g, '').slice(0, 6));
                setError('');
              }}
              maxLength={6}
              disabled={isLoading}
            />

            {error && (
              <div className="bg-red-100 border-2 border-red-300 text-red-800 px-4 py-3 rounded-xl">
                {error}
              </div>
            )}

            <div className="space-y-3">
              <Button
                variant="accent"
                size="lg"
                onClick={handleJoinGame}
                isLoading={isLoading}
                className="w-full"
              >
                Join Game 🚀
              </Button>

              <Button
                variant="secondary"
                size="md"
                onClick={() => {
                  setLoginMode('splash');
                  setPlayerName('');
                  setJoinCode('');
                  setError('');
                }}
                disabled={isLoading}
                className="w-full"
              >
                Back
              </Button>
            </div>
          </Card>
        )}

        {loginMode === 'success' && (
          <Card variant="elevated" className="space-y-8 bg-white bg-opacity-95 backdrop-blur-sm text-center">
            <div>
              <div className="text-6xl mb-4">✨</div>
              <h2 className="text-3xl font-bold text-rit-orange mb-2">Game Created!</h2>
              <p className="text-gray-700 font-semibold">Share this code with friends to join</p>
            </div>

            <div className="bg-rit-orange bg-opacity-10 border-2 border-rit-orange rounded-lg p-6">
              <p className="text-gray-600 text-sm mb-2">JOIN CODE</p>
              <p className="text-5xl font-bold text-rit-orange font-mono tracking-wider">
                {successJoinCode}
              </p>
            </div>

            <div className="space-y-3">
              <Button
                variant="accent"
                size="lg"
                onClick={handleContinueFromSuccess}
                className="w-full"
              >
                Continue to Game 🎮
              </Button>

              <Button
                variant="secondary"
                size="md"
                onClick={() => {
                  navigator.clipboard.writeText(successJoinCode);
                }}
                className="w-full"
              >
                Copy Code 📋
              </Button>
            </div>
          </Card>
        )}
      </div>
    </div>
  );
}
