package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
)

func (c *Controller) QRCodeController() {
	req := c.provider.Service.Request
	req.RegisterRoute(horizon.Route{
		Route:    "/qr-code/:code",
		Method:   "GET",
		Response: "TUser",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		code := ctx.Param("code")
		qr, err := c.provider.Service.QR.DecodeQR(context, &horizon.QRResult{
			Data: code,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, qr)
	})
}
