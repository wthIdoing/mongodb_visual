package config

import (
	"reflect"
	"testing"
)

func TestLoadUsesDefaultFrontendOrigins(t *testing.T) {
	t.Setenv("MONGODB_URI", "mongodb://localhost:27017")
	t.Setenv("MONGODB_DATABASE", "admin")
	t.Setenv("FRONTEND_ORIGINS", "")
	t.Setenv("FRONTEND_ORIGIN", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	want := []string{
		"http://localhost:5173",
		"http://127.0.0.1:5173",
	}
	if !reflect.DeepEqual(cfg.FrontendOrigins, want) {
		t.Fatalf("FrontendOrigins = %v, want %v", cfg.FrontendOrigins, want)
	}
}

func TestLoadUsesConfiguredFrontendOrigins(t *testing.T) {
	t.Setenv("MONGODB_URI", "mongodb://localhost:27017")
	t.Setenv("MONGODB_DATABASE", "admin")
	t.Setenv("FRONTEND_ORIGINS", " http://localhost:5173, http://127.0.0.1:5173 ")
	t.Setenv("FRONTEND_ORIGIN", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	want := []string{
		"http://localhost:5173",
		"http://127.0.0.1:5173",
	}
	if !reflect.DeepEqual(cfg.FrontendOrigins, want) {
		t.Fatalf("FrontendOrigins = %v, want %v", cfg.FrontendOrigins, want)
	}
}

func TestLoadUsesSingleConfiguredFrontendOrigin(t *testing.T) {
	t.Setenv("MONGODB_URI", "mongodb://localhost:27017")
	t.Setenv("MONGODB_DATABASE", "admin")
	t.Setenv("FRONTEND_ORIGINS", "http://localhost:4173")
	t.Setenv("FRONTEND_ORIGIN", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	want := []string{"http://localhost:4173"}
	if !reflect.DeepEqual(cfg.FrontendOrigins, want) {
		t.Fatalf("FrontendOrigins = %v, want %v", cfg.FrontendOrigins, want)
	}
}

func TestLoadExpandsLegacyFrontendOrigin(t *testing.T) {
	t.Setenv("MONGODB_URI", "mongodb://localhost:27017")
	t.Setenv("MONGODB_DATABASE", "admin")
	t.Setenv("FRONTEND_ORIGINS", "")
	t.Setenv("FRONTEND_ORIGIN", "http://localhost:4173")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	want := []string{
		"http://localhost:4173",
		"http://127.0.0.1:4173",
	}
	if !reflect.DeepEqual(cfg.FrontendOrigins, want) {
		t.Fatalf("FrontendOrigins = %v, want %v", cfg.FrontendOrigins, want)
	}
}
