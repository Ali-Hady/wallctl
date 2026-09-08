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
	FetchImage(cxt context.Context, dstDir string) (*Image, error)
}

type Store interface {
	SaveImage(image *Image) error
	GetImage(id string) (*Image, error)
	GetIDfromAlias(alias string) (*Image, error)
	ListImages(limit int, offset int, favs bool) ([]*Image, error)
	DeleteImage(id string) error
	DeleteImages(ids []string) error
	MarkFavorite(id string, alias string) error
	Exists(id string) (bool, error)
	GetCurrentImage() (*Image, error)
	Shift(step int, favs bool) (*Image, error)
	ExportImage(id string, dstPath string) error
}
