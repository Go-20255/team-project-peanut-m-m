package common_handler

import (
	"monopoly-backend/util"
	"net/http"

	"github.com/labstack/echo/v4"
)

func HealthCheckHandler(c echo.Context) error {
    log := util.GetRequestLogger(c)
    log.Info().Msg("received health check request")
    return c.String(http.StatusOK, "api is alive")
}
