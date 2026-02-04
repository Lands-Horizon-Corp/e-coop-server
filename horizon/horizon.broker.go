package horizon

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rotisserie/eris"
)

type MessageBrokerImpl struct {
	host      string
	port      int
	appID     string
	appKey    string
	appSecret string
	appClient string
	http      *http.Client
}

func NewSoketiPublisherImpl(
	host string,
	port int,
	appID, appKey, appSecret, appClient string,
) *MessageBrokerImpl {
	return &MessageBrokerImpl{
		host:      host,
		port:      port,
		appID:     appID,
		appKey:    appKey,
		appSecret: appSecret,
		appClient: appClient,
		http: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (s *MessageBrokerImpl) Dispatch(channels []string, payload any) error {
	body := map[string]any{
		"name":     s.appClient,
		"channels": channels,
		"data":     payload,
	}
	return s.send(body)
}

func (s *MessageBrokerImpl) Publish(channel string, payload any) error {
	body := map[string]any{
		"name":     s.appClient,
		"channels": []string{channel},
		"data":     payload,
	}
	return s.send(body)
}

func (s *MessageBrokerImpl) send(body map[string]any) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return eris.Wrap(err, "failed to marshal payload")
	}

	path := fmt.Sprintf("/apps/%s/events", s.appID)
	query := s.sign(path, jsonBody)

	url := fmt.Sprintf(
		"http://%s:%d%s?%s",
		s.host,
		s.port,
		path,
		query,
	)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return eris.Wrap(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.http.Do(req)
	if err != nil {
		return eris.Wrap(err, "failed to send event to soketi")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return eris.Errorf("soketi returned status %d", resp.StatusCode)
	}

	return nil
}

func (s *MessageBrokerImpl) sign(path string, body []byte) string {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	hash := sha256.Sum256(body)
	bodyMD5 := hex.EncodeToString(hash[:])

	query := fmt.Sprintf(
		"auth_key=%s&auth_timestamp=%s&auth_version=1.0&body_md5=%s",
		s.appKey,
		timestamp,
		bodyMD5,
	)

	stringToSign := fmt.Sprintf("POST\n%s\n%s", path, query)

	mac := hmac.New(sha256.New, []byte(s.appSecret))
	mac.Write([]byte(stringToSign))
	signature := hex.EncodeToString(mac.Sum(nil))

	return query + "&auth_signature=" + signature
}
