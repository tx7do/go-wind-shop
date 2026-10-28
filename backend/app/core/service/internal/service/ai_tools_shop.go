package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// ── 商城域工具：把商城数据暴露给对话的 Function Call 循环 ─────────────
//
// 工具只读：全部走固定 SQL（参数化、LIMIT 钳制），不接受模型拼 SQL。
// 租户边界：tenantId != 0 时强制 tenant_id 过滤。
// 买家边界：query_my_orders 强制 user_id = 当前对话用户。
// 金额口径：DB 以最小货币单位（分）存储，工具输出统一换算为元并标注。
//
// （移植自上游 cms 的 ai_tools_cms.go——cms 工具查询的 posts/post_translations
// 表在 shop 不存在，属死重，已整文件替换为本文件。）

// shopToolDefs 商城域工具清单（与内置工具合并后暴露给模型）。
func shopToolDefs() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "search_products",
				Description: "按关键词搜索在售商品（返回商品ID/名称/简介）。用户问「有什么卖」「找一下××」等时调用。",
				Parameters: json.RawMessage(`{"type":"object","properties":{` +
					`"keyword":{"type":"string","description":"搜索关键词（匹配商品名/简介）"},` +
					`"limit":{"type":"string","description":"可选，返回条数，默认 10，最大 30"}},` +
					`"required":["keyword"]}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_product_detail",
				Description: "查询商品详情：SKU 列表（编码/库存/价格）。用户问某个商品的价格、库存、规格时调用（先用 search_products 拿到商品ID）。",
				Parameters: json.RawMessage(`{"type":"object","properties":{` +
					`"product_id":{"type":"string","description":"商品ID（search_products 返回的 id）"}},` +
					`"required":["product_id"]}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "query_my_orders",
				Description: "查询当前用户自己的订单（状态/金额/下单时间）。用户问「我的订单」「买到的东西」等时调用。金额为实付额。",
				Parameters: json.RawMessage(`{"type":"object","properties":{` +
					`"status":{"type":"string","description":"可选，订单状态过滤：PENDING_PAYMENT/PAID/FULFILLED/CLOSED/CANCELLED"},` +
					`"limit":{"type":"string","description":"可选，返回条数，默认 10，最大 20"}},` +
					`"required":[]}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "query_active_coupons",
				Description: "查询当前可领的优惠券活动（面额/门槛/有效期/剩余量）。用户问「有什么优惠」「优惠券」时调用。",
				Parameters: json.RawMessage(`{"type":"object","properties":{` +
					`"limit":{"type":"string","description":"可选，返回条数，默认 10，最大 20"}},` +
					`"required":[]}`),
			},
		},
	}
}

// execShopTool 执行一个商城工具调用。返回 ok=false 表示非商城工具（回退内置执行器）。
func (s *AiChatService) execShopTool(ctx context.Context, userId, tenantId uint32, name, args string) (string, bool, error) {
	var p map[string]any
	if args != "" {
		_ = json.Unmarshal([]byte(args), &p)
	}
	limit := 10
	if v, ok := p["limit"].(string); ok {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 30 {
			limit = n
		}
	}
	status, _ := p["status"].(string)

	db := s.entClient.DB()
	switch name {
	case "search_products":
		keyword, _ := p["keyword"].(string)
		if keyword == "" {
			return "keyword 不能为空", true, nil
		}
		sqlText := `SELECT p.id, COALESCE(t.name, '') AS name, COALESCE(t.short_description, '') AS short_desc
			FROM mall_products p
			LEFT JOIN mall_product_translations t ON t.product_id = p.id AND t.language_code = 'zh-CN'
			WHERE p.status = 'PRODUCT_STATUS_ACTIVE' AND (t.name ILIKE $1 OR COALESCE(t.short_description, '') ILIKE $1`
		args := []any{"%" + keyword + "%"}
		if tenantId != 0 {
			sqlText += " OR p.tenant_id = $" + strconv.Itoa(len(args)+1)
			args = append(args, tenantId)
		}
		sqlText += ")"
		if tenantId != 0 {
			sqlText += " AND p.tenant_id = $" + strconv.Itoa(len(args)+1)
			args = append(args, tenantId)
		}
		sqlText += " ORDER BY p.id DESC LIMIT $" + strconv.Itoa(len(args)+1)
		args = append(args, limit)
		out, err := s.queryRowsAsString(ctx, db, sqlText, args...)
		if err != nil {
			return "", true, err
		}
		if out == "" {
			return "没有找到相关商品", true, nil
		}
		return out, true, nil

	case "get_product_detail":
		pidStr, _ := p["product_id"].(string)
		pid, perr := strconv.Atoi(pidStr)
		if perr != nil || pid <= 0 {
			return "product_id 非法", true, nil
		}
		out, err := s.queryRowsAsString(ctx, db,
			`SELECT s.id AS sku_id, s.sku_code, s.stock_qty AS stock,
				COALESCE(pr.amount, '0') AS price_fen, ROUND(COALESCE(pr.amount, '0')::numeric/100, 2) AS price_yuan
			FROM mall_skus s
			LEFT JOIN mall_sku_prices pr ON pr.sku_id = s.id AND pr.currency = 'CNY'
			WHERE s.product_id = $1`, pid)
		if err != nil {
			return "", true, err
		}
		if out == "" {
			return "商品不存在或没有可用规格", true, nil
		}
		return out, true, nil

	case "query_my_orders":
		sqlText := `SELECT id, status, total_amount AS total_fen,
			ROUND(total_amount::numeric/100, 2) AS total_yuan, currency,
			TO_CHAR(created_at, 'YYYY-MM-DD HH24:MI') AS created_at
			FROM mall_orders WHERE user_id = $1`
		args := []any{userId}
		if status != "" {
			sqlText += " AND status = $" + strconv.Itoa(len(args)+1)
			args = append(args, status)
		}
		if tenantId != 0 {
			sqlText += " AND tenant_id = $" + strconv.Itoa(len(args)+1)
			args = append(args, tenantId)
		}
		if v, ok := p["limit"].(string); ok {
			if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 20 {
				limit = n
			}
		}
		sqlText += " ORDER BY id DESC LIMIT $" + strconv.Itoa(len(args)+1)
		args = append(args, limit)
		out, err := s.queryRowsAsString(ctx, db, sqlText, args...)
		if err != nil {
			return "", true, err
		}
		if out == "" {
			return "没有符合条件的订单", true, nil
		}
		return out, true, nil

	case "query_active_coupons":
		if v, ok := p["limit"].(string); ok {
			if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 20 {
				limit = n
			}
		}
		sqlText := `SELECT id, discount_type, discount_value AS value_fen, discount_percentage,
			ROUND(discount_value::numeric/100, 2) AS value_yuan, currency,
			TO_CHAR(valid_from, 'YYYY-MM-DD') AS valid_from,
			TO_CHAR(valid_until, 'YYYY-MM-DD') AS valid_until,
			GREATEST(COALESCE(max_redemptions,0) - COALESCE(redeemed_count,0), 0) AS remaining
			FROM mall_coupon_templates
			WHERE status = 'ACTIVE' AND claimable = true
			AND valid_from <= now() AND valid_until >= now()`
		args := []any{}
		if tenantId != 0 {
			sqlText += " AND tenant_id = $1"
			args = append(args, tenantId)
		}
		sqlText += " ORDER BY id DESC LIMIT $" + strconv.Itoa(len(args)+1)
		args = append(args, limit)
		out, err := s.queryRowsAsString(ctx, db, sqlText, args...)
		if err != nil {
			return "", true, err
		}
		if out == "" {
			return "当前没有可领取的优惠券", true, nil
		}
		return out, true, nil

	default:
		return "", false, fmt.Errorf("unknown shop tool: %s", name)
	}
}

// queryRowsAsString 执行查询并把结果集渲染成 LLM 可读的竖线分隔文本。
// （自上游 cms 的 ai_tools_cms.go 迁入——cms 工具已随 posts 表裁剪移除。）
func (s *AiChatService) queryRowsAsString(ctx context.Context, db *sql.DB, query string, args ...any) (string, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return "", err
	}
	defer func() { _ = rows.Close() }()

	columns, err := rows.Columns()
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(strings.Join(columns, " | ") + "\n")
	count := 0
	for rows.Next() {
		values := make([]interface{}, len(columns))
		ptrs := make([]interface{}, len(columns))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err = rows.Scan(ptrs...); err != nil {
			return "", err
		}
		cells := make([]string, len(columns))
		for i, v := range values {
			cells[i] = stringifyCellValue(v)
		}
		sb.WriteString(strings.Join(cells, " | ") + "\n")
		count++
		if count >= 50 {
			break
		}
	}
	if err = rows.Err(); err != nil {
		return "", err
	}
	return strings.TrimRight(sb.String(), "\n"), nil
}
