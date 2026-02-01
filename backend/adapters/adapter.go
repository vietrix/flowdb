package adapters

import (
	"context"
	"time"
)

type Namespace struct {
	Name string `json:"name"`
}

type Entity struct {
	Name string `json:"name"`
}

type Column struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type EntityInfo struct {
	Columns []Column          `json:"columns"`
	Indexes []string          `json:"indexes"`
	Stats   map[string]any    `json:"stats"`
	Meta    map[string]string `json:"meta"`
}

type BrowseOptions struct {
	Page     int
	PageSize int
	Sort     string
	Filter   string
}

type QueryOptions struct {
	MaxRows int
	Timeout time.Duration
}

type ResultStream struct {
	Columns []Column
	Fields  []string
	Rows    <-chan []any
	Docs    <-chan map[string]any
	Err     <-chan error
	Done    <-chan struct{}
}

type Adapter interface {
	ListNamespaces(ctx context.Context) ([]Namespace, error)
	ListEntities(ctx context.Context, ns string) ([]Entity, error)
	GetEntityInfo(ctx context.Context, ns string, name string) (EntityInfo, error)
	Browse(ctx context.Context, ns string, name string, opts BrowseOptions) (*ResultStream, error)
	Query(ctx context.Context, statement string, opts QueryOptions) (*ResultStream, error)
	Explain(ctx context.Context, statement string) (any, error)
	Close() error
}
