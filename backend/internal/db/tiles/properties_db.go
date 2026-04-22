package internaldb_tiles

import (
	"context"
	"fmt"
	"monopoly-backend/internal"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type PropertyGroupData struct {
    PropertyId   int
    PropertyType string
    OwnerId      int
    Owned        bool
    IsMortgaged  bool
    Houses       int
    HasHotel     bool
    HouseCost    int
    HotelCost    int
}

func VerifyPropertyOwnerDB(log zerolog.Logger, ctx context.Context, tx *pgxpool.Tx, sessionId string, propertyId int) (int, bool, error) {
    var ownerId int

    err := tx.QueryRow(ctx, `
        SELECT owner_id
        FROM Owned_Properties
        WHERE session_id = $1 AND property_id = $2
    `, sessionId, propertyId).Scan(&ownerId)

    if err != nil {
        if err == pgx.ErrNoRows {
            return 0, false, nil
        }

        log.Trace().Err(err).Msg("failed to query property ownership")
        return 0, false, err
    }

    return ownerId, true, nil
}

func CreatePropertyOwnerDB(log zerolog.Logger, ctx context.Context, tx *pgxpool.Tx, sessionId string, playerId int, propertyId int) (int, error) {
    _, isOwned, err := VerifyPropertyOwnerDB(log, ctx, tx, sessionId, propertyId)
    if err != nil {
        return -1, err
    }

    if isOwned {
        err := fmt.Errorf("property %d is already owned in session %s", propertyId, sessionId)
        log.Trace().Err(err).Msg("attempted to buy an already owned property")
        return -1, err
    }

    var id int
    err = tx.QueryRow(ctx, `
        INSERT INTO Owned_Properties (property_id, session_id, owner_id)
        VALUES ($1, $2, $3)
        RETURNING id
    `, propertyId, sessionId, playerId).Scan(&id)
    if err != nil {
        log.Trace().Err(err).Msg("failed to insert property ownership")
        return 0, err
    }

    return id, nil
}

func GetPropertyData(log zerolog.Logger, ctx context.Context, tx *pgxpool.Tx, sessionId string, propertyId int) (internal.PropertyData, error) {
    var p internal.PropertyData

    err := tx.QueryRow(ctx,`
        SELECT
            id,
            name,
            rentvalues_id,
            purchase_cost,
            mortgage_cost,
            unmortgage_cost,
            house_cost,
            hotel_cost,
            ptype
        FROM property
        WHERE id = $1
        `, propertyId).Scan(
        &p.Id,
        &p.Name,
        &p.RentId,
        &p.PurchaseCost,
        &p.MortgageCost,
        &p.UnmortgageCost,
        &p.HotelCost,
        &p.HotelCost,
        &p.PropertyType,
        )
    if err != nil {
        log.Trace().Err(err).Msg("failed to get property data from db")
        return p, err
    }

    return p, nil
}

func GetPropertyGroupData(log zerolog.Logger, ctx context.Context, tx *pgxpool.Tx, sessionId string, propertyId int) ([]PropertyGroupData, error) {
    rows, err := tx.Query(ctx, `
        SELECT
            p.id,
            p.ptype::text,
            COALESCE(op.owner_id, 0),
            op.id IS NOT NULL,
            COALESCE(op.mortgaged, false),
            COALESCE(op.houses, 0),
            COALESCE(op.hotel, false),
            COALESCE(p.house_cost, 0),
            COALESCE(p.hotel_cost, 0)
        FROM property p
        LEFT JOIN owned_properties op
            ON op.property_id = p.id
            AND op.session_id = $1
        WHERE p.ptype = (
            SELECT ptype
            FROM property
            WHERE id = $2
        )
        ORDER BY p.id
        `, sessionId, propertyId)
    if err != nil {
        log.Trace().Err(err).Msg("failed to query property group data")
        return nil, err
    }
    defer rows.Close()

    var propertyGroup []PropertyGroupData
    for rows.Next() {
        var propertyData PropertyGroupData
        if err := rows.Scan(
            &propertyData.PropertyId,
            &propertyData.PropertyType,
            &propertyData.OwnerId,
            &propertyData.Owned,
            &propertyData.IsMortgaged,
            &propertyData.Houses,
            &propertyData.HasHotel,
            &propertyData.HouseCost,
            &propertyData.HotelCost,
        ); err != nil {
            log.Trace().Err(err).Msg("failed to scan property group data")
            return nil, err
        }
        propertyGroup = append(propertyGroup, propertyData)
    }

    if err := rows.Err(); err != nil {
        log.Trace().Err(err).Msg("failed while iterating over property group data")
        return nil, err
    }

    return propertyGroup, nil
}

func UpdatePropertyBuildings(log zerolog.Logger, ctx context.Context, tx *pgxpool.Tx, sessionId string, propertyId int, houses int, hotel bool) error {
    commandTag, err := tx.Exec(ctx, `
        UPDATE owned_properties
        SET houses = $1, hotel = $2
        WHERE session_id = $3 AND property_id = $4
        `, houses, hotel, sessionId, propertyId)
    if err != nil {
        log.Trace().Err(err).Msg("failed to update property buildings")
        return err
    }

    if commandTag.RowsAffected() == 0 {
        return pgx.ErrNoRows
    }

    return nil
}

func GetAvailableHouses(log zerolog.Logger, ctx context.Context, tx *pgxpool.Tx, sessionId string) (int, error) {
    var availableHouses int
    err := tx.QueryRow(ctx, `
        SELECT 32 - COALESCE(SUM(houses), 0)
        FROM owned_properties
        WHERE session_id = $1
        `, sessionId).Scan(&availableHouses)
    if err != nil {
        log.Trace().Err(err).Msg("failed to query available houses")
        return 0, err
    }

    return availableHouses, nil
}

func GetAvailableHotels(log zerolog.Logger, ctx context.Context, tx *pgxpool.Tx, sessionId string) (int, error) {
    var availableHotels int
    err := tx.QueryRow(ctx, `
        SELECT 12 - COUNT(*)
        FROM owned_properties
        WHERE session_id = $1 AND hotel = true
        `, sessionId).Scan(&availableHotels)
    if err != nil {
        log.Trace().Err(err).Msg("failed to query available hotels")
        return 0, err
    }

    return availableHotels, nil
}
