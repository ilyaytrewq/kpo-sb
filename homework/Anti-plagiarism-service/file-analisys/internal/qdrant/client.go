package qdrant

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	defaultTimeout  = 10 * time.Second
	defaultHTTPPort = 6333
	defaultGRPCPort = 6334

	envAddress    = "QDRANT_ADDRESS"
	envHost       = "QDRANT_HOST"
	envPort       = "QDRANT_PORT"
	envGRPCPort   = "QDRANT_GRPC_PORT"
	envAPIKey     = "QDRANT_API_KEY"
	envCollection = "QDRANT_COLLECTION"
	envVectorSize = "QDRANT_VECTOR_SIZE"
)

type Client struct {
	collection  string
	vectorSize  int
	conn        *grpc.ClientConn
	collections qdrant.CollectionsClient
	points      qdrant.PointsClient
}

type SearchResult struct {
	ID      string
	Score   float32
	Payload map[string]interface{}
}

type Config struct {
	Address    string
	Collection string
	VectorSize int
}

func NewClient(ctx context.Context, address, collection string, vectorSize int, dialOptions ...grpc.DialOption) (*Client, error) {

	if strings.TrimSpace(address) == "" {
		return nil, errors.New("qdrant address is empty")
	}
	if strings.TrimSpace(collection) == "" {
		return nil, errors.New("qdrant collection is empty")
	}
	if vectorSize <= 0 {
		return nil, errors.New("qdrant vector size must be positive")
	}
	if len(dialOptions) == 0 {
		dialOptions = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	}

	dialCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	conn, err := grpc.DialContext(dialCtx, address, dialOptions...)
	if err != nil {
		return nil, fmt.Errorf("dial qdrant: %w", err)
	}

	return &Client{
		collection:  collection,
		vectorSize:  vectorSize,
		conn:        conn,
		collections: qdrant.NewCollectionsClient(conn),
		points:      qdrant.NewPointsClient(conn),
	}, nil
}

func LoadConfigFromEnv() (*Config, error) {
	address := strings.TrimSpace(os.Getenv(envAddress))
	if address == "" {
		host := strings.TrimSpace(os.Getenv(envHost))
		if host == "" {
			return nil, fmt.Errorf("%s environment variable is not set", envHost)
		}

		portKey := envGRPCPort
		portStr := strings.TrimSpace(os.Getenv(envGRPCPort))
		if portStr == "" {
			return nil, fmt.Errorf("%s environment variable is not set", envGRPCPort)
		}

		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", portKey, err)
		}
		if port <= 0 {
			return nil, fmt.Errorf("%s must be positive", portKey)
		}

		address = net.JoinHostPort(host, strconv.Itoa(port))
	}

	collection := strings.TrimSpace(os.Getenv(envCollection))
	if collection == "" {
		return nil, fmt.Errorf("%s environment variable is not set", envCollection)
	}

	vectorSizeStr := strings.TrimSpace(os.Getenv(envVectorSize))
	if vectorSizeStr == "" {
		return nil, fmt.Errorf("%s environment variable is not set", envVectorSize)
	}
	vectorSize, err := strconv.Atoi(vectorSizeStr)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", envVectorSize, err)
	}
	if vectorSize <= 0 {
		return nil, fmt.Errorf("%s must be positive", envVectorSize)
	}

	return &Config{
		Address:    address,
		Collection: collection,
		VectorSize: vectorSize,
	}, nil
}

func NewClientFromEnv(ctx context.Context, dialOptions ...grpc.DialOption) (*Client, error) {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return nil, err
	}
	if len(dialOptions) == 0 {
		dialOptions = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	}
	if apiKey := strings.TrimSpace(os.Getenv(envAPIKey)); apiKey != "" {
		dialOptions = append(dialOptions, withAPIKey(apiKey))
	}
	return NewClient(ctx, cfg.Address, cfg.Collection, cfg.VectorSize, dialOptions...)
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) Connect(ctx context.Context) error {
	_, err := c.collections.Get(ctx, &qdrant.GetCollectionInfoRequest{})
	return err
}

func (c *Client) EnsureCollection(ctx context.Context) error {
	return c.EnsureCollectionNamed(ctx, c.collection)
}

