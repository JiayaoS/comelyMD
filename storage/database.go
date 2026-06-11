package storage

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

type Page struct {
	ID        string
	Markdown  string
	HTML      string
	IsBurn    bool
	ExpiresAt *time.Time
	Password  string
	CreatedAt time.Time
}

var DB *sql.DB

// InitDB opens the configured database, applies schema migrations, and starts expiration cleanup.
func InitDB(driver, dataSourceName string) {
	var err error
	DB, err = sql.Open(driver, dataSourceName)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	if driver == "libsql" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := DB.PingContext(ctx); err != nil {
			log.Fatalf("database connectivity check failed: %v", err)
		}
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS pages (
		id TEXT PRIMARY KEY,
		markdown TEXT NOT NULL,
		html TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = DB.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("failed to create pages table: %v", err)
	}

	// Add columns introduced by newer versions; ignore duplicate-column errors.
	_, _ = DB.Exec("ALTER TABLE pages ADD COLUMN is_burn BOOLEAN DEFAULT FALSE;")
	_, _ = DB.Exec("ALTER TABLE pages ADD COLUMN expires_at DATETIME;")
	_, _ = DB.Exec("ALTER TABLE pages ADD COLUMN password TEXT;")

	// Periodically remove expired pages.
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			DB.Exec("DELETE FROM pages WHERE expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP")
		}
	}()
}

// SavePage stores a Markdown page and returns its public ID and optional password.
func SavePage(markdown, html string, isBurn bool, expireDuration time.Duration, withPassword bool) (string, string, error) {
	var id string

	// Try a few random IDs before reporting a collision failure.
	for i := 0; i < 5; i++ {
		newID, err := GenerateID(8)
		if err != nil {
			return "", "", err
		}
		var exists bool
		err = DB.QueryRow("SELECT EXISTS(SELECT 1 FROM pages WHERE id = ?)", newID).Scan(&exists)
		if err != nil {
			return "", "", err
		}
		if !exists {
			id = newID
			break
		}
	}

	if id == "" {
		return "", "", errors.New("failed to allocate unique page ID")
	}

	var nullPwd sql.NullString
	if withPassword {
		nullPwd.Valid = true
		nullPwd.String, _ = GeneratePassword(4)
	}

	var nullExpires sql.NullTime
	if expireDuration > 0 {
		nullExpires.Valid = true
		nullExpires.Time = time.Now().UTC().Add(expireDuration)
	}

	_, err := DB.Exec(`INSERT INTO pages (id, markdown, html, is_burn, expires_at, password) VALUES (?, ?, ?, ?, ?, ?)`,
		id, markdown, html, isBurn, nullExpires, nullPwd)

	return id, nullPwd.String, err
}

// GetPage loads a page and enforces expiration before returning it.
func GetPage(id string) (*Page, error) {
	var p Page
	var rawExpires sql.NullString
	var rawPassword sql.NullString
	var rawCreatedAt sql.NullString

	// Cast nullable timestamps to text to avoid driver-specific scan issues.
	err := DB.QueryRow("SELECT id, markdown, html, is_burn, CAST(expires_at AS TEXT), password, CAST(created_at AS TEXT) FROM pages WHERE id = ?", id).
		Scan(&p.ID, &p.Markdown, &p.HTML, &p.IsBurn, &rawExpires, &rawPassword, &rawCreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("page not found")
		}
		return nil, err
	}

	if rawPassword.Valid {
		p.Password = rawPassword.String
	}

	// Parse the stored expiration timestamp when present.
	if rawExpires.Valid && rawExpires.String != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", rawExpires.String[:19]); err == nil {
			p.ExpiresAt = &t
		} else if t2, err2 := time.Parse(time.RFC3339, rawExpires.String); err2 == nil {
			p.ExpiresAt = &t2
		}
	}

	// Parse the stored creation timestamp when present.
	if rawCreatedAt.Valid && rawCreatedAt.String != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", rawCreatedAt.String[:19]); err == nil {
			p.CreatedAt = t
		} else if t2, err2 := time.Parse(time.RFC3339, rawCreatedAt.String); err2 == nil {
			p.CreatedAt = t2
		} else {
			p.CreatedAt = time.Now()
		}
	} else {
		p.CreatedAt = time.Now()
	}

	// Delete expired content before returning it.
	if p.ExpiresAt != nil && time.Now().UTC().After(*p.ExpiresAt) {
		DeletePage(id)
		return nil, errors.New("page has expired")
	}

	return &p, nil
}

// DeletePage removes a page by ID.
func DeletePage(id string) {
	DB.Exec("DELETE FROM pages WHERE id = ?", id)
}
