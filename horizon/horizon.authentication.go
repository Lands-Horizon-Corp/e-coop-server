package horizon

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

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

func NewHorizonAuthentication(cfg *HorizonConfig, log *HorizonLog,
	sec *HorizonSecurity, otp *HorizonOTP, sms *HorizonSMS, smtp *HorizonSMTP) *HorizonAuthentication {
	return &HorizonAuthentication{config: cfg, log: log, security: sec, otp: otp, sms: sms, smtp: smtp}
}

func (ha *HorizonAuthentication) GetUserFromToken(c echo.Context) (*Claim, error) {
	cookie, err := c.Cookie(ha.config.AppTokenName)
	if err != nil {
		_ = ha.CleanToken(c)
		return nil, eris.New("authentication token not found")
	}
	rawToken := cookie.Value
	if rawToken == "" {
		_ = ha.CleanToken(c)
		return nil, eris.New("authentication token is empty")
	}

	claim, err := ha.VerifyToken(rawToken)
	if err != nil {
		_ = ha.CleanToken(c)
		return nil, eris.Wrap(err, "invalid or expired authentication token")
	}
	return claim, nil
}

func (ha *HorizonAuthentication) SetToken(c echo.Context, claims Claim) error {
	tok, err := ha.GenerateToken(claims)
	if err != nil {
		return eris.Wrap(err, "GenerateToken failed")
	}
	cookie := &http.Cookie{
		Name:     ha.config.AppTokenName,
		Value:    tok,
		Path:     "/",
		Expires:  time.Now().Add(authExpiration),
		HttpOnly: true,
		Secure:   c.Request().TLS != nil,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)
	return nil
}

func (ha *HorizonAuthentication) CleanToken(c echo.Context) error {
	cookie := &http.Cookie{
		Name:     ha.config.AppTokenName,
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Secure:   c.Request().TLS != nil,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)
	return nil
}

func (ha *HorizonAuthentication) GenerateToken(c Claim) (string, error) {
	now := time.Now()
	// Default timings
	if c.NotBefore == nil {
		c.NotBefore = jwt.NewNumericDate(now)
	}
	if c.IssuedAt == nil {
		c.IssuedAt = jwt.NewNumericDate(now)
	}
	if c.ExpiresAt == nil {
		c.ExpiresAt = jwt.NewNumericDate(now.Add(authExpiration))
	}
	if c.Subject == "" {
		c.Subject = c.ID
	}

	c.Issuer = ha.config.AppTokenName
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, &c)
	signed, err := t.SignedString([]byte(ha.config.AppToken))
	if err != nil {
		return "", eris.Wrap(err, "signing token failed")
	}
	// Base64-encode the JWT so we can safely transmit it
	return base64.StdEncoding.EncodeToString([]byte(signed)), nil
}

func (ha *HorizonAuthentication) VerifyToken(encoded string) (*Claim, error) {
	// Base64 decode
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, eris.Wrap(err, "invalid base64 token")
	}
	// Parse
	tok, err := jwt.ParseWithClaims(string(raw), &Claim{}, func(tkn *jwt.Token) (any, error) {
		if _, ok := tkn.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, eris.Wrap(jwt.ErrSignatureInvalid, "unexpected signing method")
		}
		return []byte(ha.config.AppToken), nil
	})
	if err != nil {
		return nil, eris.Wrap(err, "token parse failed")
	}
	claims, ok := tok.Claims.(*Claim)
	if !ok || !tok.Valid || claims.Issuer != ha.config.AppTokenName {
		return nil, eris.New("invalid token")
	}
	return claims, nil
}

func (ha *HorizonAuthentication) GenerateSMTPLink(baseURL string, c Claim) (string, error) {
	subj := ha.config.AppTokenName + " Email Verification"
	rndr := func(f string, d map[string]string) (string, error) {
		return renderHTMLTemplate(f, d)
	}
	sndr := func(r any) error {
		req := r.(map[string]interface{})
		return ha.smtp.Send(&SMTPRequest{To: req["To"].(string), Subject: req["Subject"].(string), Body: req["Body"].(string)})
	}
	return ha.generateLink(baseURL, c, "email-forgot-password.html", rndr, sndr, subj)
}

func (ha *HorizonAuthentication) GenerateSMSLink(baseURL string, c Claim) (string, error) {
	subj := ha.config.AppTokenName + " SMS Verification"
	rndr := func(f string, d map[string]string) (string, error) {
		return renderTextTemplate(f, d)
	}
	sndr := func(r any) error {
		req := r.(map[string]interface{})
		return ha.sms.Send(&SMSRequest{To: req["To"].(string), Subject: req["Subject"].(string), Body: req["Body"].(string)})
	}
	return ha.generateLink(baseURL, c, "sms-forgot-password.txt", rndr, sndr, subj)
}

