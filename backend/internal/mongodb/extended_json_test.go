package mongodb

import (
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestNormalizeDocumentSpecialValues(t *testing.T) {
	doc, err := NormalizeDocument(map[string]any{
		"_id": map[string]any{"$oid": "507f1f77bcf86cd799439011"},
		"ts":  map[string]any{"$date": "2024-01-01T00:00:00Z"},
	})
	if err != nil {
		t.Fatalf("NormalizeDocument returned error: %v", err)
	}

	if _, ok := doc["_id"].(bson.ObjectID); !ok {
		t.Fatalf("expected _id to be ObjectID, got %T", doc["_id"])
	}

	if _, ok := doc["ts"].(time.Time); !ok {
		t.Fatalf("expected ts to be time.Time, got %T", doc["ts"])
	}
}

func TestParseFilterInvalidJSON(t *testing.T) {
	if _, err := ParseJSONFilter("{invalid}"); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseImportDocumentsConvertsHexStringID(t *testing.T) {
	documents, err := ParseImportDocuments(strings.NewReader(`[{"_id":"507f1f77bcf86cd799439011","name":"Alice"}]`), "json")
	if err != nil {
		t.Fatalf("ParseImportDocuments returned error: %v", err)
	}

	if _, ok := documents[0]["_id"].(bson.ObjectID); !ok {
		t.Fatalf("expected _id to be ObjectID, got %T", documents[0]["_id"])
	}
}

func TestParseImportDocumentsRejectsCollectionBackup(t *testing.T) {
	_, err := ParseImportDocuments(strings.NewReader(`{
		"database": "demo",
		"collection": "users",
		"documents": []
	}`), "json")
	if err == nil {
		t.Fatal("expected collection backup import to fail")
	}
}

func TestParseImportDocumentsRejectsInvalidStringID(t *testing.T) {
	_, err := ParseImportDocuments(strings.NewReader(`[{"_id":"custom-id","name":"Alice"}]`), "json")
	if err == nil {
		t.Fatal("expected invalid string _id to fail")
	}
}
