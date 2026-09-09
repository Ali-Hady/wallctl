package core

import (
	"errors"
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

func TestJSONStore_SaveNoActive(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "index.json")
	store, _ := NewJSONStore(storePath)

	// 1. First run fallback: when store is empty, SaveNoActive must anchor CurID
	img1 := newSampleImage("img1")
	if err := store.SaveNoActive(img1); err != nil {
		t.Fatalf("unexpected error on initial SaveNoActive: %v", err)
	}

	curr, err := store.Current()
	if err != nil {
		t.Fatalf("expected CurID to be anchored on empty store, got error: %v", err)
	}
	if curr.ID != "img1" {
		t.Errorf("expected initial CurID 'img1', got %q", curr.ID)
	}

	// 2. Subsequent SaveNoActive must append image but NOT alter CurID
	img2 := newSampleImage("img2")
	if err := store.SaveNoActive(img2); err != nil {
		t.Fatalf("unexpected error on second SaveNoActive: %v", err)
	}

	curr, err = store.Current()
	if err != nil {
		t.Fatalf("unexpected error getting current: %v", err)
	}
	if curr.ID != "img1" {
		t.Errorf("expected CurID to remain 'img1', but shifted to %q", curr.ID)
	}

	// Verify both images are stored in chronological order
	if len(store.Data.Images) != 2 || store.Data.Images[1].ID != "img2" {
		t.Fatalf("expected 2 images with img2 at index 1")
	}
}

func TestJSONStore_Save_UpdatePreservesMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "index.json")
	store, _ := NewJSONStore(storePath)

	img := newSampleImage("img1")
	_ = store.Save(img)

	// User customizes image
	if err := store.MarkFavorite("img1", "custom-alias"); err != nil {
		t.Fatalf("unexpected error setting favorite: %v", err)
	}

	// Re-fetch/save the same image with updated remote metadata
	updatedImg := newSampleImage("img1")
	updatedImg.Title = "Updated Title From API"
	if err := store.Save(updatedImg); err != nil {
		t.Fatalf("unexpected error updating image: %v", err)
	}

	// Ensure slice length did not increase (in-place update)
	if len(store.Data.Images) != 1 {
		t.Fatalf("expected exactly 1 image in store, got %d", len(store.Data.Images))
	}

	saved, err := store.Get("img1")
	if err != nil {
		t.Fatalf("failed to retrieve image: %v", err)
	}

	// Verify API updates were accepted
	if saved.Title != "Updated Title From API" {
		t.Errorf("expected title to update, got %q", saved.Title)
	}

	// Verify user-defined state was preserved
	if !saved.Favorite {
		t.Errorf("expected Favorite flag to remain true")
	}
	if saved.Alias != "custom-alias" {
		t.Errorf("expected Alias to remain 'custom-alias', got %q", saved.Alias)
	}
}

func TestJSONStore_SentinelErrors(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "index.json")
	store, _ := NewJSONStore(storePath)

	// 1. Empty store checks
	if _, err := store.Current(); !errors.Is(err, ErrEmptyStore) {
		t.Errorf("expected ErrEmptyStore on empty store Current(), got %v", err)
	}
	if _, err := store.Shift(1, false); !errors.Is(err, ErrEmptyStore) {
		t.Errorf("expected ErrEmptyStore on empty store Shift(), got %v", err)
	}

	// 2. Not found check
	_ = store.Save(newSampleImage("img1"))
	if _, err := store.Get("nonexistent"); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for missing identifier, got %v", err)
	}

	// 3. Out of bounds check
	if _, err := store.Shift(1, false); !errors.Is(err, ErrOutOfBounds) {
		t.Errorf("expected ErrOutOfBounds shifting past boundary, got %v", err)
	}
}
