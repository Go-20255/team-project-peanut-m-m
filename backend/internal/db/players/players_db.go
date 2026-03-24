package internaldb_players

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

func CreatePlayerDB(log zerolog.Logger, ctx context.Context, tx *pgxpool.Tx, name string, sessionId string) error {

	_, err := tx.Exec(ctx, `
        INSERT INTO Player (name, session_id)
        VALUES ($1, $2)
        `, name, sessionId)
	if err != nil {
		log.Trace().Err(err).Msgf("failed to insert player %v into db with session id %v", name, sessionId)
		return err
	}
	return nil
}
