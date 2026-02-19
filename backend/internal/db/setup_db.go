package internaldb

import (
	"context"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)


func SetupDatabase(ctx context.Context, log zerolog.Logger) {
    monopoly_db_url_str := "postgres://postgres:<pass>@localhost:<port>/monopoly?sslmode=disable"
    monopoly_db_port := os.Getenv("POSTGRES_PORT")
    if monopoly_db_port == "" {
        monopoly_db_port = "1357"
    }
    postgres_password := os.Getenv("POSTGRES_PASSWORD")
    monopoly_db_url := strings.ReplaceAll(monopoly_db_url_str, "<pass>", postgres_password)
    monopoly_db_url = strings.ReplaceAll(monopoly_db_url, "<port>", monopoly_db_port)

    monopoly_default_db_url_str := "postgres://postgres:<pass>@localhost:<port>/postgres?sslmode=disable"

    monopoly_default_db_url := strings.ReplaceAll(monopoly_default_db_url_str, "<pass>", postgres_password)
    monopoly_default_db_url = strings.ReplaceAll(monopoly_default_db_url, "<port>", monopoly_db_port)

    db, err := pgx.Connect(ctx, monopoly_default_db_url)
    if err != nil {
        log.Fatal().Err(err).Msg("failed to connect to postgres database")
    }
    defer db.Close(ctx)

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
        return
    }

    log.Info().Msg("tables not found. creating tables...")

    tx, err := db.Begin(ctx)
    if err != nil {
        log.Fatal().Err(err).Msg("failed to begin transaction")
    }
    defer func() {
        if err != nil {
            tx.Rollback(ctx) // Rollback the transaction on error
            log.Error().Err(err).Msg("transaction rolled back due to error.")
        }
    }()


    // TODO: start adding tables here



    err = tx.Commit(ctx)
    if err != nil {
        log.Fatal().Err(err).Msg("failed to commit transaction")
    }
    log.Info().Msg("setup database transaction committed successfully.")
}

