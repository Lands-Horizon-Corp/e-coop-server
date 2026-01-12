package v1

import (
	"net/http"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/horizon"
	"github.com/labstack/echo/v4"
)

func qRCodeController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/qr-code/:code",
		Method:       "GET",
		ResponseType: horizon.QRResult{},
		Note:         "Decodes a QR code and returns the associated user information.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		code := ctx.Param("code")
		qr, err := c.provider.Service.QR.DecodeQR(context, &horizon.QRResult{
			Data: code,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to decode QR code: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, qr)
	})
}
