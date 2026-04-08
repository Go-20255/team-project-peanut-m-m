/**
 * API utility functions for communicating with the backend
 */

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:9876';

export const api = {
  /**
   * Health check endpoint
   */
  async healthCheck(): Promise<boolean> {
    try {
      const response = await fetch(`${API_BASE_URL}/api/health`);
      return response.ok;
    } catch {
      return false;
    }
  },

  /**
   * Create a new game session
   * @returns session_id and join_code
   */
  async createGame(): Promise<{ session_id: string; join_code: string }> {
    const response = await fetch(`${API_BASE_URL}/api/game`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Failed to create game: ${errorText}`);
    }

    const data = await response.text();
    // Backend returns plain text with format: "session_id:join_code"
    const [session_id, join_code] = data.split(':');
    return { session_id, join_code };
  },

  /**
   * Resolve a join code to a session ID
   * @param joinCode 6-digit code
   * @returns session_id
   */
  async resolveJoinCode(joinCode: string): Promise<{ session_id: string }> {
    const response = await fetch(`${API_BASE_URL}/api/game/join`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ join_code: joinCode }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Failed to resolve join code: ${errorText}`);
    }

    const session_id = await response.text();
    return { session_id };
  },

  /**
   * Add a player to a game session
   * @param sessionId The game session ID
   * @param playerName The player's name
   * @returns player object with id
   */
  async addPlayer(
    sessionId: string,
    playerName: string
  ): Promise<{ player_id: string; name: string; session_id: string }> {
    const response = await fetch(`${API_BASE_URL}/api/player`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        session_id: sessionId,
        name: playerName,
      }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Failed to add player: ${errorText}`);
    }

    const data = await response.text();
    // Backend returns plain text - parse accordingly
    // Expected format might be "player_id:name:session_id" or similar
    // For now, we'll create a response object
    const player_id = data;
    return {
      player_id,
      name: playerName,
      session_id: sessionId,
    };
  },

  /**
   * Subscribe to game updates via Server-Sent Events
   * @param sessionId The game session ID
   * @param onMessage Callback for message events
   * @returns cleanup function to close connection
   */
  subscribeToGameUpdates(
    sessionId: string,
    onMessage: (event: any) => void,
    onError?: (error: Error) => void
  ): () => void {
    const eventSource = new EventSource(
      `${API_BASE_URL}/api/game/join/live?session_id=${sessionId}`
    );

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        onMessage(data);
      } catch (err) {
        console.error('Failed to parse SSE message:', err);
      }
    };

    eventSource.onerror = () => {
      if (onError) {
        onError(new Error('SSE connection error'));
      }
      eventSource.close();
    };

    return () => {
      eventSource.close();
    };
  },
};

export default api;
