package internaldb_players

import (
    "context"
    "monopoly-backend/internal"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/rs/zerolog"
)

func CreatePlayerDB(log zerolog.Logger, ctx context.Context, tx *pgxpool.Tx, name string, sessionId string) (int, error) {
    query := `
        INSERT INTO Player (name, session_id)
        VALUES ($1, $2)
        RETURNING id
        `

    var id int
    err := tx.QueryRow(ctx, query, name, sessionId).Scan(&id)

    if err != nil {
        log.Trace().Err(err).Msgf("failed to insert player %v into db with session id %v", name, sessionId)
        return 0, err
    }
    return id, nil
}

func CheckPlayerExists(
    log zerolog.Logger,
    ctx context.Context,
    tx *pgxpool.Tx,
    id string,
    name string,
    session_id string,
) (bool, error) {
    var exists bool
    query := `
    SELECT EXISTS (
    SELECT 1
    FROM Player
    WHERE id = $1 AND name = $2 AND session_id = $3
    )
    `
    err := tx.QueryRow(ctx, query, id, name, session_id).Scan(&exists)
    if err != nil {
        log.Trace().Err(err).Msg("failed to query if player exists in db")
        return false, err
    }

    return exists, nil
}

func GetPlayer(log zerolog.Logger, ctx context.Context, tx *pgxpool.Tx, id int, sessionId string) (internal.Player, error) {
    var player internal.Player
    err := tx.QueryRow(ctx, `
        SELECT
            id,
            name,
            COALESCE(player_order, -1),
            money,
            position,
            get_out_of_jail_cards,
            jailed > 0,
            session_id,
            in_game
        FROM player
        WHERE id = $1 AND session_id = $2
        `, id, sessionId).Scan(
        &player.Id,
        &player.Name,
        &player.PlayerOrder,
        &player.Money,
        &player.Position,
        &player.GetOutOfJailCards,
        &player.Jailed,
        &player.SessionId,
        &player.InGame,
    )
    if err != nil {
        log.Trace().Err(err).Msg("failed to query player")
        return internal.Player{}, err
    }

    return player, nil
}

func GetPlayersInSession(log zerolog.Logger, ctx context.Context, tx *pgxpool.Tx, sessionId string) ([]internal.Player, error) {
    rows, err := tx.Query(ctx, `
        SELECT
            id,
            name,
            COALESCE(player_order, -1),
            money,
            position,
            get_out_of_jail_cards,
            jailed > 0,
            session_id,
            in_game
        FROM player
        WHERE session_id = $1
        ORDER BY
            CASE WHEN player_order IS NULL THEN 1 ELSE 0 END,
            player_order ASC,
            id ASC
        `, sessionId)
    if err != nil {
        log.Trace().Err(err).Msg("failed to query players in session")
        return nil, err
    }
    defer rows.Close()

    var players []internal.Player
    for rows.Next() {
        var player internal.Player
        if err := rows.Scan(
            &player.Id,
            &player.Name,
            &player.PlayerOrder,
            &player.Money,
            &player.Position,
            &player.GetOutOfJailCards,
            &player.Jailed,
            &player.SessionId,
            &player.InGame,
        ); err != nil {
            log.Trace().Err(err).Msg("failed to scan player in session")
            return nil, err
        }
        players = append(players, player)
    }

    if err := rows.Err(); err != nil {
        log.Trace().Err(err).Msg("failed while iterating over players in session")
        return nil, err
    }

    return players, nil
}

func UpdatePlayerPosition(log zerolog.Logger, ctx context.Context, tx *pgxpool.Tx, id int, sessionId string, position int) error {
    commandTag, err := tx.Exec(ctx, `
        UPDATE player
        SET position = $1
        WHERE id = $2 AND session_id = $3
        `, position, id, sessionId)
    if err != nil {
        log.Trace().Err(err).Msg("failed to update player position")
        return err
    }

    if commandTag.RowsAffected() == 0 {
        return pgx.ErrNoRows
    }

    return nil
}
