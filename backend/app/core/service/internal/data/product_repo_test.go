package data

import (
	"encoding/json"
	"testing"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
)

// newQueryReq 构造一个 query 为指定 JSON 的分页请求。
func newQueryReq(t *testing.T, m map[string]any) *paginationV1.PagingRequest {
	t.Helper()
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal query: %v", err)
	}
	return &paginationV1.PagingRequest{
		FilteringType: &paginationV1.PagingRequest_Query{Query: string(raw)},
	}
}

func TestExtractAndStripNameKeyword_FromName(t *testing.T) {
	req := newQueryReq(t, map[string]any{"name": "手机"})
	got := extractAndStripNameKeyword(req)
	if got != "手机" {
		t.Fatalf("expected keyword 手机, got %q", got)
	}
	// name 应被剥离，剩余字段为空 → FilteringType 置空
	if req.FilteringType != nil {
		t.Fatalf("expected FilteringType nil after strip, got %v", req.FilteringType)
	}
}

func TestExtractAndStripNameKeyword_FromKeywordAndQ(t *testing.T) {
	for _, key := range []string{"keyword", "q"} {
		req := newQueryReq(t, map[string]any{key: " phone "})
		got := extractAndStripNameKeyword(req)
		if got != "phone" {
			t.Fatalf("key=%s: expected trimmed 'phone', got %q", key, got)
		}
	}
}

func TestExtractAndStripNameKeyword_KeepsOtherFields(t *testing.T) {
	req := newQueryReq(t, map[string]any{"name": "phone", "categoryId": float64(7), "brandId": float64(3)})
	got := extractAndStripNameKeyword(req)
	if got != "phone" {
		t.Fatalf("expected phone, got %q", got)
	}
	// 剩余字段应写回 query，且不再包含 name
	q := req.GetQuery()
	if q == "" {
		t.Fatalf("expected remaining query, got empty")
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(q), &m); err != nil {
		t.Fatalf("unmarshal remaining query: %v", err)
	}
	if _, ok := m["name"]; ok {
		t.Fatalf("name should be stripped, query=%s", q)
	}
	if m["categoryId"] != float64(7) || m["brandId"] != float64(3) {
		t.Fatalf("other fields lost, query=%s", q)
	}
}

func TestExtractAndStripNameKeyword_NoQuery(t *testing.T) {
	req := &paginationV1.PagingRequest{}
	if got := extractAndStripNameKeyword(req); got != "" {
		t.Fatalf("expected empty for no query, got %q", got)
	}
}

func TestExtractAndStripNameKeyword_EmptyValueIgnored(t *testing.T) {
	req := newQueryReq(t, map[string]any{"name": "  ", "categoryId": float64(7)})
	got := extractAndStripNameKeyword(req)
	if got != "" {
		t.Fatalf("expected empty keyword for blank value, got %q", got)
	}
	// 空 name 不应触发剥离/置空，剩余字段保留
	if req.GetQuery() == "" {
		t.Fatalf("query should be preserved when name is blank")
	}
}

// 非字符串 name 值不应被剥离，保留原样交 DSL 处理，避免污染剩余 query。
func TestExtractAndStripNameKeyword_NonStringValuePreserved(t *testing.T) {
	req := newQueryReq(t, map[string]any{"name": float64(123), "categoryId": float64(7)})
	got := extractAndStripNameKeyword(req)
	if got != "" {
		t.Fatalf("expected empty keyword for non-string value, got %q", got)
	}
	// name(123) 不应被删除，query 应原样保留两个字段
	var m map[string]any
	if err := json.Unmarshal([]byte(req.GetQuery()), &m); err != nil {
		t.Fatalf("unmarshal query: %v", err)
	}
	if _, ok := m["name"]; !ok {
		t.Fatalf("non-string name should be preserved, not deleted; query=%s", req.GetQuery())
	}
	if m["categoryId"] != float64(7) {
		t.Fatalf("categoryId should be preserved, query=%s", req.GetQuery())
	}
}

// 多个候选 key 同时存在：最后一个（按 [name,keyword,q] 顺序即 q）胜出，
// 三个 key 都应被剥离。
func TestExtractAndStripNameKeyword_MultipleKeysLastWins(t *testing.T) {
	req := newQueryReq(t, map[string]any{"name": "a", "keyword": "b", "q": "c"})
	got := extractAndStripNameKeyword(req)
	if got != "c" {
		t.Fatalf("expected q value 'c' to win, got %q", got)
	}
	// 三个 key 都应被剥离 → query 为空 → FilteringType 置空
	if req.FilteringType != nil {
		t.Fatalf("expected FilteringType nil after stripping all keys, got %v", req.FilteringType)
	}
}
