package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
)

func downloadImage(client *http.Client, ctx context.Context, absoluteURL, filepath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, absoluteURL, nil)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("executing download request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("creating file %s: %w", filepath, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		_ = os.Remove(filepath) // Clean up half-written garbage
		return fmt.Errorf("writing image data to %s: %w", filepath, err)
	}

	return nil
}

// GenerateID produces a short, predictable 8-character hex hash.
func generateID(seed string) string {
	hash := sha256.Sum256([]byte(seed))
	return hex.EncodeToString(hash[:])[:8]
}
