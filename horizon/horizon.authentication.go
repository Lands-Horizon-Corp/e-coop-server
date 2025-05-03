package horizon

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rotisserie/eris"
)

var authExpiration = 16 * time.Hour

type Claim struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	ContactNumber string `json:"contact_number"`
	jwt.RegisteredClaims
}

type HorizonAuthentication struct {
	config   *HorizonConfig
	log      *HorizonLog
	security *HorizonSecurity
	otp      *HorizonOTP
	sms      *HorizonSMS
	smtp     *HorizonSMTP
}

func NewHorizonAuthentication(
	config *HorizonConfig,
	log *HorizonLog,
	security *HorizonSecurity,
	otp *HorizonOTP,
	sms *HorizonSMS,
	smtp *HorizonSMTP,
) (*HorizonAuthentication, error) {
	return &HorizonAuthentication{
		config:   config,
		log:      log,
		security: security,
		otp:      otp,
		sms:      sms,
		smtp:     smtp,
	}, nil
}

func (ha *HorizonAuthentication) SendSMTPOTP(value Claim) error {
	secure := ha.secured(value, "smtp")
	otp, err := ha.otp.Generate(secure)
	if err != nil {
		return err
	}
	err = ha.smtp.Send(&SMTPRequest{
		To:      value.Email,
		Subject: ha.config.AppName + "Email OTP Verification",
		Body:    `<h1>Hello {{ .email }}!</h1><p>This is your OTP {{ .otp }}.</p>`,
		Vars: &map[string]string{
			"email": value.Email,
			"otp":   otp,
		},
	})
	if err != nil {
		if delErr := ha.otp.Delete(secure); delErr != nil {
			return eris.Wrapf(err, "SMTP sending failed, and cleanup also failed: %v", delErr)
		}
		return eris.Wrap(err, "SMTP sending failed, OTP deleted")
	}
	return nil

}

func (ha *HorizonAuthentication) VerifySMTPOTP(value Claim, otp string) bool {
	secure := ha.secured(value, "smtp")
	valid, err := ha.otp.Verify(secure, otp)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return valid
}

func (ha *HorizonAuthentication) SendSMSOTP(value Claim) error {
	secure := ha.secured(value, "sms")
	otp, err := ha.otp.Generate(secure)
	if err != nil {
		return err
	}
	err = ha.sms.Send(&SMSRequest{
		To:      value.ContactNumber,
		Subject: ha.config.AppName + "SMS OTP Verification",
		Body:    `<h1>Hello {{ .email }}!</h1><p>This is your OTP {{ .otp }}.</p>`,
		Vars: &map[string]string{
			"email": value.Email,
			"otp":   otp,
		},
	})
	if err != nil {
		if delErr := ha.otp.Delete(secure); delErr != nil {
			return eris.Wrapf(err, "SMS sending failed, and cleanup also failed: %v", delErr)
		}
		return eris.Wrap(err, "SMS sending failed, OTP deleted")
	}
	return nil
}

func (ha *HorizonAuthentication) VerifySMSOTP(value Claim, otp string) bool {
	secure := ha.secured(value, "sms")
	valid, err := ha.otp.Verify(secure, otp)
	if err != nil {
		return false
	}
	return valid
}

func (ha *HorizonAuthentication) GenerateToken(value Claim) (string, error) {
	if authExpiration == 0 {
		authExpiration = 12 * time.Hour
	}
	claim := &Claim{
		ID:            value.ID,
		Email:         value.Email,
		ContactNumber: value.ContactNumber,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    ha.config.AppName,
			Subject:   value.ID,
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(authExpiration)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	signedToken, err := token.SignedString([]byte(ha.config.AppToken))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString([]byte(signedToken)), nil
}

func (ha *HorizonAuthentication) VerifyToken(tokenString string) (*Claim, error) {
	decode, err := base64.StdEncoding.DecodeString(tokenString)
	if err != nil {
		return nil, eris.Wrap(err, "invalid token")
	}
	token, err := jwt.ParseWithClaims(string(decode), &Claim{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, eris.Wrap(jwt.ErrSignatureInvalid, "unexpected signing method")
		}
		return []byte(ha.config.AppToken), nil
	})
	if err != nil {
		return nil, eris.Wrap(err, "failed to parse token")
	}
	if claims, ok := token.Claims.(*Claim); ok && token.Valid {
		if claims.Issuer != ha.config.AppName {
			return nil, eris.New("invalid token issuer")
		}
		return claims, nil
	}
	return nil, eris.New("invalid token")
}

func (ha *HorizonAuthentication) secured(value Claim, reason string) string {
	generated := value.Email + value.ContactNumber + value.ID + reason
	val := ha.security.Hash(generated + ha.config.AppName + "auth")
	return string(val)
}

/*
	// getting current user logged in
	// check if logged in on other device. prevet login
	// can log in and compare password
	// can send otp confirmation (this will be used for email verification, contact verification, continue to secured site, and allowing itself)
	// can send otp confirmation (this will be used for email verification, contact verification, continue to secured site, and allowing itself)
*/
