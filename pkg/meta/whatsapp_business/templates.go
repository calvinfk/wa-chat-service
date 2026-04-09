package whatsapp_business

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

const (
	endpointTemplate = "message_templates"
)

func (wb *Client) GetTemplateList(query ...string) ([]any, int, error) {
	endpoint := fmt.Sprintf("%s/%s/%s", wb.GetBaseURLVersion(), wb.WabaID, endpointTemplate)
	var queries []string
	if len(query) > 0 {
		for _, q := range query {
			if q != "" {
				queries = append(queries, q)
			}
		}
		if len(queries) > 0 {
			endpoint += "?" + strings.Join(queries, "&")
		}
	}
	body, httpCode, err := wb.accessAPI(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, httpCode, err
	}
	if httpCode != http.StatusOK {
		return parseMetaErrorResponse([]any{}, body, httpCode)
	}
	if len(queries) > 0 {
		var response struct {
			Data []any `json:"data"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, httpCode, err
		} else {
			return response.Data, httpCode, nil
		}
	} else {
		var response struct {
			Data []TemplateResponse `json:"data"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, httpCode, err
		} else {
			data := make([]any, len(response.Data))
			for i, v := range response.Data {
				data[i] = v
			}
			return data, httpCode, nil
		}
	}
}

func (wb *Client) GetTemplateByID(templateID string, query ...string) (TemplateResponse, int, error) {
	endpoint := fmt.Sprintf("%s/%s/%s", wb.GetBaseURLVersion(), wb.WabaID, templateID)
	if len(query) > 0 {
		var queries []string
		for _, q := range query {
			if q != "" {
				queries = append(queries, q)
			}
		}
		if len(queries) > 0 {
			endpoint += "?" + strings.Join(queries, "&")
		}
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
	if err := wb.validator.Validate(payload); err != nil {
		return TemplateCreateResponse{}, 0, err
	}
	endpoint := fmt.Sprintf("%s/%s/%s", wb.GetBaseURLVersion(), wb.WabaID, endpointTemplate)
	// TODO: validate payload before sending request
	body, httpCode, err := wb.accessAPI(http.MethodPost, endpoint, payload)
	if err != nil {
		return TemplateCreateResponse{}, httpCode, err
	}
	if httpCode != http.StatusOK && httpCode != http.StatusCreated {
		log.Println("[ERROR][pkg/meta/whatsapp_business/templates.go][CreateTemplate] failed to create template, response body:", string(body))
		return parseMetaErrorResponse(TemplateCreateResponse{}, body, httpCode)
	}
	var response TemplateCreateResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return TemplateCreateResponse{}, httpCode, err
	}
	return response, httpCode, nil
}

func (wb *Client) DeleteTemplate(templateID string, templateName string) (TemplateDeleteResponse, int, error) {
	var requestData TemplateDeleteRequest
	requestData.ID = templateID
	requestData.Name = templateName
	if err := wb.validator.Validate(requestData); err != nil {
		return TemplateDeleteResponse{}, 0, err
	}
	var queries []string
	endpoint := fmt.Sprintf("%s/%s/%s", wb.GetBaseURLVersion(), wb.WabaID, endpointTemplate)
	if templateID != "" {
		queries = append(queries, "id="+templateID)
	}
	if templateName != "" {
		queries = append(queries, "name="+templateName)
	}
	if len(queries) > 0 {
		endpoint += "?" + strings.Join(queries, "&")
	}
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
