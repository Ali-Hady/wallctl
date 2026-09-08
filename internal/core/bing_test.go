package core

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestBingSource_Fetch(t *testing.T) {
	src := NewBingSource()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	dstDir := t.TempDir()

	img, err := src.Fetch(ctx, dstDir)
	if err != nil {
		t.Fatalf("failed to fetch image: %v", err)
	}

	if img == nil {
		t.Fatal("expected image, got nil")
	}

	if img.ID == "" {
		t.Fatal("ID is empty")
	}

	if len(img.ID) > 8 {
		t.Fatalf("ID %q is too long (%d chars), expected <= 8", img.ID, len(img.ID))
	}

	if img.LocalPath == "" {
		t.Fatal("Local Path is empty")
	}

	info, err := os.Stat(img.LocalPath)
	if err != nil {
		t.Fatalf("file not found on disk at %s: %v", img.LocalPath, err)
	}
	if info.Size() == 0 {
		t.Errorf("downloaded image file %s is 0 bytes", img.LocalPath)
	}
}

func TestBingSource_FetchBatch(t *testing.T) {
	src := NewBingSource()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	dstDir := t.TempDir()

	count := 3
	images, err := src.FetchBatch(ctx, dstDir, count)
	if err != nil {
		t.Fatalf("failed to fetch images: %v", err)
	}

	if len(images) != count {
		t.Fatalf("expected %d images, got %d", count, len(images))
	}

	seenIDs := make(map[string]bool)
	for _, img := range images {
		if img.ID == "" {
			t.Fatal("ID is empty")
		}
		if len(img.ID) > 8 {
			t.Fatalf("ID %q is too long (%d chars), expected <= 8", img.ID, len(img.ID))
		}
		if seenIDs[img.ID] {
			t.Errorf("duplicate image ID %q found in batch", img.ID)
		}
		seenIDs[img.ID] = true

		if img.LocalPath == "" {
			t.Fatal("Local Path is empty")
		}

		info, err := os.Stat(img.LocalPath)
		if err != nil {
			t.Fatalf("file not found on disk at %s: %v", img.LocalPath, err)
		}
		if info.Size() == 0 {
			t.Errorf("downloaded image file %s is 0 bytes", img.LocalPath)
		}
	}
}
