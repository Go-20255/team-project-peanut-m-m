package turn_events

import (
    "context"
    "monopoly-backend/internal"
    internaldb_game_state "monopoly-backend/internal/db/game_state"
    internaldb_players "monopoly-backend/internal/db/players"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/rs/zerolog"
)

func GetCurrentPlayerIndex(turnNumber int, playerCount int) int {
    if turnNumber < 0 {
        return 0
    }

    return turnNumber % playerCount
}

func GetCurrentPlayer(
    ctx context.Context,
    log zerolog.Logger,
    tx *pgxpool.Tx,
    sessionId string,
) (internal.Player, []internal.Player, int, error) {
    players, err := internaldb_players.GetPlayersInSession(log, ctx, tx, sessionId)
    if err != nil {
        return internal.Player{}, nil, 0, err
    }

    if len(players) == 0 {
        return internal.Player{}, nil, 0, nil
    }

    turnNumber, err := internaldb_game_state.GetGameStateTurnNumber(log, ctx, tx, sessionId)
    if err != nil {
        return internal.Player{}, nil, 0, err
    }

    currentPlayer := players[GetCurrentPlayerIndex(turnNumber, len(players))]
    return currentPlayer, players, turnNumber, nil
}
