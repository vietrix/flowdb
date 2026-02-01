package mongodb

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"flowdb/backend/adapters"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	URI      string
	Database string
}

type Adapter struct {
	client *mongo.Client
	db     *mongo.Database
}

type dslQuery struct {
	Action     string         `json:"action"`
	Collection string         `json:"collection"`
	Filter     map[string]any `json:"filter"`
	Pipeline   []any          `json:"pipeline"`
	Document   any            `json:"document"`
	Update     any            `json:"update"`
	Options    map[string]any `json:"options"`
}

func New(ctx context.Context, cfg Config) (*Adapter, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, err
	}
	db := client.Database(cfg.Database)
	return &Adapter{client: client, db: db}, nil
}

func (a *Adapter) Close() error {
	if a.client != nil {
		return a.client.Disconnect(context.Background())
	}
	return nil
}

func (a *Adapter) ListNamespaces(ctx context.Context) ([]adapters.Namespace, error) {
	names, err := a.client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	out := make([]adapters.Namespace, 0, len(names))
	for _, name := range names {
		out = append(out, adapters.Namespace{Name: name})
	}
	return out, nil
}

func (a *Adapter) ListEntities(ctx context.Context, ns string) ([]adapters.Entity, error) {
	db := a.client.Database(ns)
	names, err := db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	out := make([]adapters.Entity, 0, len(names))
	for _, name := range names {
		out = append(out, adapters.Entity{Name: name})
	}
	return out, nil
}

func (a *Adapter) GetEntityInfo(ctx context.Context, ns string, name string) (adapters.EntityInfo, error) {
	db := a.client.Database(ns)
	collection := db.Collection(name)
	indexes, err := collection.Indexes().List(ctx)
	if err != nil {
		return adapters.EntityInfo{}, err
	}
	var idxNames []string
	for indexes.Next(ctx) {
		var doc bson.M
		if err := indexes.Decode(&doc); err != nil {
			return adapters.EntityInfo{}, err
		}
		if name, ok := doc["name"].(string); ok {
			idxNames = append(idxNames, name)
		}
	}
	stats := bson.M{}
	_ = db.RunCommand(ctx, bson.D{{Key: "collStats", Value: name}}).Decode(&stats)
	return adapters.EntityInfo{
		Indexes: idxNames,
		Stats:   stats,
	}, nil
}

func (a *Adapter) Browse(ctx context.Context, ns string, name string, opts adapters.BrowseOptions) (*adapters.ResultStream, error) {
	db := a.client.Database(ns)
	collection := db.Collection(name)
	filter := bson.M{}
	if opts.Filter != "" {
		if err := json.Unmarshal([]byte(opts.Filter), &filter); err != nil {
			return nil, err
		}
	}
	findOpts := options.Find()
	pageSize := opts.PageSize
	if pageSize <= 0 || pageSize > 1000 {
		pageSize = 100
	}
	page := opts.Page
	if page < 1 {
		page = 1
	}
	findOpts.SetLimit(int64(pageSize))
	findOpts.SetSkip(int64((page - 1) * pageSize))
	cursor, err := collection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	return streamCursor(ctx, cursor)
}

