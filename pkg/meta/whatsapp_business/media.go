package whatsapp_business

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
)

func (wb *Client) UploadMedia(fileBytes []byte, filename string, mimeType string) (UploadMediaResponse, int, error) {
	var emptyResponse UploadMediaResponse
	endpoint := fmt.Sprintf("%s/%s/media", wb.GetBaseURLVersion(), wb.PhoneNumberID)
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	_ = w.WriteField("messaging_product", "whatsapp")

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename))
	h.Set("Content-Type", mimeType)

	part, err := w.CreatePart(h)
	if err != nil {
		return emptyResponse, 0, fmt.Errorf("create part: %w", err)
	}
	if _, err = part.Write(fileBytes); err != nil {
		return emptyResponse, 0, fmt.Errorf("write bytes: %w", err)
	}
	if err := w.Close(); err != nil {
		return emptyResponse, 0, fmt.Errorf("close writer: %w", err)
	}
	req, err := http.NewRequest("POST", endpoint, &buf)
	if err != nil {
		return emptyResponse, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+wb.UserAccessToken)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := wb.httpClient().Do(req)
	if err != nil {
		return emptyResponse, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return emptyResponse, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return parseMetaErrorResponse(emptyResponse, body, resp.StatusCode)
	}
	var response UploadMediaResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return emptyResponse, resp.StatusCode, err
	}
	return response, resp.StatusCode, nil
}

func (wb *Client) GetMediaURL(mediaID string) (GetMediaURLResponse, int, error) {
	endpoint := fmt.Sprintf("%s/%s", wb.GetBaseURLVersion(), mediaID)
	body, httpCode, err := wb.accessAPI(endpoint, "GET", nil)
	if err != nil {
		return GetMediaURLResponse{}, httpCode, err
	}
	if httpCode != http.StatusOK {
		return parseMetaErrorResponse(GetMediaURLResponse{}, body, httpCode)
	}
	var response GetMediaURLResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return GetMediaURLResponse{}, httpCode, err
	}
	return response, httpCode, nil
}

func (wb *Client) DownloadMedia(mediaURL string) ([]byte, http.Header, int, error) {
	req, err := http.NewRequest("GET", mediaURL, nil)
	if err != nil {
		return nil, nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+wb.UserAccessToken)
	resp, err := wb.httpClient().Do(req)
	if err != nil {
		return nil, nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, resp.Header, resp.StatusCode, fmt.Errorf("failed to download media, status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.Header, resp.StatusCode, err
	}
	return body, resp.Header, resp.StatusCode, nil
}

func (wb *Client) DeleteMedia(mediaID string) (DeleteMediaResponse, int, error) {
	endpoint := fmt.Sprintf("%s/%s", wb.GetBaseURLVersion(), mediaID)
	body, httpCode, err := wb.accessAPI(endpoint, "DELETE", nil)
	if err != nil {
		return DeleteMediaResponse{}, httpCode, err
	}
	if httpCode != http.StatusOK {
		return parseMetaErrorResponse(DeleteMediaResponse{}, body, httpCode)
	}
	var response DeleteMediaResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return DeleteMediaResponse{}, httpCode, err
	}
	return response, httpCode, nil
}

func (wb *Client) GetHeaders(url string) (http.Header, error) {
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := wb.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp.Header, nil
}

func ParseMediaExtension(mimeType string) string {
	extension, exists := mimeTypeExtensionMap[mimeType]
	if !exists {
		return ""
	}
	return extension
}
