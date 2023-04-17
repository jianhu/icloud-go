package icloud

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
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
	RecorName string `json:"recordName,omitempty"`
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
	RecorName string `json:"recordName,omitempty"`
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

	var res UploadUrlResponse
	if err := s.client.call(ctx, http.MethodPost, path, req, &res); err != nil {
		return nil, err
	}

	return res, nil
}

// Upload to returned url
func (s *AssetsService) Upload(ctx context.Context, url string, data []byte) (*UploadResponse, error) {
	body := bytes.NewReader(data)
	resp, err := http.Post(url, "application/x-www-form-urlencoded", body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, errors.New("not 2xx status code")
	}
	defer resp.Body.Close()
	jsonData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var v struct {
		SingleFile UploadResponse `json:"singleFile"`
	}
	if err := json.Unmarshal(jsonData, &v); err != nil {
		return nil, err
	}
	return &v.SingleFile, nil
}
