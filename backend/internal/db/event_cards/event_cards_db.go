package internaldb_event_cards

import (
    "context"
    "database/sql"
    "fmt"
    "monopoly-backend/internal"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/rs/zerolog"
)

func EmptyEventCardPileDB(log zerolog.Logger, ctx context.Context, db *pgxpool.Pool, sessionId string, cardType string) (bool, error) {
	var count int
	err := db.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM event_cards ec 
		WHERE ec.type = $2 
		AND NOT EXISTS (
			SELECT 1 FROM drawn_event_cards dec 
			WHERE dec.card_id = ec.id AND dec.session_id = $1
		)
	`, sessionId, cardType).Scan(&count)

	if err != nil {
		log.Error().Err(err).Msg("failed to check if event card pile is empty")
		return false, err
	}

	return count == 0, nil
}

func ReshuffleEventCardPileDB(log zerolog.Logger, ctx context.Context, db *pgxpool.Pool, sessionId string, cardType string) error {
	_, err := db.Exec(ctx, `
		DELETE FROM drawn_event_cards
		USING event_cards 
		WHERE drawn_event_cards.card_id = event_cards.id 
		  AND drawn_event_cards.session_id = $1 
		  AND event_cards.type = $2 
		  AND event_cards.id NOT IN (9, 21)
	`, sessionId, cardType)

	if err != nil {
		log.Error().Err(err).Msg("failed to reshuffle event card pile")
		return err
	}

	return nil
}

func AssignEventCardDB(log zerolog.Logger, ctx context.Context, db *pgxpool.Pool, sessionId string, cardType string, playerId int) (int, error) {
	isEmpty, err := EmptyEventCardPileDB(log, ctx, db, sessionId, cardType)
	if err != nil {
		return -1, err
	}

	if isEmpty {
		err = ReshuffleEventCardPileDB(log, ctx, db, sessionId, cardType)
		if err != nil {
			return -1, err
		}
	}

	var cardId int
	err = db.QueryRow(ctx, `
		SELECT ec.id 
		FROM event_cards ec 
		WHERE ec.type = $2 
		AND NOT EXISTS (
			SELECT 1 FROM drawn_event_cards dec 
			WHERE dec.card_id = ec.id AND dec.session_id = $1
		)
		ORDER BY RANDOM() 
		LIMIT 1
	`, sessionId, cardType).Scan(&cardId)

	if err != nil {
		log.Error().Err(err).Msg("failed to draw random event card")
		return -1, err
	}

	_, err = db.Exec(ctx, `
		INSERT INTO drawn_event_cards (session_id, card_id)
		VALUES ($1, $2)
	`, sessionId, cardId)
	
	if err != nil {
		log.Error().Err(err).Msg("failed to mark event card as drawn")
		return -1, err
	}

	if cardId == 9 || cardId == 21 {
		_, err = db.Exec(ctx, `
			UPDATE player
			SET get_out_of_jail_cards = get_out_of_jail_cards + 1
			WHERE id = $1
		`, playerId)
		if err != nil {
			log.Error().Err(err).Msg("failed to assign get out of jail free card to player")
			return -1, err
		}
	}

	return cardId, nil
}