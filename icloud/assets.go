package icloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// UploadUrlRequest is the request to ask for upload url.
type UploadUrlRequest struct {
	// AssetFields represents asset fields in records.
	UploadField []UploadField `json:"tokens,omitempty"`
}

// UploadField is an operation on a single record.
type UploadField struct {
	// Name of the record.
	RecordName string `json:"recordName,omitempty"`
	// Type of the record.
	RecordType string `json:"recordType,omitempty"`
	// Asset field name to upload.
	FieldName string `json:"fieldName,omitempty"`
}

// UploadUrlResponse is the response of upload request.
type UploadUrlResponse []UploadURL

// UploadURL is the returned URL for asset to upload
type UploadURL struct {
	// Name of the record.
	RecordName string `json:"recordName,omitempty"`
	// Asset field name in the record.
	FieldName string `json:"fieldName,omitempty"`
	// The location to upload the asset data.
	URL string `json:"url,omitempty"`
}

// Response of upload request
type UploadResponse struct {
	WrappingKey       string  `json:"wrappingKey,omitempty"`
	FileChecksum      string  `json:"fileChecksum,omitempty"`
	Receipt           string  `json:"receipt,omitempty"`
	ReferenceChecksum string  `json:"referenceChecksum,omitempty"`
	Size              float64 `json:"size,omitempty"`
}

// AssetsService handles communication with the asset related operations of
// the CloudKit Web Services API.
//
// CloudKit Web Services Reference: https://developer.apple.com/library/archive/documentation/DataManagement/Conceptual/CloudKitWebServicesReference/UploadAssets.html
type AssetsService service

// Get assets upload URLs
func (s *AssetsService) UploadURL(ctx context.Context, database Database, req UploadUrlRequest) (UploadUrlResponse, error) {
	path := "/" + database.String() + s.basePath

	var res struct {
		Tokens UploadUrlResponse `json:"tokens,omitempty"`
	}
	if err := s.client.call(ctx, http.MethodPost, path, req, &res); err != nil {
		return nil, fmt.Errorf("get asset upload url failed, cause: %w", err)
	}

	return res.Tokens, nil
}

// Upload to returned url
func (s *AssetsService) Upload(ctx context.Context, url string, data []byte) (*UploadResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "asset")
	if _, err := io.Copy(part, bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("failed construct upload request, cause: %w", err)
	}
	writer.Close()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed construct upload request, cause: %w", err)
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload request failed, cause: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upload asset not 2xx response status code: %d, body: %s", resp.StatusCode, string(body))
	}
	var v struct {
		SingleFile UploadResponse `json:"singleFile"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("upload asset decode errro, cause: %w", err)
	}
	return &v.SingleFile, nil
}
