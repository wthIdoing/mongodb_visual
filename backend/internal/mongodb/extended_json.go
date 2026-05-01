package mongodb

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"mongoDB_visual/backend/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func NormalizeDocument(input map[string]any) (bson.M, error) {
	out, err := normalizeValue(input)
	if err != nil {
		return nil, err
	}

	doc, ok := out.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("document must be an object")
	}
	return bson.M(doc), nil
}

func ParseJSONFilter(filter string) (bson.M, error) {
	if filter == "" {
		return bson.M{}, nil
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(filter), &payload); err != nil {
		return nil, fmt.Errorf("invalid filter JSON: %w", err)
	}

	return NormalizeDocument(payload)
}

func ParseConditions(raw string) ([]model.QueryCondition, error) {
	if raw == "" {
		return nil, nil
	}

	var conditions []model.QueryCondition
	if err := json.Unmarshal([]byte(raw), &conditions); err != nil {
		return nil, fmt.Errorf("invalid conditions JSON: %w", err)
	}
	return conditions, nil
}

func BuildFilter(filterJSON, conditionsJSON, logic string) (bson.M, error) {
	if strings.TrimSpace(conditionsJSON) != "" {
		conditions, err := ParseConditions(conditionsJSON)
		if err != nil {
			return nil, err
		}
		return BuildFilterFromConditions(conditions, logic)
	}
	return ParseJSONFilter(filterJSON)
}

func BuildFilterFromConditions(conditions []model.QueryCondition, logic string) (bson.M, error) {
	if len(conditions) == 0 {
		return bson.M{}, nil
	}

	clauses := make([]bson.M, 0, len(conditions))
	for _, condition := range conditions {
		if strings.TrimSpace(condition.Field) == "" || strings.TrimSpace(condition.Operator) == "" {
			continue
		}

		value, err := castConditionValue(condition.Value, condition.ValueType, condition.Operator)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", condition.Field, err)
		}

		switch condition.Operator {
		case "eq":
			clauses = append(clauses, bson.M{condition.Field: value})
		case "ne":
			clauses = append(clauses, bson.M{condition.Field: bson.M{"$ne": value}})
		case "gt":
			clauses = append(clauses, bson.M{condition.Field: bson.M{"$gt": value}})
		case "gte":
			clauses = append(clauses, bson.M{condition.Field: bson.M{"$gte": value}})
		case "lt":
			clauses = append(clauses, bson.M{condition.Field: bson.M{"$lt": value}})
		case "lte":
			clauses = append(clauses, bson.M{condition.Field: bson.M{"$lte": value}})
		case "contains":
			clauses = append(clauses, bson.M{condition.Field: bson.M{"$regex": fmt.Sprintf("%v", value), "$options": "i"}})
		case "in":
			clauses = append(clauses, bson.M{condition.Field: bson.M{"$in": value}})
		default:
			return nil, fmt.Errorf("unsupported operator %q", condition.Operator)
		}
	}

	activeConditions := make([]model.QueryCondition, 0, len(conditions))
	for _, condition := range conditions {
		if strings.TrimSpace(condition.Field) == "" || strings.TrimSpace(condition.Operator) == "" {
			continue
		}
		activeConditions = append(activeConditions, condition)
	}

	if len(clauses) == 0 || len(activeConditions) == 0 {
		return bson.M{}, nil
	}

	expr := clauses[0]
	if len(clauses) == 1 {
		return expr, nil
	}

	defaultJoin := logic
	if defaultJoin == "" {
		defaultJoin = "and"
	}

	for index := 1; index < len(clauses); index++ {
		join := strings.ToLower(strings.TrimSpace(activeConditions[index-1].Join))
		if join == "" {
			join = strings.ToLower(defaultJoin)
		}
		if join == "or" {
			expr = bson.M{"$or": []bson.M{expr, clauses[index]}}
			continue
		}
		expr = bson.M{"$and": []bson.M{expr, clauses[index]}}
	}

	return expr, nil
}

