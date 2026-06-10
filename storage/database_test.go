package storage

import (
	"testing"
	"time"
)

func initTestDB(t *testing.T) {
	t.Helper()
	InitDB("sqlite", ":memory:")
}

func TestSaveGetAndDeletePage(t *testing.T) {
	initTestDB(t)

	id, pwd, err := SavePage("# title", "<h1>title</h1>", false, 0, true)
	if err != nil {
		t.Fatalf("SavePage() error = %v", err)
	}
	if pwd == "" {
		t.Fatal("SavePage() password = empty, want generated password")
	}

	page, err := GetPage(id)
	if err != nil {
		t.Fatalf("GetPage() error = %v", err)
	}
	if page.ID != id {
		t.Fatalf("GetPage() id = %q, want %q", page.ID, id)
	}
	if page.Markdown != "# title" {
		t.Fatalf("GetPage() markdown = %q, want original content", page.Markdown)
	}
	if page.Password != pwd {
		t.Fatalf("GetPage() password = %q, want %q", page.Password, pwd)
	}

	DeletePage(id)
	if _, err := GetPage(id); err == nil {
		t.Fatal("GetPage() after DeletePage() error = nil, want missing page error")
	}
}

func TestGetPageDeletesExpiredContent(t *testing.T) {
	initTestDB(t)

	id, _, err := SavePage("expired", "<p>expired</p>", false, time.Millisecond, false)
	if err != nil {
		t.Fatalf("SavePage() error = %v", err)
	}

	time.Sleep(25 * time.Millisecond)

	if _, err := GetPage(id); err == nil {
		t.Fatal("GetPage() error = nil, want expired page error")
	}

	var exists bool
	if err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM pages WHERE id = ?)", id).Scan(&exists); err != nil {
		t.Fatalf("query after expiration error = %v", err)
	}
	if exists {
		t.Fatal("expired page still exists after GetPage()")
	}
}

func TestInitDBIsIdempotent(t *testing.T) {
	InitDB("sqlite", "file::memory:?cache=shared")
	InitDB("sqlite", "file::memory:?cache=shared")

	id, _, err := SavePage("ok", "<p>ok</p>", false, 0, false)
	if err != nil {
		t.Fatalf("SavePage() after repeated InitDB() error = %v", err)
	}
	if _, err := GetPage(id); err != nil {
		t.Fatalf("GetPage() after repeated InitDB() error = %v", err)
	}
}
