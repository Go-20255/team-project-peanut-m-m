package internaldb_tiles

import (
	"context"
	"database/sql"
	"monopoly-backend/internal"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

func GetAllTiles(log zerolog.Logger, ctx context.Context, tx *pgxpool.Tx, sessionId string) ([]internal.Tile, error) {
    var tiles []internal.Tile

    var raw_tiles []struct {
        Id  int
        Name string
        PropertyId sql.NullInt32
    }

    rows, err := tx.Query(ctx, `
        SELECT
            id,
            name,
            property_id
        FROM tiles
        `)
    if err != nil {
        log.Trace().Err(err).Msg("failed to get rows of tiles from db")
        return nil, err
    }

    for rows.Next() {
        var raw_t struct {
            Id  int
            Name string
            PropertyId sql.NullInt32
        }
        if err := rows.Scan(
            &raw_t.Id,
            &raw_t.Name,
            &raw_t.PropertyId,
            ); err != nil {
            log.Trace().Err(err).Msg("failed to scan raw tile data into struct")
            return nil, err
        }
        raw_tiles = append(raw_tiles, raw_t)
    }



    for _, r := range raw_tiles {

        if r.PropertyId.Valid {
            p, err := GetPropertyData(
                log,
                ctx,
                tx,
                sessionId,
                int(r.PropertyId.Int32),
                )
            if err != nil {
                return nil, err
            }
            tiles = append(tiles, internal.Tile{
                Id: r.Id,
                Name: r.Name,
                PropertyData: &p,
            })
        } else {
            tiles = append(tiles, internal.Tile{
                Id: r.Id,
                Name: r.Name,
                PropertyData: nil,
            })
        }
    }

    return tiles, nil
}
