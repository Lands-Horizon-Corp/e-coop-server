package horizon

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type QRTransaction struct {
	AccountIDFrom       string  `json:"account_id_from"`
	AccountIDTo         string  `json:"account_id_to"`
	UserIDFrom          string  `json:"user_id_from"`
	UserIDTo            string  `json:"user_id_to"`
	MemberProfileIDFrom string  `json:"member_profile_id_from"`
	MemberProfileIDTo   string  `json:"member_profile_id_to"`
	OrganizationIDFrom  string  `json:"organization_id_from"`
	OrganizationIDTo    string  `json:"organization_id_to"`
	Amount              float32 `json:"amount"`
}

type QRUser struct {
	UserID     string `json:"user_id"`
	Name       string `json:"name"`
	Lastname   string `json:"lastname"`
	Firstname  string `json:"firstname"`
	Middlename string `json:"middlename"`
}

type QROrganization struct {
	OrganizationID string `json:"organization_id"`
	Name           string `json:"name"`
}

type QRMemberProfile struct {
	UserID          string `json:"user_id"`
	MemberProfileID string `json:"member_profile_id"`
	OrganizationID  string `json:"organization_id"`
	BranchID        string `json:"branch_id"`
}

type QRInvitationLink struct {
	UserID         string `json:"user_id"`
	InvitationID   string `json:"invitation_id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	MediaImageLink string `json:"media_image_link"`
	OrganizationID string `json:"organization_id"`
}

type QRResult struct {
	QRCode string `json:"qr_code"`
	Type   string `json:"type"`
}

type HorizonQR struct {
	config *HorizonConfig
}

func NewHorizonQR(config *HorizonConfig) (*HorizonQR, error) {
	return &HorizonQR{
		config: config,
	}, nil
}

func (hq *HorizonQR) Encode(data any) (*QRResult, error) {
	typeName := reflect.TypeOf(data).Name()
	wrapped := map[string]any{
		"type": typeName,
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	var dataMap map[string]any
	if err := json.Unmarshal(dataBytes, &dataMap); err != nil {
		return nil, fmt.Errorf("failed to remarshal to map: %w", err)
	}
	wrapped["data"] = dataMap

	finalBytes, err := json.Marshal(wrapped)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal wrapped data: %w", err)
	}
	key := hq.config.AppToken
	encryptedStr, err := Encrypt(key, string(finalBytes))
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}
	return &QRResult{
		QRCode: encryptedStr,
		Type:   typeName,
	}, nil
}

func (hq *HorizonQR) Decode(qrCodeBase64 string) (any, error) {
	key := hq.config.AppToken
	decryptedStr, err := Decrypt(key, qrCodeBase64)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}
	var wrapper struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal([]byte(decryptedStr), &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse wrapped data: %w", err)
	}
	switch wrapper.Type {
	case "QRTransaction":
		var tx QRTransaction
		if err := json.Unmarshal(wrapper.Data, &tx); err != nil {
			return nil, err
		}
		return tx, nil

	case "QRUser":
		var user QRUser
		if err := json.Unmarshal(wrapper.Data, &user); err != nil {
			return nil, err
		}
		return user, nil

	case "QROrganization":
		var org QROrganization
		if err := json.Unmarshal(wrapper.Data, &org); err != nil {
			return nil, err
		}
		return org, nil

	case "QRMemberProfile":
		var profile QRMemberProfile
		if err := json.Unmarshal(wrapper.Data, &profile); err != nil {
			return nil, err
		}
		return profile, nil

	case "QRInvitationLink":
		var link QRInvitationLink
		if err := json.Unmarshal(wrapper.Data, &link); err != nil {
			return nil, err
		}
		return link, nil

	default:
		return nil, fmt.Errorf("unsupported QR type: %s", wrapper.Type)
	}
}
