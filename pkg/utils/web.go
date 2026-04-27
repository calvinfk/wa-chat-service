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

	fileData := make([]byte, resp.ContentLength)
	_, err = resp.Body.Read(fileData)
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
	if strings.HasPrefix(lowerHeader, "bytes=") {
		rangeSpec = strings.TrimSpace(rangeHeader[len("bytes="):])
	}
	if idx := strings.Index(rangeSpec, ","); idx >= 0 {
		rangeSpec = rangeSpec[:idx]
	}
	rangeSpec = strings.TrimSpace(rangeSpec)
	parts := strings.SplitN(rangeSpec, "-", 2)
	if len(parts) != 2 {
		return 0, 0, false, errs.ErrGenericRangeNotSatisfiable
	}
	parts[0] = strings.TrimSpace(parts[0])
	parts[1] = strings.TrimSpace(parts[1])
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

type ProgressReader struct {
	Ctx     context.Context
	Reader  io.Reader
	Size    int64
	read    int64
	lastLog time.Time
	Log     func(string, ...any)
}

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
