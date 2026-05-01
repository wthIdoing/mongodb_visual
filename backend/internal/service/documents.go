package service

import (
	"context"
	"fmt"
	"io"
	"time"

	"mongoDB_visual/backend/internal/model"
	"mongoDB_visual/backend/internal/mongodb"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type DocumentService struct {
	client *mongodb.Client
}

func NewDocumentService(client *mongodb.Client) *DocumentService {
	return &DocumentService{client: client}
}

func (s *DocumentService) List(database, collection, filter, conditions, logic string, page, pageSize int64) (*model.PaginationResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	query, err := mongodb.BuildFilter(filter, conditions, logic)
	if err != nil {
		return nil, err
	}

	coll := s.client.Database(database).Collection(collection)
	total, err := coll.CountDocuments(ctx, query)
	if err != nil {
		return nil, err
	}

	opts := options.Find().
		SetSkip((page - 1) * pageSize).
		SetLimit(pageSize).
		SetSort(bson.D{{Key: "_id", Value: -1}})

	cursor, err := coll.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []map[string]any
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}

	return &model.PaginationResponse{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (s *DocumentService) ListByIDs(database, collection string, ids []string) ([]map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectIDs, err := objectIDsFromStrings(ids)
	if err != nil {
		return nil, err
	}

	cursor, err := s.client.Database(database).Collection(collection).Find(ctx, bson.M{
		"_id": bson.M{"$in": objectIDs},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []map[string]any
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *DocumentService) ListAll(database, collection, filter, conditions, logic string) ([]map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	query, err := mongodb.BuildFilter(filter, conditions, logic)
	if err != nil {
		return nil, err
	}

	cursor, err := s.client.Database(database).Collection(collection).Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []map[string]any
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *DocumentService) GetByID(database, collection, id string) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid object id: %w", err)
	}

	var document map[string]any
	err = s.client.Database(database).Collection(collection).
		FindOne(ctx, bson.M{"_id": objectID}).
		Decode(&document)
	if err != nil {
		return nil, err
	}

	return document, nil
}

func (s *DocumentService) Create(database, collection string, input map[string]any) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	doc, err := mongodb.NormalizeDocument(input)
	if err != nil {
		return nil, err
	}

	result, err := s.client.Database(database).Collection(collection).InsertOne(ctx, doc)
	if err != nil {
		return nil, err
	}
	doc["_id"] = result.InsertedID

	return doc, nil
}

func (s *DocumentService) Update(database, collection, id string, input map[string]any) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid object id: %w", err)
	}

	doc, err := mongodb.NormalizeDocument(input)
	if err != nil {
		return nil, err
	}
	doc["_id"] = objectID

	result, err := s.client.Database(database).Collection(collection).
		ReplaceOne(ctx, bson.M{"_id": objectID}, doc)
	if err != nil {
		return nil, err
	}
	if result.MatchedCount == 0 {
		return nil, fmt.Errorf("target document not found")
	}

	return doc, nil
}

func (s *DocumentService) Delete(database, collection, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid object id: %w", err)
	}

	result, err := s.client.Database(database).Collection(collection).
		DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("target document not found")
	}
	return nil
}

