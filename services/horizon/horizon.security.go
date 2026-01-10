package horizon

import (
	"bufio"
	"context"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"golang.org/x/crypto/argon2"
)

type SecurityService interface {
	GenerateUUID(ctx context.Context) (string, error)

	HashPassword(ctx context.Context, password string) (string, error)

	VerifyPassword(ctx context.Context, hash, password string) (bool, error)

	Encrypt(ctx context.Context, plaintext string) (string, error)

	Decrypt(ctx context.Context, ciphertext string) (string, error)

	GenerateUUIDv5(ctx context.Context, name string) (string, error)

	Firewall(ctx context.Context, callback func(ip, host string)) error
}

type Security struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
	secret      []byte
	mu          sync.RWMutex
}

func NewSecurityService(
	memory uint32,
	iterations uint32,
	parallelism uint8,
	saltLength uint32,
	keyLength uint32,
	secret []byte,
) SecurityService {
	return &Security{
		memory:      memory,
		iterations:  iterations,
		parallelism: parallelism,
		saltLength:  saltLength,
		keyLength:   keyLength,
		secret:      secret,
		mu:          sync.RWMutex{},
	}
}

func (h *Security) Decrypt(_ context.Context, s string) (string, error) {
	data, err := base64.RawStdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	key := []byte(h.secret)
	out := make([]byte, len(data))
	for i := range data {
		out[i] = data[i] ^ key[i%len(key)]
	}
	return string(out), nil
}

func (h *Security) Encrypt(_ context.Context, s string) (string, error) {
	key := []byte(h.secret)
	out := make([]byte, len(s))
	for i := range s {
		out[i] = s[i] ^ key[i%len(key)]
	}
	return base64.RawStdEncoding.EncodeToString(out), nil
}

func (h *Security) GenerateUUID(_ context.Context) (string, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (h *Security) HashPassword(_ context.Context, password string) (string, error) {
	salt, err := handlers.GenerateRandomBytes(h.saltLength)
	if err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, h.iterations, h.memory, h.parallelism, h.keyLength)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, h.memory, h.iterations, h.parallelism, b64Salt, b64Hash)
	return encodedHash, nil
}

func (h *Security) VerifyPassword(_ context.Context, hash string, password string) (bool, error) {
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

	otherHash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, h.keyLength)

	if subtle.ConstantTimeCompare(hashed, otherHash) == 1 {
		return true, nil
	}
	return false, nil
}

func (h *Security) GenerateUUIDv5(_ context.Context, name string) (string, error) {
	namespace := uuid.NameSpaceX500
	if name == "" {
		return "", errors.New("name cannot be empty")
	}

	uuid5 := uuid.NewSHA1(namespace, []byte(name))
	return uuid5.String(), nil
}

func (h *Security) Firewall(ctx context.Context, callback func(ip, host string)) error {
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
	var mu sync.Mutex // To safely write to the map
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
