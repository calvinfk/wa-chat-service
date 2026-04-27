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

func ParseMediaExtension(mimeType string) string {
	extension, exists := mimeTypeExtensionMap[mimeType]
	if !exists {
		return ""
	}
	return extension
}

func (wb *Client) UploadMedia(fileBytes []byte, filename string, mimeType string) (UploadMediaResponse, int, error) {
	var emptyResponse UploadMediaResponse
	endpoint := fmt.Sprintf("%s/%s/media", wb.GetBaseURLVersion(), wb.PhoneNumberId)
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
	resp, err := wb.httpClient.Do(req)
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
	body, httpCode, err := wb.accessAPI(http.MethodGet, endpoint, nil)
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

// Downloads media content from the given media URL. Caller is responsible for closing the response body.
func (wb *Client) DownloadMedia(mediaURL string, rangeHeader string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, mediaURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+wb.UserAccessToken)
	if rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}
	resp, err := wb.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		resp.Body.Close() // Ensure we close the body if we're not returning it
		return nil, fmt.Errorf("failed to download media, status code: %d", resp.StatusCode)
	}
	return resp, nil
}

func (wb *Client) DeleteMedia(mediaID string) (DeleteMediaResponse, int, error) {
	endpoint := fmt.Sprintf("%s/%s", wb.GetBaseURLVersion(), mediaID)
	body, httpCode, err := wb.accessAPI(http.MethodDelete, endpoint, nil)
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
	// Some upstreams return a response body to HEAD and trigger protocol warnings in Go's client.
	// Use a minimal GET with Range to fetch headers without downloading the full file.
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Range", "bytes=0-0")
	req.Header.Set("Authorization", "Bearer "+wb.UserAccessToken)
	resp, err := wb.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp.Header, nil
}

func (wb *Client) StartResumableUploadSession(fileName string, fileLength int64, fileType string) (StartUploadSessionResponse, int, error) {
	if fileName == "" || fileLength <= 0 || fileType == "" {
		return StartUploadSessionResponse{}, 0, fmt.Errorf("fileName, fileLength, and fileType are required and must be valid")
	}
	if !resumableUploadMimeTypes[fileType] {
		return StartUploadSessionResponse{}, 0, fmt.Errorf("unsupported file type: %s", fileType)
	}
	payload := StartUploadSessionRequest{
		FileName:    fileName,
		FileLength:  fileLength,
		FileType:    fileType,
		AccessToken: wb.UserAccessToken,
	}
	if err := wb.validator.Struct(payload); err != nil {
		return StartUploadSessionResponse{}, 0, fmt.Errorf("validation error: %w", err)
	}
	endpoint := fmt.Sprintf("%s/%s/uploads?file_name=%s&file_length=%d&file_type=%s&access_token=%s", wb.GetBaseURLVersion(), wb.AppId, payload.FileName, payload.FileLength, payload.FileType, payload.AccessToken)
	body, httpCode, err := wb.accessAPIWithoutAuth(http.MethodPost, endpoint, nil)
	if err != nil {
		return StartUploadSessionResponse{}, httpCode, err
	}
	if httpCode != http.StatusOK {
		return parseMetaErrorResponse(StartUploadSessionResponse{}, body, httpCode)
	}
	var response StartUploadSessionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return StartUploadSessionResponse{}, httpCode, err
	}
	return response, httpCode, nil
}

func (wb *Client) StartResumableUpload(uploadSessionID string, fileOffset int64, file []byte) (UploadFileResponse, int, error) {
	endpoint := fmt.Sprintf("%s/%s", wb.GetBaseURLVersion(), uploadSessionID)
	payload := UploadFileRequest{
		UploadSessionID: uploadSessionID,
		FileOffset:      fileOffset,
		FileBytes:       file,
	}
	err := wb.validator.Struct(payload)
	if err != nil {
		return UploadFileResponse{}, 0, err
	}
	body, httpCode, err := wb.accessAPI(http.MethodPost, endpoint, payload)
	if err != nil {
		return UploadFileResponse{}, httpCode, err
	}
	if httpCode != http.StatusOK {
		return parseMetaErrorResponse(UploadFileResponse{}, body, httpCode)
	}
	var response UploadFileResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return UploadFileResponse{}, httpCode, err
	}
	return response, httpCode, nil
}
