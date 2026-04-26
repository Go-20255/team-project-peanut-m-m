package util

import (
    "context"
    internaldbplayers "monopoly-backend/internal/db/player"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/rs/zerolog"
)

// PLAYER_TOKENS defines the available piece tokens for players
// Tokens are numbered 0-7 representing the 8 classic Monopoly pieces:
// 0: Top Hat, 1: Boot, 2: Thimble, 3: Iron, 4: Race Car, 5: Battleship, 6: Wheelbarrow, 7: Scottie Dog
const (
    TOTAL_PLAYER_TOKENS = 8
)

// AssignPlayerToken assigns a unique piece token to a player based on available tokens in session
func AssignPlayerToken(log zerolog.Logger, ctx context.Context, tx *pgxpool.Tx, sessionId string) (int, error) {
    // Get the count of existing players in the session to find next available token
    playerCount, err := internaldbplayers.GetPlayerCountInSession(log, ctx, tx, sessionId)
    if err != nil {
        log.Trace().Err(err).Msgf("failed to get player count for session %v", sessionId)
        return -1, err
    }

    // Assign token based on player count (cycling through available tokens if needed)
    tokenId := playerCount % TOTAL_PLAYER_TOKENS
    return tokenId, nil
}
