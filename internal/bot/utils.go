package bot

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

const downloadTimeout = 10 * time.Second

func download(url string) (content []byte, err error) {
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
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}

	content, err = io.ReadAll(resp.Body)
	return
}
