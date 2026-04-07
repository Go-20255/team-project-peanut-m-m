package internaldb_properties

import (
    "context"
	"fmt"
	"github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/rs/zerolog"
)

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

	// ownerid, owned, err
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