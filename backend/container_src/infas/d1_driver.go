package infas

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type d1Query struct {
	SQL    string `json:"sql"`
	Params []any  `json:"params,omitempty"`
}

type d1ExecRequest struct {
	SQL    string `json:"sql"`
	Params []any  `json:"params,omitempty"`
}

type d1QueryRequest struct {
	SQL    string `json:"sql"`
	Params []any  `json:"params,omitempty"`
	Raw    bool   `json:"raw"`
}

type d1BatchRequest struct {
	Batch []d1Query `json:"batch"`
}

type d1ExecResponse struct {
	Success bool           `json:"success"`
	Meta    map[string]any `json:"meta"`
	Error   string         `json:"error"`
}

type d1RawResponse struct {
	Success bool     `json:"success"`
	Columns []string `json:"columns"`
	Rows    [][]any  `json:"rows"`
	Error   string   `json:"error"`
}

type d1BatchResponse struct {
	Success bool             `json:"success"`
	Results []d1ExecResponse `json:"results"`
	Error   string           `json:"error"`
}

type d1Driver struct{}

func init() {
	sql.Register("d1", d1Driver{})
}

func (d1Driver) Open(string) (driver.Conn, error) {
	client, err := newD1ClientFromEnv()
	if err != nil {
		return nil, err
	}
	return &d1Conn{client: client}, nil
}

type d1Client struct {
	endpoint string
	token    string
	http     *http.Client
}

