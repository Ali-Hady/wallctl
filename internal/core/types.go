package core

import (
	"context"
	"time"
)

type Image struct {
	ID        string    `json:"id"`
	Alias     string    `json:"alias,omitempty"`
	URL       string    `json:"url"`
	LocalPath string    `json:"local_path"`
	Source    string    `json:"source"`
	Title     string    `json:"title"`
	Credit    string    `json:"credit,omitempty"`
	FetchedAt time.Time `json:"fetched_at"`
	Favorite  bool      `json:"favorite"`
}

type Source interface {
	Name() string
	Fetch(ctx context.Context, dstDir string) (*Image, error)
}

type Store interface {
	Save(img *Image) error
	Get(identifier string) (*Image, error)
	List(limit int, favs bool) ([]Image, error)
	MarkFavorite(id string, alias string) error
	Exists(id string) bool
	Current() (*Image, error)
	Shift(step int, favs bool) (*Image, error)
}
