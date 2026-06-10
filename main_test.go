package main

import (
	"testing"
)

func TestSQLiteDatabasePath(t *testing.T) {
	tests := []struct {
		name           string
		dataSourceName string
		want           string
	}{
		{
			name:           "memory database",
			dataSourceName: ":memory:",
			want:           "",
		},
		{
			name:           "plain relative path",
			dataSourceName: "./data/app.db",
			want:           "./data/app.db",
		},
		{
			name:           "plain absolute path",
			dataSourceName: "/tmp/comelymd/comelymd.db",
			want:           "/tmp/comelymd/comelymd.db",
		},
		{
			name:           "file URI relative path",
			dataSourceName: "file:./data/app.db?cache=shared",
			want:           "./data/app.db",
		},
		{
			name:           "file URI absolute path",
			dataSourceName: "file:/tmp/comelymd.db?mode=rwc",
			want:           "/tmp/comelymd.db",
		},
		{
			name:           "file URI absolute path with authority-style slashes",
			dataSourceName: "file:///tmp/comelymd.db?mode=rwc",
			want:           "/tmp/comelymd.db",
		},
		{
			name:           "file URI memory database",
			dataSourceName: "file::memory:?cache=shared",
			want:           "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sqliteDatabasePath(tt.dataSourceName); got != tt.want {
				t.Fatalf("sqliteDatabasePath(%q) = %q, want %q", tt.dataSourceName, got, tt.want)
			}
		})
	}
}

func TestResolveDatabaseConfig(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    databaseConfig
		wantErr string
	}{
		{
			name: "default sqlite config",
			want: databaseConfig{
				driver:         "sqlite",
				dataSourceName: "./data/app.db",
				localPath:      "./data/app.db",
				logTarget:      "./data/app.db",
			},
		},
		{
			name: "explicit sqlite file URI",
			env: map[string]string{
				"DB_DRIVER": "sqlite",
				"DB_PATH":   "file:./data/app.db?cache=shared",
			},
			want: databaseConfig{
				driver:         "sqlite",
				dataSourceName: "file:./data/app.db?cache=shared",
				localPath:      "./data/app.db",
				logTarget:      "file:./data/app.db?cache=shared",
			},
		},
		{
			name: "sqlite memory database",
			env: map[string]string{
				"DB_PATH": ":memory:",
			},
			want: databaseConfig{
				driver:         "sqlite",
				dataSourceName: ":memory:",
				localPath:      "",
				logTarget:      ":memory:",
			},
		},
		{
			name: "libsql without auth token",
			env: map[string]string{
				"DB_DRIVER": "libsql",
				"DB_URL":    "libsql://example.turso.io",
			},
			want: databaseConfig{
				driver:         "libsql",
				dataSourceName: "libsql://example.turso.io",
				localPath:      "",
				logTarget:      "libsql://example.turso.io",
			},
		},
		{
			name: "libsql with auth token",
			env: map[string]string{
				"DB_DRIVER":     "libsql",
				"DB_URL":        "libsql://example.turso.io?tls=1",
				"DB_AUTH_TOKEN": "secret token",
			},
			want: databaseConfig{
				driver:         "libsql",
				dataSourceName: "libsql://example.turso.io?tls=1&authToken=secret+token",
				localPath:      "",
				logTarget:      "libsql://example.turso.io?tls=1",
			},
		},
		{
			name: "missing libsql URL",
			env: map[string]string{
				"DB_DRIVER": "libsql",
			},
			wantErr: "DB_URL is required when DB_DRIVER=libsql",
		},
		{
			name: "unsupported driver",
			env: map[string]string{
				"DB_DRIVER": "postgres",
			},
			wantErr: "unsupported DB_DRIVER \"postgres\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getenv := func(key string) string {
				return tt.env[key]
			}

			got, err := resolveDatabaseConfig(getenv)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("resolveDatabaseConfig() error = %v, want %q", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("resolveDatabaseConfig() unexpected error = %v", err)
			}

			if got != tt.want {
				t.Fatalf("resolveDatabaseConfig() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