func (c *Client) EnsureCollectionNamed(ctx context.Context, collection string) error {
	if strings.TrimSpace(collection) == "" {
		return errors.New("qdrant collection is empty")
	}

	_, err := c.collections.Get(ctx, &qdrant.GetCollectionInfoRequest{CollectionName: collection})
	if err == nil {
		return nil
	}
	if status.Code(err) != codes.NotFound {
		return err
	}

	_, err = c.collections.Create(ctx, &qdrant.CreateCollection{
		CollectionName: collection,
		VectorsConfig: &qdrant.VectorsConfig{
			Config: &qdrant.VectorsConfig_Params{
				Params: &qdrant.VectorParams{
					Size:     uint64(c.vectorSize),
					Distance: qdrant.Distance_Cosine,
				},
			},
		},
	})
	if err != nil && status.Code(err) == codes.AlreadyExists {
		return nil
	}
	return err
}

func (c *Client) Search(ctx context.Context, vector []float32, limit int) ([]SearchResult, error) {
	return c.SearchInCollection(ctx, c.collection, vector, limit)
}

func (c *Client) SearchInCollection(ctx context.Context, collection string, vector []float32, limit int) ([]SearchResult, error) {
	if strings.TrimSpace(collection) == "" {
		return nil, errors.New("qdrant collection is empty")
	}
	if len(vector) != c.vectorSize {
		return nil, fmt.Errorf("qdrant vector size mismatch: expected %d, got %d", c.vectorSize, len(vector))
	}
	if limit <= 0 {
		return nil, errors.New("qdrant search limit must be positive")
	}

	resp, err := c.points.Search(ctx, &qdrant.SearchPoints{
		CollectionName: collection,
		Vector:         vector,
		Limit:          uint64(limit),
		WithPayload: &qdrant.WithPayloadSelector{
			SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true},
		},
	})
	if err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, len(resp.Result))
	for _, point := range resp.Result {
		payload, err := payloadFromQdrant(point.Payload)
		if err != nil {
			return nil, err
		}
		results = append(results, SearchResult{
			ID:      pointIDString(point.Id),
			Score:   point.Score,
			Payload: payload,
		})
	}
	return results, nil
}

func (c *Client) AddVector(ctx context.Context, id string, vector []float32, payload map[string]interface{}) error {
	return c.AddVectorToCollection(ctx, c.collection, id, vector, payload)
}

func (c *Client) AddVectorToCollection(ctx context.Context, collection, id string, vector []float32, payload map[string]interface{}) error {
	if strings.TrimSpace(collection) == "" {
		return errors.New("qdrant collection is empty")
	}
	if strings.TrimSpace(id) == "" {
		return errors.New("qdrant id is empty")
	}
	if len(vector) != c.vectorSize {
		return fmt.Errorf("qdrant vector size mismatch: expected %d, got %d", c.vectorSize, len(vector))
	}

	payloadValue, err := payloadToQdrant(payload)
	if err != nil {
		return err
	}

	pointID, err := pointIDFromString(id)
	if err != nil {
		return err
	}

	wait := true
	_, err = c.points.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collection,
		Wait:           &wait,
		Points: []*qdrant.PointStruct{
			{
				Id: pointID,
				Vectors: &qdrant.Vectors{
					VectorsOptions: &qdrant.Vectors_Vector{
						Vector: &qdrant.Vector{
							Vector: &qdrant.Vector_Dense{
								Dense: &qdrant.DenseVector{Data: vector},
							},
						},
					},
				},
				Payload: payloadValue,
			},
		},
	})
	return err
}

func pointIDFromString(id string) (*qdrant.PointId, error) {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return nil, errors.New("qdrant id is empty")
	}
	if parsed, err := uuid.Parse(trimmed); err == nil {
		return &qdrant.PointId{
			PointIdOptions: &qdrant.PointId_Uuid{Uuid: parsed.String()},
		}, nil
	}
	if num, err := strconv.ParseUint(trimmed, 10, 64); err == nil {
		return &qdrant.PointId{
			PointIdOptions: &qdrant.PointId_Num{Num: num},
		}, nil
	}
	nameUUID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(trimmed))
	return &qdrant.PointId{
		PointIdOptions: &qdrant.PointId_Uuid{Uuid: nameUUID.String()},
	}, nil
}

func pointIDString(id *qdrant.PointId) string {
	if id == nil {
		return ""
	}
	switch value := id.PointIdOptions.(type) {
	case *qdrant.PointId_Uuid:
		return value.Uuid
	case *qdrant.PointId_Num:
		return fmt.Sprintf("%d", value.Num)
	default:
		return ""
	}
}

