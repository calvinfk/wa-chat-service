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

func (wb *Client) DownloadMedia(mediaURL string) ([]byte, http.Header, int, error) {
	req, err := http.NewRequest(http.MethodGet, mediaURL, nil)
	if err != nil {
		return nil, nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+wb.UserAccessToken)
	resp, err := wb.httpClient.Do(req)
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
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := wb.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp.Header, nil
}

func (wb *Client) startResumableUploadSession(payload StartUploadSessionRequest) (StartUploadSessionResponse, int, error) {
	endpoint := fmt.Sprintf("%s/%s/uploads", wb.GetBaseURLVersion(), wb.AppID)
	err := wb.validator.Validate(payload)
	if err != nil {
		return StartUploadSessionResponse{}, 0, err
	}
	body, httpCode, err := wb.accessAPIWithoutAuth(http.MethodPost, endpoint, payload)
	if err != nil {
		return StartUploadSessionResponse{}, httpCode, err
	}
	if httpCode != http.StatusOK {
		return StartUploadSessionResponse{}, httpCode, fmt.Errorf("failed to start upload session, status code: %d", httpCode)
	}
	var response StartUploadSessionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return StartUploadSessionResponse{}, httpCode, err
	}
	return response, httpCode, nil
}

func (wb *Client) uploadFile(payload UploadFileRequest) (UploadFileResponse, int, error) {
	endpoint := fmt.Sprintf("%s/%s", wb.GetBaseURLVersion(), payload.UploadSessionID)
	err := wb.validator.Validate(payload)
	if err != nil {
		return UploadFileResponse{}, 0, err
	}
	body, httpCode, err := wb.accessAPI(http.MethodPut, endpoint, payload)
	if err != nil {
		return UploadFileResponse{}, httpCode, err
	}
	if httpCode != http.StatusOK {
		return UploadFileResponse{}, httpCode, fmt.Errorf("failed to upload file, status code: %d", httpCode)
	}
	var response UploadFileResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return UploadFileResponse{}, httpCode, err
	}
	return response, httpCode, nil
}

func (wb *Client) ResumableUpload(fileHeader *multipart.FileHeader) (string, int, error) {
	if fileHeader == nil {
		return "", 0, fmt.Errorf("file is nil")
	}
	if fileHeader.Size == 0 {
		return "", 0, fmt.Errorf("file is empty")
	}
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		return "", 0, fmt.Errorf("content type is required")
	}
	if !resumableUploadMimeTypes[contentType] {
		return "", 0, fmt.Errorf("unsupported content type: %s", contentType)
	}
	startResponse, _, err := wb.startResumableUploadSession(StartUploadSessionRequest{
		FileName:    fileHeader.Filename,
		FileLength:  fileHeader.Size,
		FileType:    contentType,
		AccessToken: wb.UserAccessToken,
	})
	if err != nil {
		return "", 0, fmt.Errorf("start upload session: %w", err)
	}
	file, err := fileHeader.Open()
	if err != nil {
		return "", 0, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()
	// TODO: For large files, we should read and upload in chunks instead of reading the entire file into memory.
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", 0, fmt.Errorf("read file: %w", err)
	}
	uploadResponse, httpCode, err := wb.uploadFile(UploadFileRequest{
		UploadSessionID: startResponse.ID,
		FileOffset:      startResponse.FileOffset,
		FileBytes:       fileBytes,
	})
	if err != nil {
		return "", httpCode, fmt.Errorf("upload file: %w", err)
	}
	return uploadResponse.H, httpCode, nil
}