func ParseImportDocuments(reader io.Reader, format string) ([]bson.M, error) {
	switch strings.ToLower(format) {
	case "json":
		return parseJSONDocuments(reader)
	case "csv":
		return parseCSVDocuments(reader)
	default:
		return nil, fmt.Errorf("unsupported import format %q", format)
	}
}

func ParseCollectionBackup(reader io.Reader) (*model.CollectionBackup, []bson.M, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, nil, err
	}

	var backup model.CollectionBackup
	if err := json.Unmarshal(content, &backup); err != nil {
		return nil, nil, fmt.Errorf("invalid backup payload")
	}

	if backup.Collection == "" {
		return nil, nil, fmt.Errorf("backup payload missing collection")
	}

	documents, err := normalizeDocuments(backup.Documents)
	if err != nil {
		return nil, nil, err
	}
	return &backup, documents, nil
}

func ExportJSON(items []map[string]any) ([]byte, error) {
	return json.MarshalIndent(items, "", "  ")
}

func ExportCollectionBackup(database, collection string, items []map[string]any) ([]byte, error) {
	payload := model.CollectionBackup{
		Database:   database,
		Collection: collection,
		ExportedAt: time.Now().Format(time.RFC3339),
		Documents:  items,
	}
	return json.MarshalIndent(payload, "", "  ")
}

func ExportCSV(items []map[string]any) ([]byte, error) {
	headersMap := make(map[string]struct{})
	for _, item := range items {
		for key := range item {
			headersMap[key] = struct{}{}
		}
	}

	headers := make([]string, 0, len(headersMap))
	for key := range headersMap {
		headers = append(headers, key)
	}
	sort.Strings(headers)

	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	for _, item := range items {
		row := make([]string, 0, len(headers))
		for _, key := range headers {
			row = append(row, flattenCSVValue(item[key]))
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return []byte(builder.String()), nil
}

func IsSystemCollection(name string) bool {
	if strings.HasPrefix(name, "system.") {
		return true
	}
	switch name {
	case "startup_log":
		return true
	default:
		return false
	}
}

func normalizeValue(value any) (any, error) {
	switch typed := value.(type) {
	case map[string]any:
		if converted, matched, err := trySpecialValue(typed); matched || err != nil {
			return converted, err
		}

		out := make(map[string]any, len(typed))
		for key, item := range typed {
			normalized, err := normalizeValue(item)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", key, err)
			}
			out[key] = normalized
		}
		return out, nil
	case []any:
		out := make([]any, len(typed))
		for index, item := range typed {
			normalized, err := normalizeValue(item)
			if err != nil {
				return nil, fmt.Errorf("[%d]: %w", index, err)
			}
			out[index] = normalized
		}
		return out, nil
	default:
		return value, nil
	}
}

func trySpecialValue(value map[string]any) (any, bool, error) {
	if len(value) != 1 {
		return nil, false, nil
	}

	if raw, ok := value["$oid"]; ok {
		str, ok := raw.(string)
		if !ok {
			return nil, true, fmt.Errorf("$oid must be a string")
		}
		id, err := bson.ObjectIDFromHex(str)
		if err != nil {
			return nil, true, err
		}
		return id, true, nil
	}

	if raw, ok := value["$date"]; ok {
		str, ok := raw.(string)
		if !ok {
			return nil, true, fmt.Errorf("$date must be a string")
		}
		t, err := time.Parse(time.RFC3339, str)
		if err != nil {
			return nil, true, err
		}
		return t, true, nil
	}

	return nil, false, nil
}

func castConditionValue(value any, valueType, operator string) (any, error) {
	if operator == "in" {
		switch typed := value.(type) {
		case []any:
			out := make([]any, 0, len(typed))
			for _, item := range typed {
				parsed, err := castScalarValue(item, valueType)
				if err != nil {
					return nil, err
				}
				out = append(out, parsed)
			}
			return out, nil
		case string:
			parts := strings.Split(typed, ",")
			out := make([]any, 0, len(parts))
			for _, part := range parts {
				parsed, err := castScalarValue(strings.TrimSpace(part), valueType)
				if err != nil {
					return nil, err
				}
				out = append(out, parsed)
			}
			return out, nil
		default:
			parsed, err := castScalarValue(value, valueType)
			if err != nil {
				return nil, err
			}
			return []any{parsed}, nil
		}
	}

	return castScalarValue(value, valueType)
}

func castScalarValue(value any, valueType string) (any, error) {
	switch strings.ToLower(valueType) {
	case "number":
		switch typed := value.(type) {
		case float64, float32, int, int32, int64:
			return typed, nil
		case string:
			if strings.Contains(typed, ".") {
				return strconv.ParseFloat(typed, 64)
			}
			return strconv.ParseInt(typed, 10, 64)
		default:
			return nil, fmt.Errorf("invalid number value")
		}
	case "boolean":
		switch typed := value.(type) {
		case bool:
			return typed, nil
		case string:
			return strconv.ParseBool(typed)
		default:
			return nil, fmt.Errorf("invalid boolean value")
		}
	case "date":
		str, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("invalid date value")
		}
		return time.Parse(time.RFC3339, str)
	case "objectid":
		str, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("invalid ObjectId value")
		}
		return bson.ObjectIDFromHex(str)
	case "null":
		return nil, nil
	default:
		return value, nil
	}
}

