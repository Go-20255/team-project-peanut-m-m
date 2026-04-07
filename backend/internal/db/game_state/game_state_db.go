package internaldb_game_state

import (
    "context"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/rs/zerolog"
)

func GameStateExists(log zerolog.Logger, ctx context.Context, tx *pgxpool.Tx, sessionId string) (bool, error) {
    var exists bool

    err := tx.QueryRow(ctx, `
        SELECT EXISTS (
            SELECT 1
            FROM game_state
            WHERE session_id = $1
        )
    `, sessionId).Scan(&exists)
    if err != nil {
        log.Trace().Err(err).Msg("failed to query game state")
        return false, err
    }

    return exists, nil
}

func CheckGameStateCode(log zerolog.Logger, ctx context.Context, tx *pgxpool.Tx, code string) (string, error) {
    var id string
    err := tx.QueryRow(ctx, `
        SELECT session_id FROM Game_State WHERE code = $1
        `, code).Scan(&id)
    if err != nil {
        return "", err
    }
    return id, nil
}

func CreateGameState(log zerolog.Logger, ctx context.Context, tx *pgxpool.Tx) (string, error) {

    var id string
    err := tx.QueryRow(ctx, `
        INSERT INTO game_state
        DEFAULT VALUES
        RETURNING session_id
        `).Scan(&id)
    if err != nil {
        log.Trace().Err(err).Msg("failed to create new game state")
        return "", err
    }

    return id, nil
}

func GetGameStateTurnNumber(log zerolog.Logger, ctx context.Context, tx *pgxpool.Tx, sessionId string) (int, error) {
    var turnNumber int
    err := tx.QueryRow(ctx, `
        SELECT turn_number
        FROM game_state
        WHERE session_id = $1
        `, sessionId).Scan(&turnNumber)
    if err != nil {
        log.Trace().Err(err).Msg("failed to query game state turn number")
        return 0, err
    }

    return turnNumber, nil
}

func UpdateGameStateTurnNumber(log zerolog.Logger, ctx context.Context, tx *pgxpool.Tx, sessionId string, turnNumber int) error {
    _, err := tx.Exec(ctx, `
        UPDATE game_state
        SET turn_number = $1
        WHERE session_id = $2
        `, turnNumber, sessionId)
    if err != nil {
        log.Trace().Err(err).Msg("failed to update game state turn number")
        return err
    }

    return nil
}

func GetGameSessions(log zerolog.Logger, ctx context.Context, tx *pgxpool.Tx) ([]string, error) {

    var sessionIds []string
    rows, err := tx.Query(ctx, `
        SELECT session_id
        FROM game_state
        `)
    if err != nil {
        log.Trace().Err(err).Msg("failed to query db for game sessions")
        return nil, err
    }

    for rows.Next() {
        var sessionId string
        if err := rows.Scan(&sessionId); err != nil {
            log.Trace().Err(err).Msg("failed to unmarshal session_id")
            return nil, err
        }
        sessionIds = append(sessionIds, sessionId)
    }

    return sessionIds, nil
}