func newD1ClientFromEnv() (*d1Client, error) {
	endpoint := strings.TrimSpace(os.Getenv("D1_PROXY_URL"))
	if endpoint == "" {
		endpoint = "https://mocatestdeploy.duylongart.workers.dev"
	}
	token := strings.TrimSpace(os.Getenv("D1_PROXY_TOKEN"))
	if token == "" {
		return nil, fmt.Errorf("D1_PROXY_TOKEN is required")
	}
	return &d1Client{
		endpoint: strings.TrimRight(endpoint, "/") + "/internal/d1/query",
		token:    token,
		http:     &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (client *d1Client) post(ctx context.Context, payload any, target any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		client.endpoint,
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	request.Header.Set("content-type", "application/json")
	request.Header.Set("x-d1-proxy-token", client.token)
	response, err := client.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("d1 proxy returned %d: %s", response.StatusCode, strings.TrimSpace(string(contents)))
	}
	if err := json.Unmarshal(contents, target); err != nil {
		return fmt.Errorf("decode d1 response: %w", err)
	}
	return nil
}

type d1Conn struct {
	client *d1Client
	tx     *d1Tx
}

func (conn *d1Conn) Prepare(query string) (driver.Stmt, error) {
	return &d1Stmt{conn: conn, query: query}, nil
}

func (conn *d1Conn) Close() error {
	return nil
}

func (conn *d1Conn) Begin() (driver.Tx, error) {
	return &d1Tx{conn: conn, queries: []d1Query{}}, nil
}

func (conn *d1Conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	statement := d1Query{SQL: query, Params: d1Params(args)}
	if conn.tx != nil {
		conn.tx.queries = append(conn.tx.queries, statement)
		return d1Result{}, nil
	}
	var response d1ExecResponse
	if err := conn.client.post(ctx, d1ExecRequest(statement), &response); err != nil {
		return nil, err
	}
	if !response.Success {
		return nil, fmt.Errorf("d1 exec failed: %s", response.Error)
	}
	return d1Result{meta: response.Meta}, nil
}

func (conn *d1Conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	var response d1RawResponse
	payload := d1QueryRequest{SQL: query, Params: d1Params(args), Raw: true}
	if err := conn.client.post(ctx, payload, &response); err != nil {
		return nil, err
	}
	if !response.Success {
		return nil, fmt.Errorf("d1 query failed: %s", response.Error)
	}
	return &d1Rows{columns: response.Columns, rows: response.Rows}, nil
}

type d1Stmt struct {
	conn  *d1Conn
	query string
}

func (stmt *d1Stmt) Close() error {
	return nil
}

func (stmt *d1Stmt) NumInput() int {
	return -1
}

func (stmt *d1Stmt) Exec(args []driver.Value) (driver.Result, error) {
	return stmt.ExecContext(context.Background(), valuesToNamed(args))
}

func (stmt *d1Stmt) Query(args []driver.Value) (driver.Rows, error) {
	return stmt.QueryContext(context.Background(), valuesToNamed(args))
}

func (stmt *d1Stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	return stmt.conn.ExecContext(ctx, stmt.query, args)
}

func (stmt *d1Stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	return stmt.conn.QueryContext(ctx, stmt.query, args)
}

func valuesToNamed(values []driver.Value) []driver.NamedValue {
	named := make([]driver.NamedValue, 0, len(values))
	for _, value := range values {
		named = append(named, driver.NamedValue{Value: value})
	}
	return named
}

type d1Tx struct {
	conn    *d1Conn
	queries []d1Query
	done    bool
}

func (tx *d1Tx) Commit() error {
	if tx.done {
		return nil
	}
	tx.done = true
	tx.conn.tx = nil
	if len(tx.queries) == 0 {
		return nil
	}
	var response d1BatchResponse
	if err := tx.conn.client.post(context.Background(), d1BatchRequest{Batch: tx.queries}, &response); err != nil {
		return err
	}
	if !response.Success {
		return fmt.Errorf("d1 batch failed: %s", response.Error)
	}
	for _, result := range response.Results {
		if !result.Success {
			return fmt.Errorf("d1 batch statement failed: %s", result.Error)
		}
	}
	return nil
}

func (tx *d1Tx) Rollback() error {
	tx.done = true
	tx.conn.tx = nil
	return nil
}

type d1Result struct {
	meta map[string]any
}

func (result d1Result) LastInsertId() (int64, error) {
	return d1Int64(result.meta["last_row_id"]), nil
}

func (result d1Result) RowsAffected() (int64, error) {
	if changes, ok := result.meta["changes"]; ok {
		return d1Int64(changes), nil
	}
	return d1Int64(result.meta["rows_written"]), nil
}

type d1Rows struct {
	columns []string
	rows    [][]any
	index   int
}

func (rows *d1Rows) Columns() []string {
	return rows.columns
}

func (rows *d1Rows) Close() error {
	return nil
}

func (rows *d1Rows) Next(dest []driver.Value) error {
	if rows.index >= len(rows.rows) {
		return io.EOF
	}
	row := rows.rows[rows.index]
	for index := range rows.columns {
		if index >= len(row) || index >= len(dest) {
			break
		}
		dest[index] = d1Value(row[index])
	}
	rows.index++
	return nil
}

func d1Params(args []driver.NamedValue) []any {
	params := make([]any, 0, len(args))
	for _, arg := range args {
		params = append(params, d1Param(arg.Value))
	}
	return params
}

func d1Param(value any) any {
	switch typed := value.(type) {
	case nil:
		return nil
	case []byte:
		return base64.StdEncoding.EncodeToString(typed)
	case bool:
		if typed {
			return 1
		}
		return 0
	case time.Time:
		return typed.Unix()
	case int:
		return int64(typed)
	case int8:
		return int64(typed)
	case int16:
		return int64(typed)
	case int32:
		return int64(typed)
	case uint:
		return int64(typed)
	case uint8:
		return int64(typed)
	case uint16:
		return int64(typed)
	case uint32:
		return int64(typed)
	case uint64:
		return int64(typed)
	case float32:
		return float64(typed)
	default:
		return value
	}
}

func d1Value(value any) driver.Value {
	switch typed := value.(type) {
	case nil:
		return nil
	case string:
		return typed
	case bool:
		if typed {
			return int64(1)
		}
		return int64(0)
	case float64:
		if typed == float64(int64(typed)) {
			return int64(typed)
		}
		return typed
	case []any, map[string]any:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprint(typed)
		}
		return string(encoded)
	default:
		return fmt.Sprint(typed)
	}
}

func d1Int64(value any) int64 {
	switch typed := value.(type) {
	case int64:
		return typed
	case int:
		return int64(typed)
	case float64:
		return int64(typed)
	case json.Number:
		parsed, err := typed.Int64()
		if err == nil {
			return parsed
		}
	}
	return 0
}
