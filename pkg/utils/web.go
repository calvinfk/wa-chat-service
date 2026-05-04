package utils

import (
	"context"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"wa_chat_service/pkg/errs"
)

func GetURLHeaders(fileURL string) (http.Header, error) {
	client := http.Client{}
	resp, err := client.Head(fileURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp.Header, nil
}

func DownloadFile(fileURL string) ([]byte, http.Header, error) {
	client := http.Client{}
	resp, err := client.Get(fileURL)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	return fileData, resp.Header, nil
}

func GetFileNameFromURL(urlHeaders http.Header, fileURL string) string {
	contentDisposition := urlHeaders.Get("Content-Disposition")
	if contentDisposition != "" {
		if _, params, err := mime.ParseMediaType(contentDisposition); err == nil {
			return params["filename"]
		}
	}
	u, err := url.Parse(fileURL)
	if err != nil {
		return ""
	}
	urlParts := strings.Split(u.Path, "/")
	if len(urlParts) > 0 {
		tempName := urlParts[len(urlParts)-1]
		tempName = strings.Split(tempName, "?")[0] // remove query params
		// check if the last part of the URL path has a valid filename format (e.g., has an extension)
		if strings.Contains(tempName, ".") {
			return tempName
		}
	}
	return ""
}

// ParseRangeHeader parses the Range header from an HTTP request and returns the start and end byte positions for the requested range.
// It also returns a boolean indicating whether a valid range was specified and an error if the range is not satisfiable.
// The function handles both standard byte ranges (e.g., "bytes=0-499") and suffix byte ranges (e.g., "bytes=-500").
// If the Range header is empty, it returns default values indicating that the entire content should be returned.
func ParseRangeHeader(rangeHeader string, totalSize int64) (int64, int64, bool, error) {
	rangeHeader = strings.TrimSpace(rangeHeader)
	if rangeHeader == "" {
		return 0, 0, false, nil
	}
	if totalSize <= 0 {
		return 0, 0, false, errs.ErrGenericRangeNotSatisfiable
	}
	lowerHeader := strings.ToLower(rangeHeader)
	rangeSpec := rangeHeader
	// Trim the "bytes=" prefix if it exists
	if strings.HasPrefix(lowerHeader, "bytes=") {
		rangeSpec = strings.TrimSpace(rangeHeader[len("bytes="):])
	}
	// Handle multiple ranges (e.g., "bytes=0-499,500-999") by taking only the first range specified
	if idx := strings.Index(rangeSpec, ","); idx >= 0 {
		rangeSpec = rangeSpec[:idx]
	}
	// Trim whitespace and split the range specification into start and end parts using the "-" delimiter
	rangeSpec = strings.TrimSpace(rangeSpec)
	parts := strings.SplitN(rangeSpec, "-", 2)
	if len(parts) != 2 {
		return 0, 0, false, errs.ErrGenericRangeNotSatisfiable
	}
	parts[0] = strings.TrimSpace(parts[0])
	parts[1] = strings.TrimSpace(parts[1])
	// Handle suffix byte range (e.g., "bytes=-500") where the start part is empty.
	// In this case, we calculate the start position based on the total size and the specified suffix length.
	if parts[0] == "" {
		suffixLength, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil || suffixLength <= 0 {
			return 0, 0, false, errs.ErrGenericRangeNotSatisfiable
		}
		if suffixLength > totalSize {
			suffixLength = totalSize
		}
		return totalSize - suffixLength, totalSize - 1, true, nil
	}
	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || start < 0 || start >= totalSize {
		return 0, 0, false, errs.ErrGenericRangeNotSatisfiable
	}
	var end int64
	// Handle the case where the end part is empty (e.g., "bytes=500-") which means the range extends to the end of the content.
	if parts[1] == "" {
		end = totalSize - 1
	} else {
		end, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil || end < start {
			return 0, 0, false, errs.ErrGenericRangeNotSatisfiable
		}
		if end >= totalSize {
			end = totalSize - 1
		}
	}
	return start, end, true, nil
}

// ProgressReader is a custom io.Reader that wraps another io.Reader and provides progress logging functionality.
// It tracks the number of bytes read and logs the progress at regular intervals (e.g., every second) using the provided Log function.
// The reader also checks for context cancellation to allow for graceful termination of long-running read operations.
type ProgressReader struct {
	Ctx     context.Context
	Reader  io.Reader
	Size    int64
	read    int64
	lastLog time.Time
	Log     func(string, ...any)
}

// Read reads data from the underlying io.Reader and updates the progress tracking.
// It checks for context cancellation before reading and logs the progress at regular intervals.
// If the context is canceled, it returns an error to allow for graceful termination of the read operation.
// Chunk size is determined by the length of the provided byte slice, and the number of bytes read is accumulated to track overall progress.
func (p *ProgressReader) Read(b []byte) (int, error) {
	if p.Ctx.Err() != nil {
		return 0, p.Ctx.Err()
	}
	n, err := p.Reader.Read(b)
	if p.Log != nil && n > 0 {
		p.read += int64(n)
		if p.lastLog.IsZero() || time.Since(p.lastLog) >= time.Second {
			if p.Size > 0 {
				p.Log("[getMedia] stream progress: %d/%d bytes (%.1f%%)", p.read, p.Size, float64(p.read)*100/float64(p.Size))
			} else {
				p.Log("[getMedia] stream progress: %d bytes", p.read)
			}
			p.lastLog = time.Now()
		}
	}
	return n, err
}

func IsClientClosedStreamError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, io.ErrClosedPipe) {
		return true
	}
	errText := strings.ToLower(err.Error())
	return strings.Contains(errText, "response body closed") ||
		strings.Contains(errText, "stream closed") ||
		strings.Contains(errText, "broken pipe") ||
		strings.Contains(errText, "connection reset by peer") ||
		strings.Contains(errText, "connection closed")
}
