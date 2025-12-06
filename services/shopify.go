package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"mgsearch/config"
)

type ShopifyService struct {
	apiKey     string
	apiSecret  string
	appURL     string
	scopes     string
	httpClient *http.Client
}

type accessTokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
}

func NewShopifyService(cfg *config.Config) *ShopifyService {
	return &ShopifyService{
		apiKey:    cfg.ShopifyAPIKey,
		apiSecret: cfg.ShopifyAPISecret,
		appURL:    cfg.ShopifyAppURL,
		scopes:    cfg.ShopifyScopes,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (s *ShopifyService) BuildInstallURL(shop string, state string, redirectURI string) (string, error) {
	if shop == "" {
		return "", fmt.Errorf("shop domain is required")
	}

	query := url.Values{}
	query.Set("client_id", s.apiKey)
	query.Set("scope", s.scopes)
	query.Set("redirect_uri", redirectURI) // Use redirectURI exactly as provided
	query.Set("state", state)
	query.Set("grant_options[]", "per-user")

	return fmt.Sprintf("https://%s/admin/oauth/authorize?%s", shop, query.Encode()), nil
}

func (s *ShopifyService) ExchangeAccessToken(ctx context.Context, shop string, code string) (string, error) {
	endpoint := fmt.Sprintf("https://%s/admin/oauth/access_token", shop)

	body, err := json.Marshal(map[string]string{
		"client_id":     s.apiKey,
		"client_secret": s.apiSecret,
		"code":          code,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal token request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token exchange failed with status %d", resp.StatusCode)
	}

	var tokenResp accessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	return tokenResp.AccessToken, nil
}

// ValidateHMAC validates the HMAC parameter on OAuth callbacks.
func (s *ShopifyService) ValidateHMAC(values url.Values) bool {
	hmacValue := values.Get("hmac")
	if hmacValue == "" {
		return false
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		if key == "hmac" || key == "signature" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var messageParts []string
	for _, key := range keys {
		messageParts = append(messageParts, fmt.Sprintf("%s=%s", key, values.Get(key)))
	}
	message := strings.Join(messageParts, "&")

	computed := computeHexHMAC([]byte(message), s.apiSecret)
	return hmac.Equal([]byte(hmacValue), []byte(computed))
}

// VerifyWebhookSignature ensures the webhook HMAC matches the body payload.
func (s *ShopifyService) VerifyWebhookSignature(signature string, body []byte) bool {
	if signature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(s.apiSecret))
	mac.Write(body)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expected))
}

func computeHexHMAC(message []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(message)
	return hex.EncodeToString(mac.Sum(nil))
}