func parseJSONDocuments(reader io.Reader) ([]bson.M, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var arrayPayload []map[string]any
	if err := json.Unmarshal(content, &arrayPayload); err == nil {
		return normalizeDocuments(arrayPayload)
	}

	var singlePayload map[string]any
	if err := json.Unmarshal(content, &singlePayload); err != nil {
		return nil, fmt.Errorf("invalid JSON import payload")
	}
	if isCollectionBackupPayload(singlePayload) {
		return nil, fmt.Errorf("collection backup JSON must be restored with restore collection, not imported as document")
	}
	return normalizeDocuments([]map[string]any{singlePayload})
}

func parseCSVDocuments(reader io.Reader) ([]bson.M, error) {
	csvReader := csv.NewReader(reader)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("csv must contain header and at least one data row")
	}

	headers := records[0]
	items := make([]map[string]any, 0, len(records)-1)
	for _, record := range records[1:] {
		item := make(map[string]any, len(headers))
		for index, header := range headers {
			value := ""
			if index < len(record) {
				value = record[index]
			}
			item[header] = inferCSVValue(value)
		}
		items = append(items, item)
	}
	return normalizeDocuments(items)
}

func normalizeDocuments(items []map[string]any) ([]bson.M, error) {
	out := make([]bson.M, 0, len(items))
	for _, item := range items {
		normalized, err := NormalizeDocument(item)
		if err != nil {
			return nil, err
		}
		if err := normalizeImportedID(normalized); err != nil {
			return nil, err
		}
		out = append(out, normalized)
	}
	return out, nil
}

func isCollectionBackupPayload(payload map[string]any) bool {
	_, hasDocuments := payload["documents"]
	_, hasCollection := payload["collection"]
	return hasDocuments && hasCollection
}

func normalizeImportedID(document bson.M) error {
	value, exists := document["_id"]
	if !exists || value == nil {
		return nil
	}

	switch typed := value.(type) {
	case bson.ObjectID:
		return nil
	case string:
		objectID, err := bson.ObjectIDFromHex(typed)
		if err != nil {
			return fmt.Errorf("_id must be a valid ObjectId hex string")
		}
		document["_id"] = objectID
		return nil
	default:
		return fmt.Errorf("_id must be ObjectId or ObjectId hex string")
	}
}

func inferCSVValue(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if strings.EqualFold(trimmed, "null") {
		return nil
	}
	if booleanValue, err := strconv.ParseBool(trimmed); err == nil {
		return booleanValue
	}
	if intValue, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		return intValue
	}
	if floatValue, err := strconv.ParseFloat(trimmed, 64); err == nil && strings.Contains(trimmed, ".") {
		return floatValue
	}
	return trimmed
}

func flattenCSVValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case bson.ObjectID:
		return typed.Hex()
	case time.Time:
		return typed.Format(time.RFC3339)
	case fmt.Stringer:
		return typed.String()
	default:
		bytes, err := json.Marshal(typed)
		if err == nil {
			return string(bytes)
		}
		return fmt.Sprintf("%v", typed)
	}
}