func payloadToQdrant(payload map[string]interface{}) (map[string]*qdrant.Value, error) {
	if payload == nil {
		return nil, nil
	}
	result := make(map[string]*qdrant.Value, len(payload))
	for key, value := range payload {
		converted, err := toQdrantValue(value)
		if err != nil {
			return nil, fmt.Errorf("payload key %q: %w", key, err)
		}
		result[key] = converted
	}
	return result, nil
}

func toQdrantValue(value interface{}) (*qdrant.Value, error) {
	switch typed := value.(type) {
	case nil:
		return &qdrant.Value{Kind: &qdrant.Value_NullValue{NullValue: 0}}, nil
	case bool:
		return &qdrant.Value{Kind: &qdrant.Value_BoolValue{BoolValue: typed}}, nil
	case int:
		return &qdrant.Value{Kind: &qdrant.Value_IntegerValue{IntegerValue: int64(typed)}}, nil
	case int32:
		return &qdrant.Value{Kind: &qdrant.Value_IntegerValue{IntegerValue: int64(typed)}}, nil
	case int64:
		return &qdrant.Value{Kind: &qdrant.Value_IntegerValue{IntegerValue: typed}}, nil
	case float32:
		return &qdrant.Value{Kind: &qdrant.Value_DoubleValue{DoubleValue: float64(typed)}}, nil
	case float64:
		return &qdrant.Value{Kind: &qdrant.Value_DoubleValue{DoubleValue: typed}}, nil
	case string:
		return &qdrant.Value{Kind: &qdrant.Value_StringValue{StringValue: typed}}, nil
	case []interface{}:
		values := make([]*qdrant.Value, 0, len(typed))
		for _, item := range typed {
			converted, err := toQdrantValue(item)
			if err != nil {
				return nil, err
			}
			values = append(values, converted)
		}
		return &qdrant.Value{Kind: &qdrant.Value_ListValue{ListValue: &qdrant.ListValue{Values: values}}}, nil
	case map[string]interface{}:
		fields := make(map[string]*qdrant.Value, len(typed))
		for key, item := range typed {
			converted, err := toQdrantValue(item)
			if err != nil {
				return nil, err
			}
			fields[key] = converted
		}
		return &qdrant.Value{Kind: &qdrant.Value_StructValue{StructValue: &qdrant.Struct{Fields: fields}}}, nil
	default:
		return nil, fmt.Errorf("unsupported payload type %T", value)
	}
}

func payloadFromQdrant(payload map[string]*qdrant.Value) (map[string]interface{}, error) {
	if payload == nil {
		return nil, nil
	}
	result := make(map[string]interface{}, len(payload))
	for key, value := range payload {
		decoded, err := fromQdrantValue(value)
		if err != nil {
			return nil, fmt.Errorf("payload key %q: %w", key, err)
		}
		result[key] = decoded
	}
	return result, nil
}

func fromQdrantValue(value *qdrant.Value) (interface{}, error) {
	if value == nil {
		return nil, nil
	}
	switch typed := value.Kind.(type) {
	case *qdrant.Value_NullValue:
		return nil, nil
	case *qdrant.Value_BoolValue:
		return typed.BoolValue, nil
	case *qdrant.Value_IntegerValue:
		return typed.IntegerValue, nil
	case *qdrant.Value_DoubleValue:
		return typed.DoubleValue, nil
	case *qdrant.Value_StringValue:
		return typed.StringValue, nil
	case *qdrant.Value_ListValue:
		values := make([]interface{}, 0, len(typed.ListValue.GetValues()))
		for _, item := range typed.ListValue.GetValues() {
			decoded, err := fromQdrantValue(item)
			if err != nil {
				return nil, err
			}
			values = append(values, decoded)
		}
		return values, nil
	case *qdrant.Value_StructValue:
		fields := make(map[string]interface{}, len(typed.StructValue.GetFields()))
		for key, item := range typed.StructValue.GetFields() {
			decoded, err := fromQdrantValue(item)
			if err != nil {
				return nil, err
			}
			fields[key] = decoded
		}
		return fields, nil
	default:
		return nil, fmt.Errorf("unsupported payload kind %T", value.Kind)
	}
}

func (c *Client) IsNotFound(err error) bool {
	return status.Code(err) == codes.NotFound
}

func withAPIKey(apiKey string) grpc.DialOption {
	return grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if apiKey != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "api-key", apiKey)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	})
}
