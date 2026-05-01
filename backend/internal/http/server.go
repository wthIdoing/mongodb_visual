package http

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"mongoDB_visual/backend/internal/config"
	"mongoDB_visual/backend/internal/model"
	"mongoDB_visual/backend/internal/mongodb"
	"mongoDB_visual/backend/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const (
	headerMongoHost       = "X-Mongo-Host"
	headerMongoPort       = "X-Mongo-Port"
	headerMongoDatabase   = "X-Mongo-Database"
	headerMongoUsername   = "X-Mongo-Username"
	headerMongoPassword   = "X-Mongo-Password"
	headerMongoAuthSource = "X-Mongo-AuthSource"
)

type Server struct {
	cfg  *config.Config
	pool *mongodb.Pool
}

func NewServer(cfg *config.Config, pool *mongodb.Pool) *gin.Engine {
	server := &Server{
		cfg:  cfg,
		pool: pool,
	}

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.FrontendOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", headerMongoHost, headerMongoPort, headerMongoDatabase, headerMongoUsername, headerMongoPassword, headerMongoAuthSource},
		AllowCredentials: true,
	}))

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/api/openapi.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(openAPISpec))
	})

	api := router.Group("/api/v1")
	{
		api.GET("/health", server.health)
		api.POST("/connection/test", server.testConnection)
		api.GET("/meta/connection", server.connection)
		api.GET("/databases", server.listDatabases)
		api.POST("/databases", server.createDatabase)
		api.DELETE("/databases/:db", server.deleteDatabase)
		api.GET("/databases/:db/collections", server.listCollections)
		api.POST("/databases/:db/collections", server.createCollection)
		api.DELETE("/databases/:db/collections/:collection", server.deleteCollection)
		api.GET("/databases/:db/collections/:collection/documents", server.listDocuments)
		api.GET("/databases/:db/collections/:collection/documents/:id", server.getDocument)
		api.POST("/databases/:db/collections/:collection/documents", server.createDocument)
		api.PUT("/databases/:db/collections/:collection/documents/:id", server.updateDocument)
		api.DELETE("/databases/:db/collections/:collection/documents/:id", server.deleteDocument)
		api.POST("/databases/:db/collections/:collection/documents/bulk-delete", server.bulkDeleteDocuments)
		api.GET("/databases/:db/collections/:collection/export", server.exportDocuments)
		api.POST("/databases/:db/collections/:collection/import", server.importDocuments)
		api.GET("/databases/:db/collections/:collection/backup", server.backupCollection)
		api.POST("/databases/:db/collections/:collection/restore", server.restoreCollection)
		api.GET("/databases/:db/collections/:collection/indexes", server.listIndexes)
		api.POST("/databases/:db/collections/:collection/indexes", server.createIndex)
		api.DELETE("/databases/:db/collections/:collection/indexes/:name", server.deleteIndex)
	}

	return router
}