func (a *Adapter) Query(ctx context.Context, statement string, opts adapters.QueryOptions) (*adapters.ResultStream, error) {
	var q dslQuery
	if err := json.Unmarshal([]byte(statement), &q); err != nil {
		return nil, err
	}
	if q.Collection == "" {
		return nil, errors.New("collection required")
	}
	ctx = withTimeout(ctx, opts.Timeout)
	coll := a.db.Collection(q.Collection)
	switch q.Action {
	case "find", "":
		filter := bson.M(q.Filter)
		findOpts := options.Find()
		if q.Options != nil {
			if limit, ok := toInt64(q.Options["limit"]); ok {
				findOpts.SetLimit(limit)
			}
			if sort, ok := q.Options["sort"].(map[string]any); ok {
				findOpts.SetSort(sort)
			}
			if proj, ok := q.Options["projection"].(map[string]any); ok {
				findOpts.SetProjection(proj)
			}
		}
		cursor, err := coll.Find(ctx, filter, findOpts)
		if err != nil {
			return nil, err
		}
		return streamCursor(ctx, cursor)
	case "aggregate":
		pipeline := bson.A{}
		for _, stage := range q.Pipeline {
			pipeline = append(pipeline, stage)
		}
		cursor, err := coll.Aggregate(ctx, pipeline)
		if err != nil {
			return nil, err
		}
		return streamCursor(ctx, cursor)
	case "insert":
		switch doc := q.Document.(type) {
		case []any:
			_, err := coll.InsertMany(ctx, doc)
			if err != nil {
				return nil, err
			}
		default:
			_, err := coll.InsertOne(ctx, doc)
			if err != nil {
				return nil, err
			}
		}
		return emptyStream(), nil
	case "update":
		filter := bson.M(q.Filter)
		update := q.Update
		multi := false
		if q.Options != nil {
			if m, ok := q.Options["multi"].(bool); ok {
				multi = m
			}
		}
		if multi {
			_, err := coll.UpdateMany(ctx, filter, update)
			if err != nil {
				return nil, err
			}
		} else {
			_, err := coll.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}
		}
		return emptyStream(), nil
	case "delete":
		filter := bson.M(q.Filter)
		multi := false
		if q.Options != nil {
			if m, ok := q.Options["multi"].(bool); ok {
				multi = m
			}
		}
		if multi {
			_, err := coll.DeleteMany(ctx, filter)
			if err != nil {
				return nil, err
			}
		} else {
			_, err := coll.DeleteOne(ctx, filter)
			if err != nil {
				return nil, err
			}
		}
		return emptyStream(), nil
	default:
		return nil, errors.New("unsupported action")
	}
}

func (a *Adapter) Explain(ctx context.Context, statement string) (any, error) {
	var q dslQuery
	if err := json.Unmarshal([]byte(statement), &q); err != nil {
		return nil, err
	}
	if q.Collection == "" {
		return nil, errors.New("collection required")
	}
	switch q.Action {
	case "find", "":
		cmd := bson.D{
			{Key: "explain", Value: bson.D{
				{Key: "find", Value: q.Collection},
				{Key: "filter", Value: q.Filter},
			}},
		}
		var out bson.M
		if err := a.db.RunCommand(ctx, cmd).Decode(&out); err != nil {
			return nil, err
		}
		return out, nil
	case "aggregate":
		cmd := bson.D{
			{Key: "explain", Value: bson.D{
				{Key: "aggregate", Value: q.Collection},
				{Key: "pipeline", Value: q.Pipeline},
				{Key: "cursor", Value: bson.M{}},
			}},
		}
		var out bson.M
		if err := a.db.RunCommand(ctx, cmd).Decode(&out); err != nil {
			return nil, err
		}
		return out, nil
	default:
		return nil, errors.New("explain only supports read operations")
	}
}

func streamCursor(ctx context.Context, cursor *mongo.Cursor) (*adapters.ResultStream, error) {
	docChan := make(chan map[string]any, 64)
	errChan := make(chan error, 1)
	done := make(chan struct{})
	go func() {
		defer cursor.Close(ctx)
		defer close(docChan)
		defer close(done)
		for cursor.Next(ctx) {
			var doc map[string]any
			if err := cursor.Decode(&doc); err != nil {
				errChan <- err
				return
			}
			docChan <- doc
		}
		if cursor.Err() != nil {
			errChan <- cursor.Err()
		}
	}()
	return &adapters.ResultStream{
		Docs: docChan,
		Err:  errChan,
		Done: done,
	}, nil
}

func emptyStream() *adapters.ResultStream {
	done := make(chan struct{})
	close(done)
	return &adapters.ResultStream{
		Rows: make(chan []any),
		Docs: make(chan map[string]any),
		Err:  make(chan error, 1),
		Done: done,
	}
}

func withTimeout(ctx context.Context, timeout time.Duration) context.Context {
	if timeout <= 0 {
		return ctx
	}
	ctx, _ = context.WithTimeout(ctx, timeout)
	return ctx
}

func toInt64(v any) (int64, bool) {
	switch t := v.(type) {
	case float64:
		return int64(t), true
	case int64:
		return t, true
	case int:
		return int64(t), true
	default:
		return 0, false
	}
}
