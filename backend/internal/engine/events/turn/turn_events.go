package turn

import (
    "context"
    "monopoly-backend/internal"
    internaldb_game_state "monopoly-backend/internal/db/game_state"
    internaldb_players "monopoly-backend/internal/db/player"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/rs/zerolog"
)

// GetCurrentPlayerIndex returns the player index for the given turn number.
// If the turn number is negative, it returns 0 as a default value.
func GetCurrentPlayerIndex(turnNumber int, playerCount int) int {
    if turnNumber < 0 {
        return 0
    }

    return turnNumber % playerCount
}

// GetCurrentPlayer retrieves the current player for a session based on the
// stored turn number.
// It also returns the full player list and the current turn number so callers
// can reuse that state without querying again. If the session has no players,
// it returns nil for the current player.
func GetCurrentPlayer(
    ctx context.Context,
    log zerolog.Logger,
    tx *pgxpool.Tx,
    sessionId string,
) (*internal.Player, []internal.Player, int, error) {
    players, err := internaldb_players.GetPlayersInSession(log, ctx, tx, sessionId)
    if err != nil {
        return nil, nil, 0, err
    }

    if len(players) == 0 {
        return nil, nil, 0, nil
    }

    turnNumber, err := internaldb_game_state.GetGameStateTurnNumber(log, ctx, tx, sessionId)
    if err != nil {
        return nil, nil, 0, err
    }

    currentPlayerIndex := GetCurrentPlayerIndex(turnNumber, len(players))
    for i := 0; i < len(players); i++ {
        player := players[(currentPlayerIndex+i)%len(players)]
        if !player.Bankrupt {
            return &player, players, turnNumber, nil
        }
    }

    return nil, players, turnNumber, nil
}

func GetActionPlayers(
    ctx context.Context,
    log zerolog.Logger,
    tx *pgxpool.Tx,
    sessionId string,
    playerId int,
) (*internal.Player, *internal.Player, []internal.Player, int, error) {
    currentPlayer, players, turnNumber, err := GetCurrentPlayer(ctx, log, tx, sessionId)
    if err != nil {
        return nil, nil, nil, 0, err
    }

    actingPlayer, err := internaldb_players.GetPlayer(log, ctx, tx, playerId, sessionId)
    if err != nil {
        return nil, nil, nil, 0, err
    }

    return currentPlayer, &actingPlayer, players, turnNumber, nil
}