func (ha *HorizonAuthentication) ValidateLink(input string) (*Claim, error) {
	// Extract token suffix
	if idx := strings.LastIndex(input, "/"); idx != -1 {
		input = input[idx+1:]
	}
	tok, err := url.PathUnescape(input)
	if err != nil {
		return nil, eris.Wrap(err, "invalid link encoding")
	}
	return ha.VerifyToken(tok)
}

func (ha *HorizonAuthentication) Password(pw string) (string, error) {
	enc := base64.StdEncoding.EncodeToString([]byte(pw))
	return ha.security.PasswordHash(enc)
}

func (ha *HorizonAuthentication) VerifyPassword(hash, pw string) bool {
	enc := base64.StdEncoding.EncodeToString([]byte(pw))
	ok, _ := ha.security.VerifyPassword(hash, enc)
	return ok
}

func (ha *HorizonAuthentication) SendSMTPOTP(c Claim) error {
	key := ha.secureKey(c, "smtp")
	otp, err := ha.otp.Generate(key)
	if err != nil {
		return err
	}
	body, err := renderHTMLTemplate("email-otp.html", map[string]string{"AppTokenName": ha.config.AppTokenName, "otp": otp})
	if err != nil {
		return eris.Wrap(err, "OTP template failed")
	}
	if err := ha.smtp.Send(&SMTPRequest{To: c.Email, Subject: ha.config.AppTokenName + " Email OTP Verification", Body: body}); err != nil {
		ha.otp.Delete(key)
		return eris.Wrap(err, "SMTP OTP send failed")
	}
	return nil
}

func (ha *HorizonAuthentication) VerifySMTPOTP(c Claim, otp string) bool {
	ok, _ := ha.otp.Verify(ha.secureKey(c, "smtp"), otp)
	return ok
}

func (ha *HorizonAuthentication) SendSMSOTP(c Claim) error {
	key := ha.secureKey(c, "sms")
	otp, err := ha.otp.Generate(key)
	if err != nil {
		return err
	}
	body, err := renderTextTemplate("sms-otp.txt", map[string]string{"AppTokenName": ha.config.AppTokenName, "otp": otp})
	if err != nil {
		return eris.Wrap(err, "OTP template failed")
	}
	if err := ha.sms.Send(&SMSRequest{To: c.ContactNumber, Subject: ha.config.AppTokenName + " SMS OTP Verification", Body: body}); err != nil {
		ha.otp.Delete(key)
		return eris.Wrap(err, "SMS OTP send failed")
	}
	return nil
}

func (ha *HorizonAuthentication) VerifySMSOTP(c Claim, otp string) bool {
	ok, _ := ha.otp.Verify(ha.secureKey(c, "sms"), otp)
	return ok
}

func (ha *HorizonAuthentication) secureKey(c Claim, channel string) string {
	base := c.Email + c.ContactNumber + c.ID + channel
	return string(ha.security.Hash(base + ha.config.AppTokenName + "auth"))
}

func (ha *HorizonAuthentication) generateLink(baseURL string, c Claim, tplFile string, render func(string, map[string]string) (string, error), send func(any) error, subject string) (string, error) {
	c.ExpiresAt = jwt.NewNumericDate(time.Now().Add(tokenLinkValidity))
	token, err := ha.GenerateToken(c)
	if err != nil {
		return "", eris.Wrap(err, "token generation failed")
	}

	esc := url.PathEscape(token)
	link := fmt.Sprintf("%s/%s", strings.TrimRight(baseURL, "/"), esc)

	body, err := render(tplFile, map[string]string{
		"AppTokenName":  ha.config.AppTokenName,
		"email":         c.Email,
		"contactnumber": c.ContactNumber,
		"link":          link,
	})
	if err != nil {
		return "", eris.Wrap(err, "template rendering failed")
	}

	req := map[string]interface{}{
		"To":      c.Email,
		"Subject": subject,
		"Body":    body,
	}
	if err := send(req); err != nil {
		return "", eris.Wrap(err, "sending failed")
	}
	return link, nil
}

func renderHTMLTemplate(filename string, data map[string]string) (string, error) {
	t, err := template.ParseFiles(filepath.Join("template", filename))
	if err != nil {
		return "", eris.Wrap(err, "parse HTML template failed")
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", eris.Wrap(err, "execute HTML template failed")
	}
	return buf.String(), nil
}

func renderTextTemplate(filename string, data map[string]string) (string, error) {
	t, err := template.ParseFiles(filepath.Join("template", filename))
	if err != nil {
		return "", eris.Wrap(err, "parse text template failed")
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", eris.Wrap(err, "execute text template failed")
	}
	return buf.String(), nil
}
