package internaldb_tiles

import (
    "context"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/rs/zerolog"
)

type RentTileData struct {
    Position       int
	Name           string
    PropertyId     int
    HasProperty    bool
    PropertyType   string
    OwnerId        int
    Owned          bool
    IsMortgaged    bool
    Houses         int
    HasHotel       bool
    BaseRent       int
    ColorSetRent   int
    OneHouseRent   int
    TwoHouseRent   int
    ThreeHouseRent int
    FourHouseRent  int
    HotelRent      int
}

func GetRentTileData(log zerolog.Logger, ctx context.Context, tx *pgxpool.Tx, sessionId string, position int) (RentTileData, error) {
    var rentTileData RentTileData
    err := tx.QueryRow(ctx, `
        SELECT
            t.id,
			t.name,
            COALESCE(t.property_id, 0),
            t.property_id IS NOT NULL,
            COALESCE(p.ptype::text, ''),
            COALESCE(op.owner_id, 0),
            op.id IS NOT NULL,
            COALESCE(op.mortgaged, false),
            COALESCE(op.houses, 0),
            COALESCE(op.hotel, false),
            COALESCE(r.base, 0),
            COALESCE(r.color_set, 0),
            COALESCE(r.one_house, 0),
            COALESCE(r.two_house, 0),
            COALESCE(r.three_house, 0),
            COALESCE(r.four_house, 0),
            COALESCE(r.hotel, 0)
        FROM tiles t
        LEFT JOIN property p
            ON p.id = t.property_id
        LEFT JOIN owned_properties op
            ON op.property_id = t.property_id
            AND op.session_id = $1
        LEFT JOIN rents r
            ON r.id = p.rentvalues_id
        WHERE t.id = $2
        `, sessionId, position).Scan(
        &rentTileData.Position,
		&rentTileData.Name,
        &rentTileData.PropertyId,
        &rentTileData.HasProperty,
        &rentTileData.PropertyType,
        &rentTileData.OwnerId,
        &rentTileData.Owned,
        &rentTileData.IsMortgaged,
        &rentTileData.Houses,
        &rentTileData.HasHotel,
        &rentTileData.BaseRent,
        &rentTileData.ColorSetRent,
        &rentTileData.OneHouseRent,
        &rentTileData.TwoHouseRent,
        &rentTileData.ThreeHouseRent,
        &rentTileData.FourHouseRent,
        &rentTileData.HotelRent,
    )
    if err != nil {
        log.Trace().Err(err).Msg("failed to query rent tile data")
        return RentTileData{}, err
    }

    return rentTileData, nil
}

func GetOwnedPropertyTypeCount(log zerolog.Logger, ctx context.Context, tx *pgxpool.Tx, sessionId string, ownerId int, propertyType string) (int, error) {
    var propertyCount int
    err := tx.QueryRow(ctx, `
        SELECT COUNT(*)
        FROM owned_properties op
        JOIN property p
            ON p.id = op.property_id
        WHERE op.session_id = $1
            AND op.owner_id = $2
            AND p.ptype::text = $3
        `, sessionId, ownerId, propertyType).Scan(&propertyCount)
    if err != nil {
        log.Trace().Err(err).Msg("failed to query owned property type count")
        return 0, err
    }

    return propertyCount, nil
}
