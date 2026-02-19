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
    monopoly_db_url_str := "postgres://postgres:<pass>@localhost:<port>/monopoly?sslmode=disable"
    monopoly_db_port := os.Getenv("POSTGRES_PORT")
    if monopoly_db_port == "" {
        monopoly_db_port = "1357"
    }
    postgres_password := os.Getenv("POSTGRES_PASSWORD")
    monopoly_db_url := strings.ReplaceAll(monopoly_db_url_str, "<pass>", postgres_password)
    monopoly_db_url = strings.ReplaceAll(monopoly_db_url, "<port>", monopoly_db_port)

    db_pool, err := pgxpool.New(ctx, monopoly_db_url)
    if err != nil {
        retry_timeout := 1
        for i := 0; i < 5; i++ {
            log.Warn().Err(err).Int("retry_count",i).Int("retry_timeout (s)", retry_timeout).Msg("failed to connect to db")
            db_pool, err = pgxpool.New(ctx, monopoly_db_url)
            if err == nil {
                break
            }
            time.Sleep(time.Duration(retry_timeout) * time.Second)
            retry_timeout = retry_timeout * 2
        }
        return nil, err
    }

    return db_pool, nil
}


