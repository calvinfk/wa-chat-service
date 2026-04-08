package whatsapp_business

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	endpointTemplate = "message_templates"
)

func (wb *Client) GetTemplateList(query ...string) ([]TemplateResponse, int, error) {
	endpoint := fmt.Sprintf("%s/%s/%s", wb.GetBaseURLVersion(), wb.WabaID, endpointTemplate)
	if len(query) > 0 {
		endpoint += "?" + strings.Join(query, "&")
	}
	body, httpCode, err := wb.accessAPI(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, httpCode, err
	}
	if httpCode != http.StatusOK {
		return parseMetaErrorResponse([]TemplateResponse{}, body, httpCode)
	}
	var response struct {
		Data []TemplateResponse `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, httpCode, err
	}
	return response.Data, httpCode, nil
}

func (wb *Client) GetTemplateByID(templateID string, query ...string) (TemplateResponse, int, error) {
	endpoint := fmt.Sprintf("%s/%s/%s", wb.GetBaseURLVersion(), wb.WabaID, templateID)
	if len(query) > 0 {
		endpoint += "?" + strings.Join(query, "&")
	}
	body, httpCode, err := wb.accessAPI(http.MethodGet, endpoint, nil)
	if err != nil {
		return TemplateResponse{}, httpCode, err
	}
	if httpCode != http.StatusOK {
		return parseMetaErrorResponse(TemplateResponse{}, body, httpCode)
	}
	var response TemplateResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return TemplateResponse{}, httpCode, err
	}
	return response, httpCode, nil
}

func (wb *Client) CreateTemplate(payload TemplateCreateRequest) (TemplateCreateResponse, int, error) {
	endpoint := fmt.Sprintf("%s/%s/%s", wb.GetBaseURLVersion(), wb.WabaID, endpointTemplate)
	// TODO: validate payload before sending request
	body, httpCode, err := wb.accessAPI(http.MethodPost, endpoint, payload)
	if err != nil {
		return TemplateCreateResponse{}, httpCode, err
	}
	if httpCode != http.StatusOK && httpCode != http.StatusCreated {
		return parseMetaErrorResponse(TemplateCreateResponse{}, body, httpCode)
	}
	var response TemplateCreateResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return TemplateCreateResponse{}, httpCode, err
	}
	return response, httpCode, nil
}

func (wb *Client) DeleteTemplate(templateID string, templateName string) (TemplateDeleteResponse, int, error) {
	if templateID == "" || templateName == "" {
		return TemplateDeleteResponse{}, 0, fmt.Errorf("templateID and templateName are required")
	}
	endpoint := fmt.Sprintf("%s/%s/%s?hsm_id=%s&name=%s", wb.GetBaseURLVersion(), wb.WabaID, endpointTemplate, templateID, templateName)
	body, httpCode, err := wb.accessAPI(http.MethodDelete, endpoint, nil)
	if err != nil {
		return TemplateDeleteResponse{}, httpCode, err
	}
	if httpCode != http.StatusOK && httpCode != http.StatusNoContent {
		return parseMetaErrorResponse(TemplateDeleteResponse{}, body, httpCode)
	}
	var response TemplateDeleteResponse
	if err := json.Unmarshal(body, &response); err != nil {
		// If the response body is empty (which can happen with 204 No Content), we can assume the delete was successful
		return TemplateDeleteResponse{Success: true}, httpCode, nil
	}
	return response, httpCode, nil
}

func (wb *Client) UpdateTemplate(templateID string, payload TemplateCreateRequest) (TemplateCreateResponse, int, error) {
	endpoint := fmt.Sprintf("%s/%s/%s", wb.GetBaseURLVersion(), wb.WabaID, templateID)
	if templateID == "" {
		return TemplateCreateResponse{}, 0, fmt.Errorf("templateID is required")
	}
	body, httpCode, err := wb.accessAPI(http.MethodPut, endpoint, payload)
	if err != nil {
		return TemplateCreateResponse{}, httpCode, err
	}
	if httpCode != http.StatusOK {
		return parseMetaErrorResponse(TemplateCreateResponse{}, body, httpCode)
	}
	var response TemplateCreateResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return TemplateCreateResponse{}, httpCode, err
	}
	return response, httpCode, nil
}
