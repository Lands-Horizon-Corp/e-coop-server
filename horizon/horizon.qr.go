package horizon

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/rotisserie/eris"
	"go.uber.org/zap"
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
	config   *HorizonConfig
	log      *HorizonLog
	security *HorizonSecurity
}

func NewHorizonQR(
	config *HorizonConfig,
	log *HorizonLog,
	security *HorizonSecurity,
) (*HorizonQR, error) {
	return &HorizonQR{
		config:   config,
		log:      log,
		security: security,
	}, nil
}

func (hq *HorizonQR) Encode(data any) (*QRResult, error) {
	typeName := reflect.TypeOf(data).Name()
	dataBytes, err := json.Marshal(data)
	if err != nil {
		hq.log.Log(LogEntry{
			Category: CategoryQR,
			Level:    LevelError,
			Message:  "failed to marshal original data to JSON",
			Fields: []zap.Field{
				zap.String("type", typeName),
				zap.Any("data", data),
				zap.Error(err),
			},
		})
		return nil, eris.Wrap(err, "failed to marshal data")
	}
	var dataMap map[string]any
	if err := json.Unmarshal(dataBytes, &dataMap); err != nil {
		hq.log.Log(LogEntry{
			Category: CategoryQR,
			Level:    LevelError,
			Message:  "failed to unmarshal data into map[string]any",
			Fields: []zap.Field{
				zap.String("type", typeName),
				zap.ByteString("json", dataBytes),
				zap.Error(err),
			},
		})
		return nil, eris.Wrap(err, "failed to convert data to map")
	}
	wrapped := map[string]any{
		"type": typeName,
		"data": dataMap,
	}
	finalBytes, err := json.Marshal(wrapped)
	if err != nil {
		hq.log.Log(LogEntry{
			Category: CategoryQR,
			Level:    LevelError,
			Message:  "failed to marshal wrapped data",
			Fields: []zap.Field{
				zap.String("type", typeName),
				zap.Any("wrapped", wrapped),
				zap.Error(err),
			},
		})
		return nil, eris.Wrap(err, "failed to marshal wrapped data")
	}
	encryptedStr, err := hq.security.Encrypt(string(finalBytes))
	if err != nil {
		hq.log.Log(LogEntry{
			Category: CategoryQR,
			Level:    LevelError,
			Message:  "failed to encrypt marshaled data",
			Fields: []zap.Field{
				zap.String("type", typeName),
				zap.ByteString("json", finalBytes),
				zap.Error(err),
			},
		})
		return nil, eris.Wrap(err, "encryption failed")
	}
	decoded, err := hq.Decode(encryptedStr)
	if err != nil {
		hq.log.Log(LogEntry{
			Category: CategoryQR,
			Level:    LevelError,
			Message:  "failed to decode encrypted data",
			Fields: []zap.Field{
				zap.String("type", typeName),
				zap.String("encrypted_str", encryptedStr),
				zap.Error(err),
			},
		})
		return nil, eris.Wrap(err, "failed to decode encrypted data")
	}
	hq.log.Log(LogEntry{
		Category: CategoryQR,
		Level:    LevelInfo,
		Message:  "Successfully converted data to QR",
		Fields: []zap.Field{
			zap.String("qr_code", encryptedStr),
			zap.String("type", typeName),
			zap.Any("decoded_data", decoded),
		},
	})

	return &QRResult{
		QRCode: encryptedStr,
		Type:   typeName,
	}, nil
}

func (hq *HorizonQR) Decode(qrCodeBase64 string) (any, error) {
	decryptedStr, err := hq.security.Decrypt(qrCodeBase64)
	if err != nil {
		hq.log.Log(LogEntry{
			Category: CategoryQR,
			Level:    LevelError,
			Message:  "Decryption failed",
			Fields: []zap.Field{
				zap.String("qr_code_base64", qrCodeBase64),
				zap.Error(err),
			},
		})

		return nil, eris.Wrap(err, "decryption failed")
	}

	var wrapper struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal([]byte(decryptedStr), &wrapper); err != nil {
		hq.log.Log(LogEntry{
			Category: CategoryQR,
			Level:    LevelError,
			Message:  "Failed to parse wrapped data",
			Fields: []zap.Field{
				zap.String("decrypted_str", decryptedStr),
				zap.Error(err),
			},
		})
		return nil, eris.Wrap(err, "failed to parse wrapped data")
	}

	switch wrapper.Type {
	case "QRTransaction":
		var tx QRTransaction
		if err := json.Unmarshal(wrapper.Data, &tx); err != nil {
			return nil, eris.Wrap(err, "failed to unmarshal QRTransaction")
		}
		return tx, nil

	case "QRUser":
		var user QRUser
		if err := json.Unmarshal(wrapper.Data, &user); err != nil {
			return nil, eris.Wrap(err, "failed to unmarshal QRUser")
		}
		return user, nil

	case "QROrganization":
		var org QROrganization
		if err := json.Unmarshal(wrapper.Data, &org); err != nil {
			return nil, eris.Wrap(err, "failed to unmarshal QROrganization")
		}
		return org, nil

	case "QRMemberProfile":
		var profile QRMemberProfile
		if err := json.Unmarshal(wrapper.Data, &profile); err != nil {
			return nil, eris.Wrap(err, "failed to unmarshal QRMemberProfile")
		}
		return profile, nil

	case "QRInvitationLink":
		var link QRInvitationLink
		if err := json.Unmarshal(wrapper.Data, &link); err != nil {
			return nil, eris.Wrap(err, "failed to unmarshal QRInvitationLink")
		}
		return link, nil

	default:
		hq.log.Log(LogEntry{
			Category: CategoryQR,
			Level:    LevelError,
			Message:  "Unsupported QR type",
			Fields: []zap.Field{
				zap.String("type", wrapper.Type),
			},
		})
		return nil, eris.New(fmt.Sprintf("unsupported QR type: %s", wrapper.Type))
	}
}