func (s *DocumentService) BulkDelete(database, collection string, ids []string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectIDs, err := objectIDsFromStrings(ids)
	if err != nil {
		return 0, err
	}

	result, err := s.client.Database(database).Collection(collection).DeleteMany(ctx, bson.M{
		"_id": bson.M{"$in": objectIDs},
	})
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

func (s *DocumentService) ExportJSON(database, collection, filter, conditions, logic string) ([]byte, error) {
	items, err := s.ListAll(database, collection, filter, conditions, logic)
	if err != nil {
		return nil, err
	}
	return mongodb.ExportJSON(items)
}

func (s *DocumentService) ExportCSV(database, collection, filter, conditions, logic string) ([]byte, error) {
	items, err := s.ListAll(database, collection, filter, conditions, logic)
	if err != nil {
		return nil, err
	}
	return mongodb.ExportCSV(items)
}

func (s *DocumentService) ExportSelectedJSON(database, collection string, ids []string) ([]byte, error) {
	items, err := s.ListByIDs(database, collection, ids)
	if err != nil {
		return nil, err
	}
	return mongodb.ExportJSON(items)
}

func (s *DocumentService) ExportSelectedCSV(database, collection string, ids []string) ([]byte, error) {
	items, err := s.ListByIDs(database, collection, ids)
	if err != nil {
		return nil, err
	}
	return mongodb.ExportCSV(items)
}

func (s *DocumentService) BackupCollection(database, collection string) ([]byte, error) {
	items, err := s.ListAll(database, collection, "", "", "and")
	if err != nil {
		return nil, err
	}
	return mongodb.ExportCollectionBackup(database, collection, items)
}

func (s *DocumentService) ImportDocuments(database, collection string, reader io.Reader, format string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	documents, err := mongodb.ParseImportDocuments(reader, format)
	if err != nil {
		return 0, err
	}
	if len(documents) == 0 {
		return 0, fmt.Errorf("import file does not contain documents")
	}

	payload := make([]any, 0, len(documents))
	for _, document := range documents {
		payload = append(payload, document)
	}

	result, err := s.client.Database(database).Collection(collection).InsertMany(ctx, payload)
	if err != nil {
		return 0, err
	}

	return len(result.InsertedIDs), nil
}

func (s *DocumentService) RestoreCollection(database, collection string, reader io.Reader) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	_, documents, err := mongodb.ParseCollectionBackup(reader)
	if err != nil {
		return 0, err
	}

	coll := s.client.Database(database).Collection(collection)
	if _, err := coll.DeleteMany(ctx, bson.M{}); err != nil {
		return 0, err
	}

	if len(documents) == 0 {
		return 0, nil
	}

	payload := make([]any, 0, len(documents))
	for _, document := range documents {
		payload = append(payload, document)
	}

	result, err := coll.InsertMany(ctx, payload)
	if err != nil {
		return 0, err
	}
	return len(result.InsertedIDs), nil
}

func (s *DocumentService) ListIndexes(database, collection string) ([]model.IndexItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cursor, err := s.client.Database(database).Collection(collection).Indexes().List(ctx)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rawIndexes []map[string]any
	if err := cursor.All(ctx, &rawIndexes); err != nil {
		return nil, err
	}

	items := make([]model.IndexItem, 0, len(rawIndexes))
	for _, raw := range rawIndexes {
		keys := map[string]any{}
		if keyMap, ok := raw["key"].(map[string]any); ok {
			keys = keyMap
		} else if keyMap, ok := raw["key"].(bson.M); ok {
			for key, value := range keyMap {
				keys[key] = value
			}
		}

		item := model.IndexItem{
			Name:   fmt.Sprintf("%v", raw["name"]),
			Keys:   keys,
			Unique: toBool(raw["unique"]),
			Sparse: toBool(raw["sparse"]),
		}
		items = append(items, item)
	}

	return items, nil
}

func (s *DocumentService) CreateIndex(database, collection string, request model.CreateIndexRequest) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	order := request.Order
	if order != -1 {
		order = 1
	}

	name, err := s.client.Database(database).Collection(collection).Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.D{{Key: request.Field, Value: order}},
			Options: options.Index().
				SetName(request.Name).
				SetUnique(request.Unique).
				SetSparse(request.Sparse),
		},
	)
	if err != nil {
		return "", err
	}
	return name, nil
}

func (s *DocumentService) DeleteIndex(database, collection, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if name == "_id_" {
		return fmt.Errorf("default _id index cannot be deleted")
	}

	return s.client.Database(database).Collection(collection).Indexes().DropOne(ctx, name)
}

func objectIDsFromStrings(ids []string) ([]bson.ObjectID, error) {
	objectIDs := make([]bson.ObjectID, 0, len(ids))
	for _, id := range ids {
		objectID, err := bson.ObjectIDFromHex(id)
		if err != nil {
			return nil, fmt.Errorf("invalid object id: %w", err)
		}
		objectIDs = append(objectIDs, objectID)
	}
	return objectIDs, nil
}

func toBool(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	default:
		return false
	}
}