func (s *Server) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) testConnection(c *gin.Context) {
	var request model.ConnectionTestRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	client, err := s.pool.Get(mongodb.ConnectionConfig{
		Host:       request.Host,
		Port:       request.Port,
		Database:   request.Database,
		Username:   request.Username,
		Password:   request.Password,
		AuthSource: request.AuthSource,
	})
	if err != nil {
		respondMongoError(c, "connection_test_failed", err)
		return
	}

	result, err := service.NewMetadataService(client).Health()
	if err != nil {
		respondMongoError(c, "connection_test_failed", err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) connection(c *gin.Context) {
	meta, ok := s.metaService(c)
	if !ok {
		return
	}

	result, err := meta.Health()
	if err != nil {
		respondMongoError(c, "connection_failed", err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) listDatabases(c *gin.Context) {
	meta, ok := s.metaService(c)
	if !ok {
		return
	}

	items, err := meta.ListDatabases()
	if err != nil {
		respondMongoError(c, "list_databases_failed", err)
		return
	}
	c.JSON(http.StatusOK, model.DatabaseListResponse{Items: items})
}

func (s *Server) createDatabase(c *gin.Context) {
	meta, ok := s.metaService(c)
	if !ok {
		return
	}

	var request model.CreateDatabaseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	if err := meta.CreateDatabase(request.Database, request.FirstCollection); err != nil {
		respondMongoError(c, "create_database_failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":           "created",
		"database":         request.Database,
		"first_collection": request.FirstCollection,
	})
}

func (s *Server) deleteDatabase(c *gin.Context) {
	meta, ok := s.metaService(c)
	if !ok {
		return
	}

	if err := meta.DeleteDatabase(c.Param("db")); err != nil {
		respondMongoError(c, "delete_database_failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (s *Server) listCollections(c *gin.Context) {
	meta, ok := s.metaService(c)
	if !ok {
		return
	}

	includeSystem := parseBool(c.DefaultQuery("include_system", "false"))
	includeCounts := parseBool(c.DefaultQuery("include_counts", "false"))

	items, err := meta.ListCollections(c.Param("db"), includeSystem, includeCounts)
	if err != nil {
		respondMongoError(c, "list_collections_failed", err)
		return
	}
	c.JSON(http.StatusOK, model.CollectionListResponse{Items: items})
}

func (s *Server) createCollection(c *gin.Context) {
	meta, ok := s.metaService(c)
	if !ok {
		return
	}

	var request model.CreateCollectionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	if err := meta.CreateCollection(c.Param("db"), request.Name); err != nil {
		respondMongoError(c, "create_collection_failed", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "created", "name": request.Name})
}

func (s *Server) deleteCollection(c *gin.Context) {
	meta, ok := s.metaService(c)
	if !ok {
		return
	}

	if err := meta.DeleteCollection(c.Param("db"), c.Param("collection")); err != nil {
		respondMongoError(c, "delete_collection_failed", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (s *Server) listDocuments(c *gin.Context) {
	documents, ok := s.documentService(c)
	if !ok {
		return
	}

	page := int64(readInt(c.DefaultQuery("page", "1"), 1))
	pageSize := int64(readInt(c.DefaultQuery("page_size", "20"), 20))

	result, err := documents.List(
		c.Param("db"),
		c.Param("collection"),
		c.Query("filter"),
		c.Query("conditions"),
		c.DefaultQuery("logic", "and"),
		page,
		pageSize,
	)
	if err != nil {
		respondError(c, http.StatusBadRequest, "list_documents_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, result)
}

func (s *Server) getDocument(c *gin.Context) {
	documents, ok := s.documentService(c)
	if !ok {
		return
	}

	document, err := documents.GetByID(c.Param("db"), c.Param("collection"), c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "get_document_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": document})
}

func (s *Server) createDocument(c *gin.Context) {
	documents, ok := s.documentService(c)
	if !ok {
		return
	}

	var request model.WriteDocumentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	document, err := documents.Create(c.Param("db"), c.Param("collection"), request.Document)
	if err != nil {
		respondError(c, http.StatusBadRequest, "create_document_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": document})
}

func (s *Server) updateDocument(c *gin.Context) {
	documents, ok := s.documentService(c)
	if !ok {
		return
	}

	var request model.WriteDocumentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	document, err := documents.Update(c.Param("db"), c.Param("collection"), c.Param("id"), request.Document)
	if err != nil {
		respondError(c, http.StatusBadRequest, "update_document_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": document})
}

func (s *Server) deleteDocument(c *gin.Context) {
	documents, ok := s.documentService(c)
	if !ok {
		return
	}

	if err := documents.Delete(c.Param("db"), c.Param("collection"), c.Param("id")); err != nil {
		respondError(c, http.StatusBadRequest, "delete_document_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (s *Server) bulkDeleteDocuments(c *gin.Context) {
	documents, ok := s.documentService(c)
	if !ok {
		return
	}

	var request model.BulkDeleteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	deletedCount, err := documents.BulkDelete(c.Param("db"), c.Param("collection"), request.IDs)
	if err != nil {
		respondError(c, http.StatusBadRequest, "bulk_delete_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted", "deleted_count": deletedCount})
}

func (s *Server) exportDocuments(c *gin.Context) {
	documents, ok := s.documentService(c)
	if !ok {
		return
	}

	format := strings.ToLower(c.DefaultQuery("format", "json"))
	selectedIDs := c.QueryArray("ids")

	var (
		content     []byte
		contentType string
		err         error
	)

	switch format {
	case "csv":
		if len(selectedIDs) > 0 {
			content, err = documents.ExportSelectedCSV(c.Param("db"), c.Param("collection"), selectedIDs)
		} else {
			content, err = documents.ExportCSV(c.Param("db"), c.Param("collection"), c.Query("filter"), c.Query("conditions"), c.DefaultQuery("logic", "and"))
		}
		contentType = "text/csv; charset=utf-8"
	default:
		if len(selectedIDs) > 0 {
			content, err = documents.ExportSelectedJSON(c.Param("db"), c.Param("collection"), selectedIDs)
		} else {
			content, err = documents.ExportJSON(c.Param("db"), c.Param("collection"), c.Query("filter"), c.Query("conditions"), c.DefaultQuery("logic", "and"))
		}
		contentType = "application/json; charset=utf-8"
		format = "json"
	}

	if err != nil {
		respondError(c, http.StatusInternalServerError, "export_documents_failed", err.Error())
		return
	}

	fileName := fmt.Sprintf("%s_%s_%s.%s", c.Param("db"), c.Param("collection"), "export", format)
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	c.Data(http.StatusOK, contentType, content)
}

func (s *Server) importDocuments(c *gin.Context) {
	documents, ok := s.documentService(c)
	if !ok {
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", "import file is required")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "open_import_file_failed", err.Error())
		return
	}
	defer file.Close()

	format := strings.ToLower(c.PostForm("format"))
	if format == "" {
		format = strings.TrimPrefix(strings.ToLower(filepath.Ext(fileHeader.Filename)), ".")
	}

	insertedCount, err := documents.ImportDocuments(c.Param("db"), c.Param("collection"), file, format)
	if err != nil {
		respondError(c, http.StatusBadRequest, "import_documents_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "imported", "inserted_count": insertedCount, "format": format})
}

func (s *Server) backupCollection(c *gin.Context) {
	documents, ok := s.documentService(c)
	if !ok {
		return
	}

	content, err := documents.BackupCollection(c.Param("db"), c.Param("collection"))
	if err != nil {
		respondError(c, http.StatusInternalServerError, "backup_collection_failed", err.Error())
		return
	}

	fileName := fmt.Sprintf("%s_%s_backup.json", c.Param("db"), c.Param("collection"))
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	c.Data(http.StatusOK, "application/json; charset=utf-8", content)
}

func (s *Server) restoreCollection(c *gin.Context) {
	documents, ok := s.documentService(c)
	if !ok {
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", "backup file is required")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "open_restore_file_failed", err.Error())
		return
	}
	defer file.Close()

	restoredCount, err := documents.RestoreCollection(c.Param("db"), c.Param("collection"), file)
	if err != nil {
		respondError(c, http.StatusBadRequest, "restore_collection_failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "restored", "restored_count": restoredCount})
}

func (s *Server) listIndexes(c *gin.Context) {
	documents, ok := s.documentService(c)
	if !ok {
		return
	}

	items, err := documents.ListIndexes(c.Param("db"), c.Param("collection"))
	if err != nil {
		respondError(c, http.StatusInternalServerError, "list_indexes_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, model.IndexListResponse{Items: items})
}

func (s *Server) createIndex(c *gin.Context) {
	documents, ok := s.documentService(c)
	if !ok {
		return
	}

	var request model.CreateIndexRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	name, err := documents.CreateIndex(c.Param("db"), c.Param("collection"), request)
	if err != nil {
		respondError(c, http.StatusBadRequest, "create_index_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "created", "name": name})
}

func (s *Server) deleteIndex(c *gin.Context) {
	documents, ok := s.documentService(c)
	if !ok {
		return
	}

	if err := documents.DeleteIndex(c.Param("db"), c.Param("collection"), c.Param("name")); err != nil {
		respondError(c, http.StatusBadRequest, "delete_index_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (s *Server) connectionConfigFromRequest(c *gin.Context) mongodb.ConnectionConfig {
	defaultConn := s.pool.DefaultConnection()

	return mongodb.ConnectionConfig{
		Host:       valueOrHeader(c.GetHeader(headerMongoHost), defaultConn.Host),
		Port:       valueOrHeader(c.GetHeader(headerMongoPort), defaultConn.Port),
		Database:   valueOrHeader(c.GetHeader(headerMongoDatabase), defaultConn.Database),
		Username:   strings.TrimSpace(c.GetHeader(headerMongoUsername)),
		Password:   c.GetHeader(headerMongoPassword),
		AuthSource: valueOrHeader(c.GetHeader(headerMongoAuthSource), defaultConn.AuthSource),
	}.Normalize()
}

func (s *Server) clientForRequest(c *gin.Context) (*mongodb.Client, bool) {
	client, err := s.pool.Get(s.connectionConfigFromRequest(c))
	if err != nil {
		respondMongoError(c, "invalid_connection", err)
		return nil, false
	}
	return client, true
}

func (s *Server) metaService(c *gin.Context) (*service.MetadataService, bool) {
	client, ok := s.clientForRequest(c)
	if !ok {
		return nil, false
	}
	return service.NewMetadataService(client), true
}

func (s *Server) documentService(c *gin.Context) (*service.DocumentService, bool) {
	client, ok := s.clientForRequest(c)
	if !ok {
		return nil, false
	}
	return service.NewDocumentService(client), true
}

func valueOrHeader(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func respondError(c *gin.Context, status int, code, message string) {
	c.JSON(status, model.ErrorResponse{
		Code:    code,
		Message: message,
	})
}

func respondMongoError(c *gin.Context, code string, err error) {
	status, message := classifyMongoError(err)
	respondError(c, status, code, message)
}

func classifyMongoError(err error) (int, string) {
	if err == nil {
		return http.StatusInternalServerError, "unknown server error"
	}

	message := strings.TrimSpace(err.Error())
	lower := strings.ToLower(message)

	switch {
	case errors.Is(err, mongo.ErrNoDocuments):
		return http.StatusNotFound, "requested resource was not found"
	case strings.Contains(lower, "auth error"),
		strings.Contains(lower, "authentication failed"),
		strings.Contains(lower, "auth failed"),
		strings.Contains(lower, "unable to authenticate"),
		strings.Contains(lower, "sasl"),
		strings.Contains(lower, "scram"):
		return http.StatusUnauthorized, "MongoDB authentication failed. Check username, password, and authSource."
	case strings.Contains(lower, "not authorized"),
		strings.Contains(lower, "unauthorized"),
		strings.Contains(lower, "requires authentication"),
		strings.Contains(lower, "command create requires authentication"),
		strings.Contains(lower, "command dropdatabase requires authentication"):
		return http.StatusForbidden, "Current MongoDB user does not have permission to perform this operation."
	case strings.Contains(lower, "server selection error"),
		strings.Contains(lower, "connection refused"),
		strings.Contains(lower, "no such host"),
		strings.Contains(lower, "i/o timeout"),
		strings.Contains(lower, "context deadline exceeded"),
		strings.Contains(lower, "network is unreachable"):
		return http.StatusBadGateway, "Unable to connect to the target MongoDB server. Check host, port, and network reachability."
	default:
		return http.StatusBadRequest, message
	}
}

func parseBool(value string) bool {
	parsed, err := strconv.ParseBool(value)
	return err == nil && parsed
}

func readInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
