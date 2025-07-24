// ABOUTME: Qdrant vector database client with full gRPC client integration
// ABOUTME: Provides collection management, vector storage and similarity search using official Go client
package database

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/qdrant/go-client/qdrant"
)

type Qdrant struct {
	client *qdrant.Client
	config *qdrant.Config
}

type SimilarityResult struct {
	ID      string         `json:"id"`
	Score   float32        `json:"score"`
	Payload map[string]any `json:"payload,omitempty"`
}

type VectorPoint struct {
	ID      string         `json:"id"`
	Vector  []float32      `json:"vector"`
	Payload map[string]any `json:"payload,omitempty"`
}

// NewQdrant creates a new Qdrant client from a URL string
func NewQdrant(qdrantURL string) (*Qdrant, error) {
	// Parse the URL to extract host, port, and other connection details
	parsedURL, err := url.Parse(qdrantURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Qdrant URL: %w", err)
	}

	// Extract host and port
	host := parsedURL.Hostname()
	if host == "" {
		host = "localhost"
	}

	port := 6334 // Default gRPC port for Qdrant (HTTP is 6333)
	if parsedURL.Port() != "" {
		if p, err := strconv.Atoi(parsedURL.Port()); err == nil {
			port = p
		}
	}

	// Create client configuration
	config := &qdrant.Config{
		Host: host,
		Port: port,
	}

	// Check if TLS should be used (only https scheme or port 443)
	if parsedURL.Scheme == "https" || port == 443 {
		config.UseTLS = true
	}

	// For Docker development, explicitly disable TLS for gRPC ports with http scheme
	if parsedURL.Scheme == "http" {
		config.UseTLS = false
	}

	// Extract API key from query parameters if present
	if apiKey := parsedURL.Query().Get("api_key"); apiKey != "" {
		config.APIKey = apiKey
	}

	// Create the client
	client, err := qdrant.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Qdrant client: %w", err)
	}

	return &Qdrant{
		client: client,
		config: config,
	}, nil
}

func (q *Qdrant) Ping() error {
	ctx := context.Background()

	// Try to get a list of collections as a health check
	_, err := q.client.ListCollections(ctx)
	if err != nil {
		return fmt.Errorf("qdrant health check failed: %w", err)
	}

	return nil
}

func (q *Qdrant) CreateCollection(ctx context.Context, collectionName string, vectorDim int) error {
	err := q.client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     uint64(vectorDim),
			Distance: qdrant.Distance_Cosine, // Using cosine distance for code similarity
		}),
	})

	if err != nil {
		return fmt.Errorf("failed to create collection %s: %w", collectionName, err)
	}

	log.Printf("Successfully created Qdrant collection: %s with dimension: %d", collectionName, vectorDim)
	return nil
}

func (q *Qdrant) CollectionExists(ctx context.Context, collectionName string) (bool, error) {
	collections, err := q.client.ListCollections(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to list collections: %w", err)
	}

	for _, collection := range collections {
		if collection == collectionName {
			return true, nil
		}
	}

	return false, nil
}

func (q *Qdrant) UpsertPoints(ctx context.Context, collectionName string, points []VectorPoint) error {
	if len(points) == 0 {
		return nil
	}

	// Convert our VectorPoint structs to Qdrant PointStruct
	qdrantPoints := make([]*qdrant.PointStruct, len(points))
	for i, point := range points {
		// Create point ID from string
		var pointID *qdrant.PointId
		if point.ID != "" {
			pointID = qdrant.NewIDUUID(point.ID)
		} else {
			pointID = qdrant.NewIDNum(uint64(i + 1)) // Fallback to numeric ID
		}

		// Create vectors
		vectors := qdrant.NewVectors(point.Vector...)

		// Create payload if provided
		var payload map[string]*qdrant.Value
		if len(point.Payload) > 0 {
			// Sanitize payload to avoid unsupported types (e.g., []string) that cause panics in
			// qdrant.NewValueMap. We convert slices of concrete types (like []string) into []any so the
			// Qdrant helper recognises them as list values instead of panicking.
			sanitized := sanitizePayloadMap(point.Payload)
			payload = qdrant.NewValueMap(sanitized)
		}

		qdrantPoints[i] = &qdrant.PointStruct{
			Id:      pointID,
			Vectors: vectors,
			Payload: payload,
		}
	}

	// Upsert the points
	_, err := q.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Points:         qdrantPoints,
	})

	if err != nil {
		return fmt.Errorf("failed to upsert %d points to collection %s: %w", len(points), collectionName, err)
	}

	log.Printf("Successfully upserted %d points to collection: %s", len(points), collectionName)
	return nil
}

func (q *Qdrant) SearchSimilar(ctx context.Context, collectionName string, queryVector []float32, limit int, withPayload bool) ([]SimilarityResult, error) {
	// Prepare the query request
	queryRequest := &qdrant.QueryPoints{
		CollectionName: collectionName,
		Query:          qdrant.NewQuery(queryVector...),
		Limit:          &[]uint64{uint64(limit)}[0],
		WithPayload:    qdrant.NewWithPayload(withPayload),
	}

	// Execute the search
	result, err := q.client.Query(ctx, queryRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar vectors in collection %s: %w", collectionName, err)
	}

	// Convert results to our format
	similarityResults := make([]SimilarityResult, len(result))
	for i, point := range result {
		var pointID string
		if point.Id != nil {
			// Handle different point ID types
			switch id := point.Id.PointIdOptions.(type) {
			case *qdrant.PointId_Uuid:
				pointID = id.Uuid
			case *qdrant.PointId_Num:
				pointID = strconv.FormatUint(id.Num, 10)
			default:
				pointID = fmt.Sprintf("unknown_%d", i)
			}
		}

		var payload map[string]any
		if withPayload && point.Payload != nil {
			payload = convertValueMapToMap(point.Payload)
		}

		similarityResults[i] = SimilarityResult{
			ID:      pointID,
			Score:   point.Score,
			Payload: payload,
		}
	}

	log.Printf("Found %d similar vectors in collection: %s", len(similarityResults), collectionName)
	return similarityResults, nil
}

