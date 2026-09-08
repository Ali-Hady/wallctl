package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"time"
)

const BingEndpoint = "https://www.bing.com/HPImageArchive.aspx?format=js&idx=0&n=%d&mkt=en-US"

type BingResponse struct {
	Images []BingImage `json:"images"`
}

type BingImage struct {
	Startdate string `json:"startdate"`
	URL       string `json:"url"`
	Copyright string `json:"copyright"`
	Title     string `json:"title"`
	Hsh       string `json:"hsh"`
}

type BingSource struct {
	client *http.Client
}

func (src *BingSource) Name() string {
	return "bing"
}

func (src *BingSource) Fetch(ctx context.Context, dstDir string) (*Image, error) {
	images, err := src.FetchBatch(ctx, dstDir, 1)
	if err != nil {
		return nil, err
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("no images found")
	}

	return images[0], nil
}

func (src *BingSource) FetchBatch(ctx context.Context, dstDir string, count int) ([]*Image, error) {
	if count < 1 {
		count = 1
	}
	if count > 8 {
		count = 8
	}

	endpoint := fmt.Sprintf(BingEndpoint, count)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := src.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch image: %s", resp.Status)
	}

	var bingResp BingResponse
	if err := json.NewDecoder(resp.Body).Decode(&bingResp); err != nil {
		return nil, err
	}

	if len(bingResp.Images) == 0 {
		return nil, fmt.Errorf("no images found")
	}

	var results []*Image
	for _, img := range bingResp.Images {
		file := filepath.Join(dstDir, src.Name()+"_"+img.Startdate+".jpg")
		absoluteURL := "https://www.bing.com" + img.URL
		if err := downloadImage(src.client, ctx, absoluteURL, file); err != nil {
			return nil, err
		}

		hsh := img.Hsh
		if len(hsh) > 8 {
			hsh = hsh[:8]
		} else if hsh == "" {
			hsh = GenerateID(img.URL)
		}

		t, err := time.Parse("20060102", img.Startdate)
		if err != nil {
			t = time.Now().UTC()
		}

		results = append(results, &Image{
			URL:       absoluteURL,
			LocalPath: file,
			Source:    src.Name(),
			Title:     img.Title,
			Credit:    img.Copyright,
			FetchedAt: t,
			ID:        hsh,
			Favorite:  false,
		})
	}

	return results, nil
}

func NewBingSource() *BingSource {
	return &BingSource{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}
