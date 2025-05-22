package cooperative_tokens

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src"
)

type UserClaim struct {
	UserID        string `json:"user_id"`
	Email         string `json:"email"`
	ContactNumber string `json:"contact_number"`
	Password      string `json:"password"`
	Username      string `json:"username"`
	jwt.RegisteredClaims
}

func (c UserClaim) GetRegisteredClaims() *jwt.RegisteredClaims {
	return &c.RegisteredClaims
}

type UserToken struct {
	Token *horizon.HorizonTokenService[UserClaim]
}

func NewUserToken(provider *src.Provider) (*UserToken, error) {
	appName := provider.Service.Environment.GetString("APP_NAME", "")
	appToken := provider.Service.Environment.GetString("APP_TOKEN", "")

	token, err := provider.Service.Security.GenerateUUIDv5(context.Background(), appToken+"-user")
	if err != nil {
		return nil, err
	}
	service := &horizon.HorizonTokenService[UserClaim]{
		Name:   fmt.Sprintf("%s-%s", "X-SECURE-USER", appName),
		Secret: []byte(token),
	}
	return &UserToken{Token: service}, nil
}
