package core

import (
	"path/filepath"
	"testing"
	"time"
)

func newSampleImage(id string) *Image {
	return &Image{
		ID:        id,
		URL:       "https://example.com/" + id + ".jpg",
		LocalPath: "/tmp/" + id + ".jpg",
		Source:    "bing",
		Title:     "Sample " + id,
		Credit:    "Author",
		FetchedAt: time.Now().UTC(),
	}
}

func TestJSONStore_SaveAndReload(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "index.json")

	store1, err := NewJSONStore(storePath)
	if err != nil {
		t.Fatalf("failed to init store: %v", err)
	}

	img := newSampleImage("img1")
	if err := store1.Save(img); err != nil {
		t.Fatalf("failed to save image: %v", err)
	}

	// Simulate fresh process start reading the same file
	store2, err := NewJSONStore(storePath)
	if err != nil {
		t.Fatalf("failed to reload store: %v", err)
	}

	loaded, err := store2.Get("img1")
	if err != nil {
		t.Fatalf("failed to get image from reloaded store: %v", err)
	}

	if loaded.Title != img.Title {
		t.Errorf("expected title %q, got %q", img.Title, loaded.Title)
	}
}

func TestJSONStore_MarkFavorite_AliasConflict(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "index.json")
	store, _ := NewJSONStore(storePath)

	_ = store.Save(newSampleImage("img1"))
	_ = store.Save(newSampleImage("img2"))

	if err := store.MarkFavorite("img1", "nature"); err != nil {
		t.Fatalf("failed to favorite img1: %v", err)
	}

	// Attempting to assign the same alias to img2 must fail
	if err := store.MarkFavorite("img2", "nature"); err == nil {
		t.Fatal("expected error on duplicate alias, got nil")
	}
}

func TestJSONStore_Shift(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "index.json")
	store, _ := NewJSONStore(storePath)

	_ = store.Save(newSampleImage("img1")) // index 0 (oldest)
	_ = store.Save(newSampleImage("img2")) // index 1
	_ = store.Save(newSampleImage("img3")) // index 2 (newest, CurID)

	// Currently at img3 (index 2). Shifting forward (+1) should fail (already newest)
	if _, err := store.Shift(1, false); err == nil {
		t.Fatal("expected out of bounds error shifting forward from newest, got nil")
	}

	// Shift backward by 1 -> should land on img2
	curr, err := store.Shift(-1, false)
	if err != nil {
		t.Fatalf("unexpected error shifting back: %v", err)
	}
	if curr.ID != "img2" {
		t.Errorf("expected img2, got %s", curr.ID)
	}

	// Shift backward again -> should land on img1
	curr, err = store.Shift(-1, false)
	if err != nil {
		t.Fatalf("unexpected error shifting back: %v", err)
	}
	if curr.ID != "img1" {
		t.Errorf("expected img1, got %s", curr.ID)
	}

	// At oldest (img1, index 0). Shifting backward should fail
	if _, err := store.Shift(-1, false); err == nil {
		t.Fatal("expected out of bounds error shifting back from oldest, got nil")
	}

	// Shift step 0 should return current without moving
	curr, err = store.Shift(0, false)
	if err != nil || curr.ID != "img1" {
		t.Errorf("expected img1 on step 0, got %v (err: %v)", curr, err)
	}
}

func TestJSONStore_Shift_Favorites(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "index.json")
	store, _ := NewJSONStore(storePath)

	_ = store.Save(newSampleImage("img1")) // fav
	_ = store.Save(newSampleImage("img2")) // not fav
	_ = store.Save(newSampleImage("img3")) // not fav
	_ = store.Save(newSampleImage("img4")) // fav (current)

	_ = store.MarkFavorite("img1", "")
	_ = store.MarkFavorite("img4", "")

	// Currently at img4 (fav). Shift back with favs=true should skip img3 and img2 directly to img1
	curr, err := store.Shift(-1, true)
	if err != nil {
		t.Fatalf("unexpected error shifting to previous fav: %v", err)
	}
	if curr.ID != "img1" {
		t.Errorf("expected shift to jump straight to img1, got %s", curr.ID)
	}

	// Shifting back again should fail because no older favorites exist
	if _, err := store.Shift(-1, true); err == nil {
		t.Fatal("expected error shifting past oldest favorite, got nil")
	}
}

func TestJSONStore_List(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "index.json")
	store, _ := NewJSONStore(storePath)

	_ = store.Save(newSampleImage("img1"))
	_ = store.Save(newSampleImage("img2"))
	_ = store.Save(newSampleImage("img3"))

	// List with limit 2 should return [img3, img2]
	list, err := store.List(2, false)
	if err != nil {
		t.Fatalf("unexpected error listing: %v", err)
	}

	if len(list) != 2 {
		t.Fatalf("expected 2 items, got %d", len(list))
	}
	if list[0].ID != "img3" || list[1].ID != "img2" {
		t.Errorf("expected newest first [img3, img2], got [%s, %s]", list[0].ID, list[1].ID)
	}
}

func TestJSONStore_List_Favorites(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "index.json")
	store, _ := NewJSONStore(storePath)

	_ = store.Save(newSampleImage("img1"))
	_ = store.Save(newSampleImage("img2"))
	_ = store.Save(newSampleImage("img3"))
	_ = store.Save(newSampleImage("img4"))

	// Mark img1 and img3 as favorites
	if err := store.MarkFavorite("img1", "nature"); err != nil {
		t.Fatalf("failed to favorite img1: %v", err)
	}
	if err := store.MarkFavorite("img3", ""); err != nil {
		t.Fatalf("failed to favorite img3: %v", err)
	}

	// 1. Fetch all favorites (limit = 0)
	favs, err := store.List(0, true)
	if err != nil {
		t.Fatalf("unexpected error listing favorites: %v", err)
	}

	if len(favs) != 2 {
		t.Fatalf("expected 2 favorites, got %d", len(favs))
	}

	// Should be reverse-chronological: img3 (newer) before img1 (older)
	if favs[0].ID != "img3" || favs[1].ID != "img1" {
		t.Errorf("expected [img3, img1], got [%s, %s]", favs[0].ID, favs[1].ID)
	}

	// Verify metadata survived
	if favs[1].Alias != "nature" {
		t.Errorf("expected img1 alias to be %q, got %q", "nature", favs[1].Alias)
	}

	// 2. Fetch with limit = 1
	limitedFavs, err := store.List(1, true)
	if err != nil {
		t.Fatalf("unexpected error listing limited favorites: %v", err)
	}

	if len(limitedFavs) != 1 {
		t.Fatalf("expected 1 favorite with limit=1, got %d", len(limitedFavs))
	}
	if limitedFavs[0].ID != "img3" {
		t.Errorf("expected most recent favorite img3, got %s", limitedFavs[0].ID)
	}
}
