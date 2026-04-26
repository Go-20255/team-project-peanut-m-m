package util

import (
    "context"
    internaldbplayers "monopoly-backend/internal/db/player"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/rs/zerolog"
)

const (
    TOTAL_PLAYER_TOKENS = 8
)

func AssignPlayerToken(log zerolog.Logger, ctx context.Context, tx *pgxpool.Tx, sessionId string) (int, error) {
    playerCount, err := internaldbplayers.GetPlayerCountInSession(log, ctx, tx, sessionId)
    if err != nil {
        log.Trace().Err(err).Msgf("failed to get player count for session %v", sessionId)
        return -1, err
    }

    tokenId := playerCount % TOTAL_PLAYER_TOKENS
    return tokenId, nil
}
