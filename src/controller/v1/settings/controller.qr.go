package settings

import (
	"net/http"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/labstack/echo/v4"
)

func QRCodeController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/qr-code/:code",
		Method:       "GET",
		ResponseType: horizon.QRResult{},
		Note:         "Decodes a QR code and returns the associated user information.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		code := ctx.Param("code")
		qr, err := service.QR.DecodeQR(context, &horizon.QRResult{
			Data: code,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to decode QR code: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, qr)
	})
}
