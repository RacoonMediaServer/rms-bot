package bot

import (
	"io"
	"net/http"
	"time"
)

const downloadTimeout = 10 * time.Second

func downloadPhoto(url string) (content []byte, err error) {
	var req *http.Request
	var resp *http.Response

	cli := http.Client{Timeout: downloadTimeout}
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	resp, err = cli.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	content, err = io.ReadAll(resp.Body)
	return
}
