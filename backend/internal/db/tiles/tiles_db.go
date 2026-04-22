package internaldb_tiles

import (
	"context"
	"monopoly-backend/internal"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

func GetAllTiles(log zerolog.Logger, ctx context.Context, tx *pgxpool.Tx, sessionId string) ([]internal.Tile, error) {
    var tiles []internal.Tile

    return tiles, nil
}
