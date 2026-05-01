package model

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type PaginationResponse struct {
	Items    []map[string]any `json:"items"`
	Page     int64            `json:"page"`
	PageSize int64            `json:"page_size"`
	Total    int64            `json:"total"`
}

type ConnectionResponse struct {
	Status      string `json:"status"`
	Version     string `json:"version"`
	Database    string `json:"database"`
	Server      string `json:"server"`
	RoundTripMS int64  `json:"round_trip_ms"`
}

type ConnectionTestRequest struct {
	Host       string `json:"host"`
	Port       string `json:"port"`
	Database   string `json:"database"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	AuthSource string `json:"auth_source"`
}

type DatabaseListResponse struct {
	Items []string `json:"items"`
}

type CollectionItem struct {
	Name          string `json:"name"`
	DocumentCount *int64 `json:"document_count,omitempty"`
	IsSystem      bool   `json:"is_system"`
}

type CollectionListResponse struct {
	Items []CollectionItem `json:"items"`
}

type CreateCollectionRequest struct {
	Name string `json:"name" binding:"required"`
}

type CreateDatabaseRequest struct {
	Database        string `json:"database" binding:"required"`
	FirstCollection string `json:"first_collection" binding:"required"`
}

type WriteDocumentRequest struct {
	Document map[string]any `json:"document" binding:"required"`
}

type QueryCondition struct {
	Field     string `json:"field"`
	Operator  string `json:"operator"`
	ValueType string `json:"value_type"`
	Value     any    `json:"value"`
	Join      string `json:"join,omitempty"`
}

type CollectionBackup struct {
	Database   string           `json:"database"`
	Collection string           `json:"collection"`
	ExportedAt string           `json:"exported_at"`
	Documents  []map[string]any `json:"documents"`
}

type BulkDeleteRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

type IndexItem struct {
	Name   string         `json:"name"`
	Keys   map[string]any `json:"keys"`
	Unique bool           `json:"unique"`
	Sparse bool           `json:"sparse"`
}

type IndexListResponse struct {
	Items []IndexItem `json:"items"`
}

type CreateIndexRequest struct {
	Name   string `json:"name" binding:"required"`
	Field  string `json:"field" binding:"required"`
	Order  int    `json:"order"`
	Unique bool   `json:"unique"`
	Sparse bool   `json:"sparse"`
}
