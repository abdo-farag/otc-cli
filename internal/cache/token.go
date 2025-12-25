package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type TokenCache struct {
	UnscopedToken string    `json:"unscoped_token"`
	IDToken       string    `json:"id_token"`
	RefreshToken  string    `json:"refresh_token"`
	ExpiresAt     time.Time `json:"expires_at"`
	Domain        string    `json:"domain"`
	Region        string    `json:"region"`
}

func GetCacheDir() string {
	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".otc-cli")
	os.MkdirAll(cacheDir, 0700)
	return cacheDir
}

func GetTokenPath() string {
	return filepath.Join(GetCacheDir(), "token.json")
}

func SaveToken(cache *TokenCache) error {
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(GetTokenPath(), data, 0600)
}

func LoadToken() (*TokenCache, error) {
	data, err := os.ReadFile(GetTokenPath())
	if err != nil {
		return nil, err
	}

	var cache TokenCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	// Check if expired
	if time.Now().After(cache.ExpiresAt) {
		return nil, fmt.Errorf("token expired")
	}

	return &cache, nil
}

func ClearToken() error {
	return os.Remove(GetTokenPath())
}