func (q *Qdrant) DeletePoints(ctx context.Context, collectionName string, pointIDs []string) error {
	if len(pointIDs) == 0 {
		return nil
	}

	// Convert string IDs to Qdrant PointId format
	qdrantIDs := make([]*qdrant.PointId, len(pointIDs))
	for i, id := range pointIDs {
		// Try to parse as UUID first, fallback to treating as string
		if isUUID(id) {
			qdrantIDs[i] = qdrant.NewIDUUID(id)
		} else {
			// Try to parse as number
			if num, err := strconv.ParseUint(id, 10, 64); err == nil {
				qdrantIDs[i] = qdrant.NewIDNum(num)
			} else {
				// Fallback to UUID string representation
				qdrantIDs[i] = qdrant.NewIDUUID(id)
			}
		}
	}

	// Delete the points using Delete method instead of DeletePoints
	_, err := q.client.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: collectionName,
		Points: &qdrant.PointsSelector{
			PointsSelectorOneOf: &qdrant.PointsSelector_Points{
				Points: &qdrant.PointsIdsList{
					Ids: qdrantIDs,
				},
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to delete %d points from collection %s: %w", len(pointIDs), collectionName, err)
	}

	log.Printf("Successfully deleted %d points from collection: %s", len(pointIDs), collectionName)
	return nil
}

func (q *Qdrant) BaseURL() string {
	scheme := "http"
	if q.config.UseTLS {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s:%d", scheme, q.config.Host, q.config.Port)
}

func (q *Qdrant) Close() error {
	// The qdrant.Client uses gRPC connections that should be closed
	// However, the current client doesn't expose a Close method
	// This is here for future compatibility
	if q.client != nil {
		log.Printf("Closing Qdrant client connection")
	}
	return nil
}

// Helper function to convert Qdrant ValueMap to Go map
func convertValueMapToMap(valueMap map[string]*qdrant.Value) map[string]any {
	if valueMap == nil {
		return nil
	}

	result := make(map[string]any)
	for key, value := range valueMap {
		result[key] = convertValueToInterface(value)
	}
	return result
}

// Helper function to convert Qdrant Struct to Go map
func convertStructToMap(s *qdrant.Struct) map[string]any {
	if s == nil || s.Fields == nil {
		return nil
	}

	result := make(map[string]any)
	for key, value := range s.Fields {
		result[key] = convertValueToInterface(value)
	}
	return result
}

// Helper function to convert Qdrant Value to Go interface{}
func convertValueToInterface(v *qdrant.Value) any {
	if v == nil {
		return nil
	}

	switch kind := v.Kind.(type) {
	case *qdrant.Value_StringValue:
		return kind.StringValue
	case *qdrant.Value_IntegerValue:
		return kind.IntegerValue
	case *qdrant.Value_DoubleValue:
		return kind.DoubleValue
	case *qdrant.Value_BoolValue:
		return kind.BoolValue
	case *qdrant.Value_StructValue:
		return convertStructToMap(kind.StructValue)
	case *qdrant.Value_ListValue:
		if kind.ListValue == nil || kind.ListValue.Values == nil {
			return []any{}
		}
		result := make([]any, len(kind.ListValue.Values))
		for i, val := range kind.ListValue.Values {
			result[i] = convertValueToInterface(val)
		}
		return result
	default:
		return nil
	}
}

// Helper function to check if a string is a valid UUID format
func isUUID(s string) bool {
	// Basic UUID format check (8-4-4-4-12 hex digits)
	if len(s) != 36 {
		return false
	}

	parts := strings.Split(s, "-")
	if len(parts) != 5 {
		return false
	}

	expectedLengths := []int{8, 4, 4, 4, 12}
	for i, part := range parts {
		if len(part) != expectedLengths[i] {
			return false
		}
		// Check if all characters are hex digits
		for _, char := range part {
			if (char < '0' || char > '9') && (char < 'a' || char > 'f') && (char < 'A' || char > 'F') {
				return false
			}
		}
	}

	return true
}

// sanitizePayloadMap walks through a payload map recursively and converts any slice with a concrete
// element type (e.g., []string) into a slice of interface{} (i.e., []any). The Qdrant Go client only
// supports []any for list values; passing an untyped concrete slice causes it to panic with
// "invalid type: []T".
func sanitizePayloadMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}

	sanitized := make(map[string]any, len(input))
	for k, v := range input {
		sanitized[k] = sanitizePayloadValue(v)
	}
	return sanitized
}

// sanitizePayloadValue converts unsupported value types to forms accepted by qdrant.NewValueMap.
// Currently this handles slices of strings (and other primitives) by converting them to []any and
// processes nested maps recursively.
func sanitizePayloadValue(v any) any {
	switch vv := v.(type) {
	case []string:
		out := make([]any, len(vv))
		for i, s := range vv {
			out[i] = s
		}
		return out
	case []int:
		out := make([]any, len(vv))
		for i, n := range vv {
			out[i] = n
		}
		return out
	case []float64:
		out := make([]any, len(vv))
		for i, n := range vv {
			out[i] = n
		}
		return out
	case []any:
		// Recursively sanitize each element
		for i, elem := range vv {
			vv[i] = sanitizePayloadValue(elem)
		}
		return vv
	case map[string]any:
		return sanitizePayloadMap(vv)
	default:
		return v
	}
}
