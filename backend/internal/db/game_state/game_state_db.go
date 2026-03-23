package internaldb_game_state

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)


func GameStateExists(log zerolog.Logger, ctx context.Context, tx *pgxpool.Tx, session_id int) error {

    rows, err := tx.Query(ctx, `
        SELECT session_id FROM Game_State WHERE session_id = $1
        `, session_id)
    if err != nil{
        // session id does not exist
        rows.Close()
        return err
    }
    rows.Close()
    return nil
}

