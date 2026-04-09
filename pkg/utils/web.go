package utils

import (
	"log"
	"net/http"
	"net/url"
	"strings"
)

func GetURLHeaders(fileURL string) (http.Header, error) {
	// Implementation for checking MIME type based on file URL
	log.Println("[INFO][utils/utils.go][GetURLHeaders] Checking URL headers for:", fileURL)
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

func GetFileNameFromURL(fileURL string) string {
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
