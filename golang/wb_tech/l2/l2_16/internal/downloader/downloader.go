package downloader

import (
	"io"
	"net/http"
	"time"
)

type Downloader struct {
	client http.Client
}

func NewDownloader(timeout time.Duration) *Downloader {
	return &Downloader{
		client: http.Client{
			Timeout: timeout,
		},
	}
}

func (d *Downloader) Download(rawURL string) ([]byte, string, error) {
	resp, err := d.client.Get(rawURL)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	return body, resp.Header.Get("Content-Type"), nil
}
