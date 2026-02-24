package config

import "testing"

func TestLoadDefaultsAndEnvOverrides(t *testing.T) {
	t.Setenv("DB_HOST", "db.local")
	t.Setenv("DB_PORT", "15432")
	t.Setenv("PUBLIC_API_TOKEN", "pub")
	t.Setenv("ADMIN_API_TOKEN", "adm")
	t.Setenv("RESERVATION_ITEM_MAX_QUANTITY", "99")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Database.Host != "db.local" || cfg.Database.Port != 15432 {
		t.Fatalf("unexpected db config: %+v", cfg.Database)
	}
	if cfg.Auth.PublicToken != "pub" || cfg.Auth.AdminToken != "adm" {
		t.Fatalf("unexpected auth config: %+v", cfg.Auth)
	}
	if cfg.ReservationItemSettings.MaxQuantity != 99 {
		t.Fatalf("unexpected max quantity: %d", cfg.ReservationItemSettings.MaxQuantity)
	}
}

func TestGetEnvIntFallbackAndDSN(t *testing.T) {
	t.Setenv("SERVER_PORT", "not-an-int")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Fatalf("expected default server port 8080, got %d", cfg.Server.Port)
	}

	dsn := cfg.GetDatabaseDSN()
	if dsn == "" {
		t.Fatalf("expected non-empty dsn")
	}
}
