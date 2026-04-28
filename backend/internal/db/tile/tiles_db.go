package internaldb_tiles

import (
	"context"
	"database/sql"
	"monopoly-backend/internal"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

// GetAllTiles retrieves all tiles from the database and includes property
// data when a tile is linked to a property.
func GetAllTiles(
    log zerolog.Logger,
    ctx context.Context,
    tx *pgxpool.Tx,
    sessionId string,
) ([]internal.Tile, error) {
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


// GetTileByPosition retrieves a single tile by its position/id and includes
// property data if the tile is linked to a property.
func GetTileByPosition(
    log zerolog.Logger,
    ctx context.Context,
    tx *pgxpool.Tx,
    sessionId string,
    position int,
) (*internal.Tile, error) {
    var tile internal.Tile
    var raw_t struct {
        Id  int
        Name string
        PropertyId sql.NullInt32
    }
    err := tx.QueryRow(ctx, `
        SELECT
            id,
            name,
            property_id
        FROM tiles
        WHERE id = $1
        `, position).Scan(
        &raw_t.Id,
        &raw_t.Name,
        &raw_t.PropertyId,
    )
    if err != nil {
        log.Trace().Err(err).Msg("failed to query tile data")
        return nil, err
    }
    if raw_t.PropertyId.Valid {
        p, err := GetPropertyData(
            log,
            ctx,
            tx,
            sessionId,
            int(raw_t.PropertyId.Int32),
            )
        if err != nil {
            return nil, err
        }
        tile = internal.Tile{
            Id: raw_t.Id,
            Name: raw_t.Name,
            PropertyData: &p,
        }
    } else {
        tile = internal.Tile{
            Id: raw_t.Id,
            Name: raw_t.Name,
            PropertyData: nil,
        }
    }

    return &tile, nil
}


