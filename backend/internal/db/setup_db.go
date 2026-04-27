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
        CREATE OR REPLACE FUNCTION generate_unique_6_digit_code()
RETURNS INTEGER AS $$
DECLARE
    new_code INTEGER;
BEGIN
    LOOP
        new_code := floor(100000 + random() * 900000)::int;

        EXIT WHEN NOT EXISTS (
            SELECT 1 FROM game_state WHERE code = new_code
        );
    END LOOP;

    RETURN new_code;
END;
$$ LANGUAGE plpgsql;
        `)
    if err != nil {
        log.Fatal().Err(err).Msg("failed to define generate_unique_6_digit_code function in db")
    }

    _, err = tx.Exec(ctx, `
        CREATE TABLE Game_State (
        session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        turn_number INTEGER NOT NULL DEFAULT -1,
        code INTEGER NOT NULL UNIQUE DEFAULT generate_unique_6_digit_code()
        )
        `)
    if err != nil {
        log.Fatal().Err(err).Msg("failed to create game state table")
    }

    _, err = tx.Exec(ctx, `
        CREATE TABLE Player (
        id SERIAL PRIMARY KEY,
        name TEXT NOT NULL,
        ready_up_status BOOLEAN NOT NULL DEFAULT FALSE,
        piece_token INTEGER,
        player_order INTEGER, -- Becomes not null later when turns are decided via "dice" roll
        money INTEGER NOT NULL DEFAULT 1500,
        position INTEGER NOT NULL DEFAULT 0, -- index of position into 1D board array
        get_out_of_jail_cards INTEGER NOT NULL DEFAULT 0, -- number of get out of jail cards held
        jailed INTEGER NOT NULL DEFAULT 0, -- number of turns stuck in jail
        session_id UUID REFERENCES Game_State(session_id) ON DELETE CASCADE NOT NULL,
        in_game BOOLEAN NOT NULL DEFAULT FALSE,
        CONSTRAINT unique_session_name UNIQUE(name, session_id),
        CONSTRAINT unique_session_token UNIQUE(piece_token, session_id)
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

    _, err = tx.Exec(ctx, `
        INSERT INTO Rents (id, base, color_set, one_house, two_house, three_house, four_house, hotel)
        VALUES
            -- Brown
            (1, 2, 4, 10, 30, 90, 160, 250),     -- Mediterranean Ave
            (3, 4, 8, 20, 60, 180, 320, 450),    -- Baltic Ave
            -- Light Blue
            (6, 6, 12, 30, 90, 270, 400, 550),   -- Oriental Ave
            (8, 6, 12, 30, 90, 270, 400, 550),   -- Vermont Ave
            (9, 8, 16, 40, 100, 300, 450, 600),  -- Connecticut Ave
            -- Pink
            (11, 10, 20, 50, 150, 450, 625, 750),  -- St. Charles Place
            (13, 10, 20, 50, 150, 450, 625, 750),  -- States Ave
            (14, 12, 24, 60, 180, 500, 700, 900),  -- Virginia Ave
            -- Orange
            (16, 14, 28, 70, 200, 550, 750, 950),  -- St. James Place
            (18, 14, 28, 70, 200, 550, 750, 950),  -- Tennessee Ave
            (19, 16, 32, 80, 220, 600, 800, 1000), -- New York Ave
            -- Red
            (21, 18, 36, 90, 250, 700, 875, 1050),   -- Kentucky Ave
            (23, 18, 36, 90, 250, 700, 875, 1050),   -- Indiana Ave
            (24, 20, 40, 100, 300, 750, 925, 1100),  -- Illinois Ave
            -- Yellow
            (26, 22, 44, 110, 330, 800, 975, 1150),  -- Atlantic Ave
            (27, 22, 44, 110, 330, 800, 975, 1150),  -- Ventnor Ave
            (29, 24, 48, 120, 360, 850, 1025, 1200), -- Marvin Gardens
            -- Green
            (31, 26, 52, 130, 390, 900, 1100, 1275), -- Pacific Ave
            (32, 26, 52, 130, 390, 900, 1100, 1275), -- North Carolina Ave
            (34, 28, 56, 150, 450, 1000, 1200, 1400),-- Pennsylvania Ave
            -- Dark Blue
            (37, 35, 70, 175, 500, 1100, 1300, 1500),-- Park Place
            (39, 50, 100, 200, 600, 1400, 1700, 2000)-- Boardwalk
        ON CONFLICT (id) DO NOTHING;
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

    _, err = tx.Exec(ctx, `
        INSERT INTO Property (id, name, rentvalues_id, purchase_cost, mortgage_cost, unmortgage_cost, house_cost, hotel_cost, ptype)
        VALUES
            -- Brown (House cost: 50)
            (1, 'Mediterranean Avenue', 1, 60, 30, 33, 50, 50, 'BROWN'),
            (2, 'Baltic Avenue', 3, 60, 30, 33, 50, 50, 'BROWN'),
            
            -- Railroad 1
            (3, 'Reading Railroad', NULL, 200, 100, 110, NULL, NULL, 'RAILROAD'),

            -- Light Blue (House cost: 50)
            (4, 'Oriental Avenue', 6, 100, 50, 55, 50, 50, 'LIGHTBLUE'),
            (5, 'Vermont Avenue', 8, 100, 50, 55, 50, 50, 'LIGHTBLUE'),
            (6, 'Connecticut Avenue', 9, 120, 60, 66, 50, 50, 'LIGHTBLUE'),

            -- Pink (House cost: 100)
            (7, 'St. Charles Place', 11, 140, 70, 77, 100, 100, 'PINK'),
            (8, 'Electric Company', NULL, 150, 75, 83, NULL, NULL, 'UTILITY'),
            (9, 'States Avenue', 13, 140, 70, 77, 100, 100, 'PINK'),
            (10, 'Virginia Avenue', 14, 160, 80, 88, 100, 100, 'PINK'),

            -- Railroad 2
            (11, 'Pennsylvania Railroad', NULL, 200, 100, 110, NULL, NULL, 'RAILROAD'),

            -- Orange (House cost: 100)
            (12, 'St. James Place', 16, 180, 90, 99, 100, 100, 'ORANGE'),
            (13, 'Tennessee Avenue', 18, 180, 90, 99, 100, 100, 'ORANGE'),
            (14, 'New York Avenue', 19, 200, 100, 110, 100, 100, 'ORANGE'),

            -- Red (House cost: 150)
            (15, 'Kentucky Avenue', 21, 220, 110, 121, 150, 150, 'RED'),
            (16, 'Indiana Avenue', 23, 220, 110, 121, 150, 150, 'RED'),
            (17, 'Illinois Avenue', 24, 240, 120, 132, 150, 150, 'RED'),

            -- Railroad 3
            (18, 'B. & O. Railroad', NULL, 200, 100, 110, NULL, NULL, 'RAILROAD'),

            -- Yellow (House cost: 150)
            (19, 'Atlantic Avenue', 26, 260, 130, 143, 150, 150, 'YELLOW'),
            (20, 'Ventnor Avenue', 27, 260, 130, 143, 150, 150, 'YELLOW'),
            (21, 'Water Works', NULL, 150, 75, 83, NULL, NULL, 'UTILITY'),
            (22, 'Marvin Gardens', 29, 280, 140, 154, 150, 150, 'YELLOW'),

            -- Green (House cost: 200)
            (23, 'Pacific Avenue', 31, 300, 150, 165, 200, 200, 'GREEN'),
            (24, 'North Carolina Avenue', 32, 300, 150, 165, 200, 200, 'GREEN'),
            (25, 'Pennsylvania Avenue', 34, 320, 160, 176, 200, 200, 'GREEN'),

            -- Railroad 4
            (26, 'Short Line', NULL, 200, 100, 110, NULL, NULL, 'RAILROAD'),

            -- Dark Blue (House cost: 200)
            (27, 'Park Place', 37, 350, 175, 193, 200, 200, 'DARKBLUE'),
            (28, 'Boardwalk', 39, 400, 200, 220, 200, 200, 'DARKBLUE')
        ON CONFLICT (id) DO NOTHING;
    `)
    if err != nil {
        log.Fatal().Err(err).Msg("failed to insert properties into db")
    }

    _, err = tx.Exec(ctx, `
        CREATE TABLE Tiles (
        id INTEGER PRIMARY KEY,
        name TEXT NOT NULL,
        property_id INTEGER REFERENCES Property(id)
        )
        `)
    if err != nil {
        log.Fatal().Err(err).Msg("failed to create tiles table")
    }

    _, err = tx.Exec(ctx, `
        INSERT INTO Tiles (id, property_id, name)
        VALUES
            (0, NULL, 'Go'),
            -- Brown (House cost: 50)
            (1, 1, 'Mediterranean Avenue'),
            (3, 2, 'Baltic Avenue'),

            (2, NULL, 'Community Chest'),
            (4, NULL, 'Income Tax'),
            
            -- Railroad 1
            (5, 3, 'Reading Railroad'),

            -- Light Blue (House cost: 50)
            (6, 4, 'Oriental Avenue'),
            (8, 5, 'Vermont Avenue'),
            (9, 6, 'Connecticut Avenue'),

            (7, NULL, 'Chance'),
            
            (10, NULL, 'Just Visiting'),

            -- Pink (House cost: 100)
            (11, 7, 'St. Charles Place'),
            (12, 8, 'Electric Company'),
            (13, 9, 'States Avenue'),
            (14, 10, 'Virginia Avenue'),

            -- Railroad 2
            (15, 11, 'Pennsylvania Railroad'),

            -- Orange (House cost: 100)
            (16, 12, 'St. James Place'),
            (18, 13, 'Tennessee Avenue'),
            (19, 14, 'New York Avenue'),

            (17, NULL, 'Community Chest'),
            (20, NULL, 'Free Parking'),

            -- Red (House cost: 150)
            (21, 15, 'Kentucky Avenue'),
            (23, 16, 'Indiana Avenue'),
            (24, 17, 'Illinois Avenue'),

            (22, NULL, 'Chance'),

            -- Railroad 3
            (25, 18, 'B. & O. Railroad'),

            -- Yellow (House cost: 150)
            (26, 19, 'Atlantic Avenue'),
            (27, 20, 'Ventnor Avenue'),
            (28, 21, 'Water Works'),
            (29, 22, 'Marvin Gardens'),

            (30, NULL, 'Go To Jail'),

            -- Green (House cost: 200)
            (31, 23, 'Pacific Avenue'),
            (32, 24, 'North Carolina Avenue'),
            (34, 25, 'Pennsylvania Avenue'),

            (33, NULL, 'Community Chest'),

            -- Railroad 4
            (35, 26, 'Short Line'),

            (36, NULL, 'Chance'),

            -- Dark Blue (House cost: 200)
            (37, 27, 'Park Place'),
            (39, 28, 'Boardwalk'),

            (38, NULL, 'Luxury Tax')

        ON CONFLICT (id) DO NOTHING;
    `)
    if err != nil {
        log.Fatal().Err(err).Msg("failed to insert tiles into db")
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
