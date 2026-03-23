package internaldb

import (
	"context"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)

func SetupDatabase(ctx context.Context, log zerolog.Logger) {
	monopolyDbUrlStr := "postgres://postgres:<pass>@localhost:<port>/monopoly?sslmode=disable"
	monopolyDbPort := os.Getenv("POSTGRES_PORT")
	if monopolyDbPort == "" {
		monopolyDbPort = "1357"
	}
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	monopolyDbUrl := strings.ReplaceAll(monopolyDbUrlStr, "<pass>", postgresPassword)
	monopolyDbUrl = strings.ReplaceAll(monopolyDbUrl, "<port>", monopolyDbPort)

	monopolyDefaultDbUrlStr := "postgres://postgres:<pass>@localhost:<port>/postgres?sslmode=disable"

	monopolyDefaultDbUrl := strings.ReplaceAll(monopolyDefaultDbUrlStr, "<pass>", postgresPassword)
	monopolyDefaultDbUrl = strings.ReplaceAll(monopolyDefaultDbUrl, "<port>", monopolyDbPort)

	db, err := pgx.Connect(ctx, monopolyDefaultDbUrl)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres database")
	}
	defer func(db *pgx.Conn, ctx context.Context) {
		err := db.Close(ctx)
		if err != nil {

		}
	}(db, ctx)

	var dbExists bool
	err = db.QueryRow(
		context.Background(),
		"SELECT EXISTS (SELECT FROM pg_database WHERE datname = 'monopoly');",
	).Scan(&dbExists)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to check if database exists")
	}

	if !dbExists {
		log.Info().Msg("database monopoly does not exist. creating it...")
		_, err = db.Exec(ctx, "CREATE DATABASE \"monopoly\";")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create database monopoly")
		}
		log.Info().Msg("database monopoly created successfully.")
	} else {
		log.Info().Msg("database monopoly already exists.")
	}

	db, err = pgx.Connect(ctx, monopolyDbUrl)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres database")
	}
	defer func(db *pgx.Conn, ctx context.Context) {
		err := db.Close(ctx)
		if err != nil {

		}
	}(db, ctx)

	var tableExists bool
	err = db.QueryRow(ctx, "SELECT EXISTS (SELECT FROM pg_tables WHERE tablename = 'rents')").Scan(&tableExists)
	if err != nil {
		log.Panic().Err(err).Msg("failed to check if tables exist")
	}
	if tableExists {
		log.Printf("Tables already exist. Skipping setup.")
		return
	}

	log.Info().Msg("tables not found. creating tables...")

	tx, err := db.Begin(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to begin transaction")
	}
	defer func() {
		if err != nil {
			err := tx.Rollback(ctx)
			if err != nil {
				return
			} // Rollback the transaction on error
			log.Error().Err(err).Msg("transaction rolled back due to error.")
		}
	}()

	// TODO: start adding tables here
	_, err = tx.Exec(ctx, `
CREATE TYPE property_type AS ENUM ('BROWN', 'LIGHTBLUE', 'PINK', 'ORANGE', 'RED', 'YELLOW', 'GREEN', 'DARKBLUE', 'RAILROAD', 'UTILITY')
        `)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create enum of property types")
	}

	_, err = tx.Exec(ctx, `
        CREATE TABLE Game_State (
        session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        turn_number INTEGER NOT NULL DEFAULT -1,
        code INTEGER NOT NULL
        )
        `)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create game state table")
	}

	_, err = tx.Exec(ctx, `
        CREATE TABLE Player (
        id SERIAL PRIMARY KEY,
        name TEXT NOT NULL,
        player_order INTEGER, -- Becomes not null later when turns are decided via "dice" roll
        money INTEGER NOT NULL DEFAULT 0,
        position INTEGER NOT NULL DEFAULT 0, -- index of position into 1D board array
        get_out_of_jail_cards INTEGER NOT NULL DEFAULT 0, -- number of get out of jail cards held
        jailed INTEGER NOT NULL DEFAULT 0, -- number of turns stuck in jail
        session_id UUID REFERENCES Game_State(session_id) ON DELETE CASCADE NOT NULL,
        in_game BOOLEAN NOT NULL DEFAULT FALSE,
        CONSTRAINT unique_session_name UNIQUE(name, session_id)
        )
        `)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create player table")
	}

	_, err = tx.Exec(ctx, `
        CREATE TABLE Rents (
        id INTEGER PRIMARY KEY,
        base INTEGER NOT NULL,
        color_set INTEGER NOT NULL,
        one_house INTEGER NOT NULL,
        two_house INTEGER NOT NULL,
        three_house INTEGER NOT NULL,
        four_house INTEGER NOT NULL,
        hotel INTEGER NOT NULL
        )
        `)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create rents table")
	}

	// TODO: update insert with real rent values
	_, err = tx.Exec(ctx, `
        INSERT INTO Rents (id, base, color_set, one_house, two_house, three_house, four_house, hotel)
        VALUES
            (0, 10, 25, 100, 200, 300, 400, 1000000) -- note the id as you will need to match it later with its respective property
        `)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to insert rents into db")
	}

	_, err = tx.Exec(ctx, `
        CREATE TABLE Property (
        id INTEGER PRIMARY KEY,
        name TEXT NOT NULL,
        rentvalues_id INTEGER REFERENCES Rents(id), -- points to a row in rents table that contains all property-specific rents (tbh these values could also just be a part of this table...) (null if utility or railroad)
        purchase_cost INTEGER NOT NULL, -- cost of property to buy
        mortgage_cost INTEGER NOT NULL, -- value gained from mortgaging
        unmortgage_cost INTEGER NOT NULL, -- price to pay to remove mortgage
        house_cost INTEGER, -- base value needed to buy a house (null if utility or railroad)
        hotel_cost INTEGER, -- base value needed to buy a hotel (null if utility or railroad)
        ptype property_type NOT NULL -- property type
        )
        `)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create property table")
	}

	// TODO: update insert with real properties
	_, err = tx.Exec(ctx, `
        INSERT INTO Property (id, name, rentvalues_id, purchase_cost, mortgage_cost, unmortgage_cost, house_cost, hotel_cost, ptype)
        VALUES
            (0, 'test property', 0, 120, 100, 110, 50, 100, 'BROWN') -- rentvalues_id value is the rent values we want referenced in the rent table
        `)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to insert properties into db")
	}

	_, err = tx.Exec(ctx, `
        CREATE TABLE Event_Cards ( -- Community and Chance Cards
        id SERIAL PRIMARY KEY,
        name TEXT NOT NULL,
        description TEXT,
        type TEXT CHECK (type IN ('COMMUNITY', 'CHANCE'))
        )
        `)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create event cards table")
	}

	// TODO: update with actual cards
	_, err = tx.Exec(ctx, `
        INSERT INTO Event_Cards (name, description, type)
        VALUES 
        ('example community card', 'example description', 'COMMUNITY'),
        ('example chance card', 'example description', 'CHANCE')
        `)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to insert event cards into db")
	}

	_, err = tx.Exec(ctx, `
        CREATE TABLE Owned_Properties (
        id SERIAL PRIMARY KEY,
        property_id INTEGER REFERENCES Property(id),
        session_id UUID REFERENCES Game_State(session_id) ON DELETE CASCADE NOT NULL,
        owner_id INTEGER REFERENCES Player(id) ON DELETE CASCADE NOT NULL,
        mortgaged BOOLEAN NOT NULL DEFAULT False,
        houses INTEGER NOT NULL DEFAULT 0,
        hotel BOOLEAN NOT NULL DEFAULT False
        )
        `)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create owned_properties table")
	}

	_, err = tx.Exec(ctx, `
        CREATE TABLE Drawn_Event_Cards (
        id SERIAL PRIMARY KEY,
        session_id UUID REFERENCES Game_State(session_id) ON DELETE CASCADE NOT NULL,
        card_id INTEGER REFERENCES Event_Cards(id) NOT NULL
        )
        `)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create drawn event cards table")
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to commit transaction")
	}
	log.Info().Msg("setup database transaction committed successfully.")
}
