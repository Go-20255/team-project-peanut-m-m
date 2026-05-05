package internaldb

import (
    "context"
    "os"
    "strings"
    "time"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/labstack/echo/v4"
    "github.com/rs/zerolog"
)

// TxMiddleware returns an Echo middleware that starts a SQL transaction for each HTTP request.
// The transaction is committed if the handler completes successfully, or rolled
// back if an error occurs. The transaction is stored in the Echo context under
// the key "tx" for use in downstream handlers.
func TxMiddleware(db *pgxpool.Pool) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            tx, err := db.BeginTx(c.Request().Context(), pgx.TxOptions{})
            if err != nil {
                return err
            }

            c.Set("tx", tx)

            if err := next(c); err != nil {
                _ = tx.Rollback(c.Request().Context())
                return err
            }

            if err := tx.Commit(c.Request().Context()); err != nil {
                return err
            }
            return nil
        }
    }
}

func CreateDbPoolConnection(ctx context.Context, log zerolog.Logger) (*pgxpool.Pool, error) {
    monopolyDbUrlStr := "postgres://postgres:<pass>@<host>:<port>/monopoly?sslmode=disable"
    monopolyDbHost := os.Getenv("POSTGRES_HOST")
    if monopolyDbHost == "" {
        monopolyDbHost = "localhost"
    }
    monopolyDbPort := os.Getenv("POSTGRES_PORT")
    if monopolyDbPort == "" {
        monopolyDbPort = "1357"
    }
    postgresPassword := os.Getenv("POSTGRES_PASSWORD")
    monopolyDbUrl := strings.ReplaceAll(monopolyDbUrlStr, "<pass>", postgresPassword)
    monopolyDbUrl = strings.ReplaceAll(monopolyDbUrl, "<host>", monopolyDbHost)
    monopolyDbUrl = strings.ReplaceAll(monopolyDbUrl, "<port>", monopolyDbPort)

    dbPool, err := pgxpool.New(ctx, monopolyDbUrl)
    if err != nil {
        retryTimeout := 1
        for i := 0; i < 5; i++ {
            log.Warn().Err(err).Int("retry_count", i).Int("retry_timeout (s)", retryTimeout).Msg("failed to connect to db")
            dbPool, err = pgxpool.New(ctx, monopolyDbUrl)
            if err == nil {
                break
            }
            time.Sleep(time.Duration(retryTimeout) * time.Second)
            retryTimeout = retryTimeout * 2
        }
        return nil, err
    }

    return dbPool, nil
}
