package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func Do(
	ctx context.Context,
	httpClient *http.Client,
	method, url string,
	headers map[string]string,
	requestBody io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, requestBody)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, res.Body)
		_ = res.Body.Close()
	}()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s %s resulted in %s %s", method, url, res.Status, string(body))
	}
	return body, nil
}
