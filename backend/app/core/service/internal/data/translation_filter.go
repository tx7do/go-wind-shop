package data

import (
	"encoding/json"
	"strconv"
	"strings"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
)

// ExtractTranslationFilters 从分页 query JSON 中抽取翻译表字段的条件，
// 并把这些键从 req.Query 原地移除（避免通用分页把它们当主表列拼进 SQL——
// 品牌/类目/属性/商品等实体的名称只存在于 *_translations 表）。
//
// 支持的键形如 "name"、"name__contains"、"name__equal"；其它操作符原样保留，
// 交由通用分页过滤处理。返回值：contains/equal 分别为 字段->值 映射。
func ExtractTranslationFilters(req *paginationV1.PagingRequest, fields ...string) (contains, equals map[string]string, err error) {
	contains = map[string]string{}
	equals = map[string]string{}
	if req == nil || req.GetQuery() == "" {
		return contains, equals, nil
	}

	var cond map[string]any
	if err = json.Unmarshal([]byte(req.GetQuery()), &cond); err != nil {
		return nil, nil, err
	}

	fieldSet := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		fieldSet[f] = struct{}{}
	}

	changed := false
	for k, v := range cond {
		base, op := k, ""
		if idx := strings.Index(k, "__"); idx >= 0 {
			base, op = k[:idx], k[idx+2:]
		}
		if _, ok := fieldSet[base]; !ok {
			continue
		}

		sv := ""
		switch tv := v.(type) {
		case string:
			sv = tv
		case float64:
			sv = strconv.FormatFloat(tv, 'f', -1, 64)
		default:
			continue
		}

		switch op {
		case "contains":
			contains[base] = sv
		case "", "equal", "eq":
			equals[base] = sv
		default:
			continue
		}
		delete(cond, k)
		changed = true
	}

	if changed {
		b, mErr := json.Marshal(cond)
		if mErr != nil {
			return nil, nil, mErr
		}
		if len(cond) == 0 {
			// 全部条件被剥离：清空 oneof 过滤，避免残留空 query。
			req.FilteringType = nil
		} else {
			req.FilteringType = &paginationV1.PagingRequest_Query{Query: string(b)}
		}
	}

	return contains, equals, nil
}

// stringIDsToUint32 把 ent Select().Strings(ctx) 查到的外键列转为 uint32。
func stringIDsToUint32(rows []string) []uint32 {
	ids := make([]uint32, 0, len(rows))
	for _, s := range rows {
		if n, err := strconv.ParseUint(s, 10, 32); err == nil {
			ids = append(ids, uint32(n))
		}
	}
	return ids
}
