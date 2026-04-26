package util

import (
    "context"
    internaldbplayers "monopoly-backend/internal/db/players"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/rs/zerolog"
)

// PLAYER_COLORS defines the available colors for players in order
var PLAYER_COLORS = []string{
    "#FF0000", // Red
    "#0000FF", // Blue
    "#FFD700", // Gold/Yellow
    "#00AA00", // Green
    "#FF8800", // Orange
    "#800080", // Purple
    "#FF69B4", // Hot Pink
    "#00FFFF", // Cyan
}

// AssignPlayerColor assigns a unique color to a player based on player count in session
func AssignPlayerColor(log zerolog.Logger, ctx context.Context, tx *pgxpool.Tx, sessionId string) (string, error) {
    // Get the count of existing players in the session
    playerCount, err := internaldbplayers.GetPlayerCountInSession(log, ctx, tx, sessionId)
    if err != nil {
        log.Trace().Err(err).Msgf("failed to get player count for session %v", sessionId)
        return "", err
    }

    // Assign color based on player count (cycling through available colors if needed)
    colorIndex := playerCount % len(PLAYER_COLORS)
    return PLAYER_COLORS[colorIndex], nil
}
