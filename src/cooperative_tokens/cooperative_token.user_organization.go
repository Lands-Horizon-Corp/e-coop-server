package cooperative_tokens

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src"
)

type UserOrganizatonClaim struct {
	UserOrganizatonID string `json:"user_organization_id"`
	UserID            string `json:"user_id"`
	BranchID          string `json:"branch_id"`
	OrganizationID    string `json:"organization_id"`
	jwt.RegisteredClaims
}

func (c UserOrganizatonClaim) GetRegisteredClaims() *jwt.RegisteredClaims {
	return &c.RegisteredClaims
}

type UserOrganizatonToken struct {
	Token *horizon.HorizonTokenService[UserOrganizatonClaim]
}

func NewUserOrganizatonToken(provider *src.Provider) (*UserOrganizatonToken, error) {
	appName := provider.Service.Environment.GetString("APP_NAME", "")
	appToken := provider.Service.Environment.GetString("APP_TOKEN", "")

	token, err := provider.Service.Security.GenerateUUIDv5(context.Background(), appToken+"-user-organization")
	if err != nil {
		return nil, err
	}

	service := &horizon.HorizonTokenService[UserOrganizatonClaim]{
		Name:   fmt.Sprintf("%s-%s", "X-SECURE-USER-ORGANIZATION", appName),
		Secret: []byte(token),
	}
	return &UserOrganizatonToken{Token: service}, nil
}
