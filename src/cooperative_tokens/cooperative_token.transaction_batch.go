package cooperative_tokens

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src"
)

type TransactionBatchClaim struct {
	TransactionBatchID string `json:"transaction_batch_id"`
	BranchID           string `json:"branch_id"`
	OrganizationID     string `json:"organization_id"`
	UserID             string `json:"user_id"`
	jwt.RegisteredClaims
}

func (c TransactionBatchClaim) GetRegisteredClaims() *jwt.RegisteredClaims {
	return &c.RegisteredClaims
}

type TransactionBatchToken struct {
	Token *horizon.HorizonTokenService[TransactionBatchClaim]
}

func NewTransactionBatchToken(provider *src.Provider) (*TransactionBatchToken, error) {

	appName := provider.Service.Environment.GetString("APP_NAME", "")
	appToken := provider.Service.Environment.GetString("APP_TOKEN", "")

	token, err := provider.Service.Security.GenerateUUIDv5(context.Background(), appToken+"-transaction-batch")
	if err != nil {
		return nil, err
	}
	service := &horizon.HorizonTokenService[TransactionBatchClaim]{
		Name:   fmt.Sprintf("%s-%s", "X-SECURE-TRANSACTION-BATCH", appName),
		Secret: []byte(token),
	}
	return &TransactionBatchToken{Token: service}, nil
}
