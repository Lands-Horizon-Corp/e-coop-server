package horizon

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"golang.org/x/crypto/argon2"
)

type SecurityImpl struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
	secret      []byte
	cache       *CacheImpl
}

func NewSecurityImpl(
	memory uint32,
	iterations uint32,
	parallelism uint8,
	saltLength uint32,
	keyLength uint32,
	secret []byte,
	cache *CacheImpl,

) *SecurityImpl {
	return &SecurityImpl{
		memory:      memory,
		iterations:  iterations,
		parallelism: parallelism,
		saltLength:  saltLength,
		keyLength:   keyLength,
		secret:      secret,
		cache:       cache,
	}
}

func (h *SecurityImpl) HashPassword(password string) (string, error) {
	salt := make([]byte, h.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, h.iterations, h.memory, h.parallelism, h.keyLength)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, h.memory, h.iterations, h.parallelism, b64Salt, b64Hash)
	return encodedHash, nil
}

func (h *SecurityImpl) Encrypt(ctx context.Context, data string, ttl time.Duration) (string, error) {
	token := uuid.New().String()
	key := fmt.Sprintf("X-SECURED:%s", token)
	if err := h.cache.Set(ctx, key, data, ttl); err != nil {
		return "", err
	}
	return token, nil
}

func (h *SecurityImpl) Decrypt(ctx context.Context, token string) (string, error) {
	key := fmt.Sprintf("X-SECURED:%s", token)
	data, err := h.cache.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if data == nil {
		return "", fmt.Errorf("token not found or expired")
	}
	return string(data), nil
}

func (h *SecurityImpl) VerifyPassword(hash string, password string) (bool, error) {
	vals := strings.Split(hash, "$")
	if len(vals) != 6 {
		return false, eris.New("the encoded hash is not in the correct format")
	}
	var version int
	_, err := fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return false, err
	}
	if version != argon2.Version {
		return false, eris.New("incompatible version of argon2")
	}
	var p struct {
		memory      uint32
		iterations  uint32
		parallelism uint8
	}

	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return false, err
	}
	hashed, err := base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return false, err
	}
	if subtle.ConstantTimeCompare(hashed, argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, h.keyLength)) == 1 {
		return true, nil
	}
	return false, nil
}

func (h *SecurityImpl) GenerateUUIDv5(name string) (string, error) {
	namespace := uuid.NameSpaceX500
	if name == "" {
		return "", errors.New("name cannot be empty")
	}

	uuid5 := uuid.NewSHA1(namespace, []byte(name))
	return uuid5.String(), nil
}

func (h *SecurityImpl) Firewall(ctx context.Context, callback func(ip, host string)) error {
	url := "https://raw.githubusercontent.com/hagezi/dns-blocklists/main/domains/ultimate.txt"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return eris.Wrap(err, "failed to create HTTP request")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return eris.Wrap(err, "failed to fetch HaGeZi Ultimate list")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return eris.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	scanner := bufio.NewScanner(resp.Body)
	domains := []string{}
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		domains = append(domains, line)
	}
	if err := scanner.Err(); err != nil {
		return eris.Wrap(err, "error reading response body")
	}
	domainToIPs := make(map[string][]string)
	var mu sync.Mutex
	var wg sync.WaitGroup
	count := 0
	for _, domain := range domains {
		wg.Add(1)
		go func(d string) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
			}
			ips, err := net.LookupIP(d)
			if err != nil {
				return
			}
			ipStrs := []string{}
			for _, ip := range ips {
				ipStrs = append(ipStrs, ip.String())
			}
			mu.Lock()
			domainToIPs[d] = ipStrs
			count++
			for _, ip := range ipStrs {
				callback(ip, d)
			}
			mu.Unlock()
		}(domain)
	}
	wg.Wait()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return nil
}
