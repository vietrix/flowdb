package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"flowdb/backend/adapters"

	"github.com/jackc/pgx/v5"
)

type Config struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
	SSLMode  string
}

type Adapter struct {
	conn *pgx.Conn
}

func New(ctx context.Context, cfg Config) (*Adapter, error) {
	if cfg.SSLMode == "" {
		cfg.SSLMode = "prefer"
	}
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return nil, err
	}
	return &Adapter{conn: conn}, nil
}

func (a *Adapter) Close() error {
	if a.conn != nil {
		return a.conn.Close(context.Background())
	}
	return nil
}

func (a *Adapter) ListNamespaces(ctx context.Context) ([]adapters.Namespace, error) {
	rows, err := a.conn.Query(ctx, `SELECT schema_name FROM information_schema.schemata ORDER BY schema_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []adapters.Namespace
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		list = append(list, adapters.Namespace{Name: name})
	}
	return list, rows.Err()
}

func (a *Adapter) ListEntities(ctx context.Context, ns string) ([]adapters.Entity, error) {
	rows, err := a.conn.Query(ctx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema=$1 AND table_type='BASE TABLE'
		ORDER BY table_name
	`, ns)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []adapters.Entity
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		list = append(list, adapters.Entity{Name: name})
	}
	return list, rows.Err()
}

func (a *Adapter) GetEntityInfo(ctx context.Context, ns string, name string) (adapters.EntityInfo, error) {
	info := adapters.EntityInfo{}
	rows, err := a.conn.Query(ctx, `
		SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_schema=$1 AND table_name=$2
		ORDER BY ordinal_position
	`, ns, name)
	if err != nil {
		return info, err
	}
	defer rows.Close()
	for rows.Next() {
		var col adapters.Column
		if err := rows.Scan(&col.Name, &col.Type); err != nil {
			return info, err
		}
		info.Columns = append(info.Columns, col)
	}
	idxRows, err := a.conn.Query(ctx, `
		SELECT indexname FROM pg_indexes WHERE schemaname=$1 AND tablename=$2 ORDER BY indexname
	`, ns, name)
	if err == nil {
		defer idxRows.Close()
		for idxRows.Next() {
			var idx string
			if err := idxRows.Scan(&idx); err != nil {
				return info, err
			}
			info.Indexes = append(info.Indexes, idx)
		}
	}
	return info, nil
}

func (a *Adapter) Browse(ctx context.Context, ns string, name string, opts adapters.BrowseOptions) (*adapters.ResultStream, error) {
	page := opts.Page
	if page < 1 {
		page = 1
	}
	pageSize := opts.PageSize
	if pageSize <= 0 || pageSize > 1000 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize
	sortClause := ""
	if safeIdent(opts.Sort) {
		sortClause = fmt.Sprintf(" ORDER BY %s", pqQuoteIdent(opts.Sort))
	}
	filterClause := ""
	if safeFilter(opts.Filter) {
		filterClause = fmt.Sprintf(" WHERE %s", opts.Filter)
	}
	stmt := fmt.Sprintf("SELECT * FROM %s.%s%s%s LIMIT %d OFFSET %d",
		pqQuoteIdent(ns), pqQuoteIdent(name), filterClause, sortClause, pageSize, offset)
	return a.Query(ctx, stmt, adapters.QueryOptions{})
}

func (a *Adapter) Query(ctx context.Context, statement string, opts adapters.QueryOptions) (*adapters.ResultStream, error) {
	ctx = withTimeout(ctx, opts.Timeout)
	if isWrite(statement) {
		_, err := a.conn.Exec(ctx, statement)
		if err != nil {
			return nil, err
		}
		done := make(chan struct{})
		close(done)
		return &adapters.ResultStream{
			Columns: nil,
			Rows:    make(chan []any),
			Err:     make(chan error, 1),
			Done:    done,
		}, nil
	}
	rows, err := a.conn.Query(ctx, statement)
	if err != nil {
		return nil, err
	}
	fds := rows.FieldDescriptions()
	cols := make([]adapters.Column, len(fds))
	for i, fd := range fds {
		cols[i] = adapters.Column{Name: string(fd.Name), Type: fmt.Sprintf("%d", fd.DataTypeOID)}
	}
	rowChan := make(chan []any, 64)
	errChan := make(chan error, 1)
	done := make(chan struct{})
	go func() {
		defer rows.Close()
		defer close(rowChan)
		defer close(done)
		for rows.Next() {
			values, err := rows.Values()
			if err != nil {
				errChan <- err
				return
			}
			rowChan <- values
		}
		if rows.Err() != nil {
			errChan <- rows.Err()
		}
	}()
	return &adapters.ResultStream{
		Columns: cols,
		Rows:    rowChan,
		Err:     errChan,
		Done:    done,
	}, nil
}

func (a *Adapter) Explain(ctx context.Context, statement string) (any, error) {
	ctx = withTimeout(ctx, 10*time.Second)
	rows, err := a.conn.Query(ctx, `EXPLAIN (FORMAT JSON) `+statement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		var data any
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		return data, nil
	}
	return nil, errors.New("no explain output")
}

func isWrite(statement string) bool {
	stmt := strings.TrimSpace(strings.ToLower(statement))
	return strings.HasPrefix(stmt, "insert") ||
		strings.HasPrefix(stmt, "update") ||
		strings.HasPrefix(stmt, "delete") ||
		strings.HasPrefix(stmt, "create") ||
		strings.HasPrefix(stmt, "alter") ||
		strings.HasPrefix(stmt, "drop")
}

func safeIdent(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if !(r == '_' || r == '.' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}

func safeFilter(value string) bool {
	if value == "" {
		return false
	}
	if strings.Contains(value, ";") || strings.Contains(value, "--") {
		return false
	}
	return true
}

func pqQuoteIdent(value string) string {
	parts := strings.Split(value, ".")
	for i, p := range parts {
		parts[i] = `"` + strings.ReplaceAll(p, `"`, `""`) + `"`
	}
	return strings.Join(parts, ".")
}

func withTimeout(ctx context.Context, timeout time.Duration) context.Context {
	if timeout <= 0 {
		return ctx
	}
	ctx, _ = context.WithTimeout(ctx, timeout)
	return ctx
}
