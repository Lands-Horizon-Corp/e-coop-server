package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// GET /qr-code/:code
func (c *Controller) QRCode(ctx echo.Context) error {
	code := ctx.Param("code")
	qr, err := c.qr.Decode(code)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, qr)
}
