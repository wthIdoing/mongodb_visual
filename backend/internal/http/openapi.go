package http

const openAPISpec = `{
  "openapi": "3.0.3",
  "info": {
    "title": "MongoDB Visual API",
    "version": "1.0.0"
  },
  "servers": [
    { "url": "http://localhost:8080" }
  ],
  "paths": {
    "/api/v1/health": {
      "get": {
        "summary": "Health check",
        "responses": { "200": { "description": "OK" } }
      }
    },
    "/api/v1/meta/connection": {
      "get": {
        "summary": "Connection metadata",
        "responses": { "200": { "description": "OK" } }
      }
    },
    "/api/v1/databases": {
      "get": {
        "summary": "List databases",
        "responses": { "200": { "description": "OK" } }
      }
    },
    "/api/v1/databases/{db}/collections": {
      "get": {
        "summary": "List collections",
        "parameters": [{ "name": "db", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": { "200": { "description": "OK" } }
      },
      "post": {
        "summary": "Create collection",
        "parameters": [{ "name": "db", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": { "200": { "description": "OK" } }
      }
    },
    "/api/v1/databases/{db}/collections/{collection}": {
      "delete": {
        "summary": "Delete collection",
        "parameters": [
          { "name": "db", "in": "path", "required": true, "schema": { "type": "string" } },
          { "name": "collection", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": { "200": { "description": "OK" } }
      }
    },
    "/api/v1/databases/{db}/collections/{collection}/documents": {
      "get": {
        "summary": "List documents",
        "parameters": [
          { "name": "db", "in": "path", "required": true, "schema": { "type": "string" } },
          { "name": "collection", "in": "path", "required": true, "schema": { "type": "string" } },
          { "name": "page", "in": "query", "schema": { "type": "integer" } },
          { "name": "page_size", "in": "query", "schema": { "type": "integer" } },
          { "name": "filter", "in": "query", "schema": { "type": "string" } }
        ],
        "responses": { "200": { "description": "OK" } }
      },
      "post": {
        "summary": "Create document",
        "responses": { "200": { "description": "OK" } }
      }
    },
    "/api/v1/databases/{db}/collections/{collection}/documents/{id}": {
      "get": {
        "summary": "Get document by id",
        "responses": { "200": { "description": "OK" } }
      },
      "put": {
        "summary": "Replace document by id",
        "responses": { "200": { "description": "OK" } }
      },
      "delete": {
        "summary": "Delete document by id",
        "responses": { "200": { "description": "OK" } }
      }
    }
  }
}`
