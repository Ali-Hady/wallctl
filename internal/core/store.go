package core

import (
	"fmt"
)

type Data struct {
	CurID  string  `json:"current_id"`
	Images []Image `json:"images"`
}

type JSONStore struct {
	FilePath string
	Data     *Data
}

func (s *JSONStore) Save(img *Image) error {
	for i := range s.Data.Images {
		if s.Data.Images[i].ID == img.ID {
			var (
				alias    string
				favorite bool
			)
			alias = s.Data.Images[i].Alias
			favorite = s.Data.Images[i].Favorite
			s.Data.Images[i] = *img
			s.Data.Images[i].Alias = alias
			s.Data.Images[i].Favorite = favorite
			s.Data.CurID = img.ID
			return nil
		}
	}

	s.Data.Images = append(s.Data.Images, *img)
	s.Data.CurID = img.ID
	return nil
}

func (s *JSONStore) Get(identifier string) (*Image, error) {
	if identifier == "" {
		return nil, fmt.Errorf("identifier cannot be empty")
	}
	for i := range s.Data.Images {
		if s.Data.Images[i].ID == identifier || s.Data.Images[i].Alias == identifier {
			return &s.Data.Images[i], nil
		}
	}

	return nil, fmt.Errorf("image with ID or alias %q not found", identifier)
}

func (s *JSONStore) List(limit int, favs bool) ([]Image, error) {
	var images []Image
	for i := len(s.Data.Images) - 1; i >= 0 && (limit == 0 || len(images) < limit); i-- {
		if !favs || s.Data.Images[i].Favorite {
			images = append(images, s.Data.Images[i])
		}
	}
	return images, nil
}

func (s *JSONStore) MarkFavorite(id string, alias string) error {
	if alias != "" {
		if existing, err := s.Get(alias); err == nil && existing.ID != id {
			return fmt.Errorf("alias %q is already in use by image %s", alias, existing.ID)
		}
	}

	img, err := s.Get(id)
	if err != nil {
		return err
	}

	img.Favorite = true
	if alias != "" {
		img.Alias = alias
	}
	return nil
}

func (s *JSONStore) Exists(id string) bool {
	for i := range s.Data.Images {
		if s.Data.Images[i].ID == id {
			return true
		}
	}
	return false
}

func (s *JSONStore) Current() (*Image, error) {
	if s.Data.CurID == "" {
		return nil, fmt.Errorf("no current image set")
	}
	return s.Get(s.Data.CurID)
}

func (s *JSONStore) Shift(step int, favs bool) (*Image, error) {
	if step == 0 {
		return s.Current()
	}
	if len(s.Data.Images) == 0 {
		return nil, fmt.Errorf("no images in store")
	}

	currentIndex := len(s.Data.Images) - 1
	if s.Data.CurID != "" {
		for i := range s.Data.Images {
			if s.Data.Images[i].ID == s.Data.CurID {
				currentIndex = i
				break
			}
		}
	}

	for newIndex := currentIndex + step; ; {
		if newIndex < 0 || newIndex >= len(s.Data.Images) {
			return nil, fmt.Errorf("shift out of bounds")
		}
		if !favs || s.Data.Images[newIndex].Favorite {
			s.Data.CurID = s.Data.Images[newIndex].ID
			return &s.Data.Images[newIndex], nil
		}
		newIndex += step
	}
}
