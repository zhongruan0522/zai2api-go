package ocr

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"
)

const UpstreamURL = "https://ocr.z.ai/api/v1/z-ocr/tasks/process"

func SendRequest(file io.Reader, filename string, token string, maxRespBytes int64) ([]byte, error) {
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)
	contentType := writer.FormDataContentType()

	writeErrCh := make(chan error, 1)
	go func() {
		safeName := filepath.Base(filename)
		part, err := writer.CreateFormFile("file", safeName)
		if err != nil {
			_ = pw.CloseWithError(err)
			writeErrCh <- fmt.Errorf("create form file failed: %w", err)
			return
		}
		if _, err = io.Copy(part, file); err != nil {
			_ = pw.CloseWithError(err)
			writeErrCh <- fmt.Errorf("copy file failed: %w", err)
			return
		}
		if err = writer.Close(); err != nil {
			_ = pw.CloseWithError(err)
			writeErrCh <- fmt.Errorf("close writer failed: %w", err)
			return
		}
		_ = pw.Close()
		writeErrCh <- nil
	}()

	req, err := http.NewRequest("POST", UpstreamURL, pr)
	if err != nil {
		_ = pr.CloseWithError(err)
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Origin", "https://ocr.z.ai")
	req.Header.Set("Referer", "https://ocr.z.ai/")
	req.Header.Set("User-Agent", "Mozilla/5.0 AppleWebKit/537.36 Chrome/143 Safari/537.36")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		_ = pr.CloseWithError(err)
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	var reader io.Reader = resp.Body
	if maxRespBytes > 0 {
		reader = io.LimitReader(resp.Body, maxRespBytes+1)
	}
	respBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}
	if maxRespBytes > 0 && int64(len(respBody)) > maxRespBytes {
		return nil, fmt.Errorf("upstream response too large")
	}

	if writeErr := <-writeErrCh; writeErr != nil {
		return nil, writeErr
	}

	return respBody, nil
}
