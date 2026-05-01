package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"mongoDB_visual/backend/internal/model"
	"mongoDB_visual/backend/internal/mongodb"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type MetadataService struct {
	client *mongodb.Client
}

func NewMetadataService(client *mongodb.Client) *MetadataService {
	return &MetadataService{client: client}
}

func (s *MetadataService) Health() (*model.ConnectionResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	info, duration, err := s.client.HealthInfo(ctx)
	if err != nil {
		return nil, err
	}

	return &model.ConnectionResponse{
		Status:      "ok",
		Version:     info["version"].(string),
		Database:    info["database"].(string),
		Server:      info["server"].(string),
		RoundTripMS: duration.Milliseconds(),
	}, nil
}

func (s *MetadataService) ListDatabases() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	names, err := s.client.Raw().ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	filtered := make([]string, 0, len(names))
	for _, name := range names {
		if isSystemDatabase(name) {
			continue
		}
		filtered = append(filtered, name)
	}

	sort.Strings(filtered)
	return filtered, nil
}

func (s *MetadataService) ListCollections(database string, includeSystem, includeCounts bool) ([]model.CollectionItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	names, err := s.client.Database(database).ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	sort.Strings(names)
	items := make([]model.CollectionItem, 0, len(names))
	for _, name := range names {
		isSystem := mongodb.IsSystemCollection(name)
		if isSystem && !includeSystem {
			continue
		}

		item := model.CollectionItem{
			Name:     name,
			IsSystem: isSystem,
		}

		if includeCounts {
			count, err := s.client.Database(database).Collection(name).CountDocuments(ctx, bson.M{})
			if err == nil {
				item.DocumentCount = &count
			}
		}

		items = append(items, item)
	}

	return items, nil
}

func (s *MetadataService) CreateCollection(database, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.client.Database(database).CreateCollection(ctx, name)
}

func (s *MetadataService) DeleteCollection(database, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.client.Database(database).Collection(name).Drop(ctx)
}

func (s *MetadataService) CreateDatabase(database, firstCollection string) error {
	if isSystemDatabase(database) {
		return fmt.Errorf("system database %q cannot be created from the workspace", database)
	}
	return s.CreateCollection(database, firstCollection)
}

func (s *MetadataService) DeleteDatabase(database string) error {
	if isSystemDatabase(database) {
		return fmt.Errorf("system database %q cannot be deleted", database)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return s.client.Database(database).Drop(ctx)
}

func isSystemDatabase(name string) bool {
	switch name {
	case "admin", "config", "local":
		return true
	default:
		return false
	}
}
