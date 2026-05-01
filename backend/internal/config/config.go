package config

import (
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI        string
	MongoDatabase   string
	ServerPort      string
	FrontendOrigins []string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		MongoURI:        valueOrDefault(os.Getenv("MONGODB_URI"), "mongodb://127.0.0.1:27017/admin"),
		MongoDatabase:   valueOrDefault(os.Getenv("MONGODB_DATABASE"), "admin"),
		ServerPort:      valueOrDefault(os.Getenv("SERVER_PORT"), "8080"),
		FrontendOrigins: frontendOriginsFromEnv(),
	}

	return cfg, nil
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func frontendOriginsFromEnv() []string {
	if origins := parseOrigins(os.Getenv("FRONTEND_ORIGINS")); len(origins) > 0 {
		return origins
	}

	if legacyOrigin := strings.TrimSpace(os.Getenv("FRONTEND_ORIGIN")); legacyOrigin != "" {
		return expandLegacyOrigin(legacyOrigin)
	}

	return []string{
		"http://localhost:5173",
		"http://127.0.0.1:5173",
	}
}

func parseOrigins(value string) []string {
	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))

	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin == "" {
			continue
		}
		origins = append(origins, origin)
	}

	return origins
}

func expandLegacyOrigin(origin string) []string {
	parsed, err := url.Parse(origin)
	if err != nil {
		return []string{origin}
	}

	host := parsed.Hostname()
	if host != "localhost" && host != "127.0.0.1" {
		return []string{origin}
	}

	port := parsed.Port()
	if port == "" {
		return []string{origin}
	}

	origins := []string{
		parsed.Scheme + "://localhost:" + port,
		parsed.Scheme + "://127.0.0.1:" + port,
	}

	if host == "127.0.0.1" {
		origins[0], origins[1] = origins[1], origins[0]
	}

	return origins
}
