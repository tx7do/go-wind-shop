package gateway

import (
	"context"
	"errors"
	"net/http"
)

// MockGateway 模拟渠道：不出真实资金、无服务端回调。
// C 端通过 mock-confirm 接口（登录态 + 仅限 MOCK 流水）完成确认，
// 全链路可本地演示支付闭环；生产部署应保持 MOCK 渠道停用。
type MockGateway struct{}

func NewMockGateway() *MockGateway {
	return &MockGateway{}
}

func (g *MockGateway) CreatePayment(ctx context.Context, req PayRequest) (*PayResult, error) {
	// 模拟渠道无需拉起外部支付：C 端展示「模拟支付」按钮直接确认
	return &PayResult{}, nil
}

func (g *MockGateway) ParseNotify(ctx context.Context, r *http.Request) (*NotifyResult, error) {
	return nil, errors.New("mock gateway has no server-side notify")
}

func (g *MockGateway) CloseOrder(ctx context.Context, outTradeNo string) error {
	// 模拟渠道无网关预支付单可关：恒成功（调用方联动路径因此无感）
	return nil
}

func (g *MockGateway) Refund(ctx context.Context, req RefundRequest) error {
	// 模拟渠道不出真实资金：恒成功，让退款闭环可本地演示
	return nil
}

func (g *MockGateway) QueryOrder(ctx context.Context, outTradeNo string) (*OrderQueryResult, error) {
	// 模拟渠道无网关侧状态可查：nil 表示跳过对账
	return nil, nil
}
