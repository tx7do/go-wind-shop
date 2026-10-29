// Package payment 支付域共享类型：渠道凭据的强类型与校验。
//
// 合规约束（代码交付物定位）：仅支持「部署方自带资质」的直连商户模式——
// 交易资金由支付渠道直接清算给部署方自己的商户号，本系统不归集、不经手、
// 不二次分配任何交易资金。凭据由部署方在管理端录入，服务端 AES-GCM 加密
// 落库；代码与仓库内不含任何真实密钥。
package payment

import (
	"encoding/json"
	"fmt"
)

// WechatCredentials 微信支付 Native（直连商户）凭据。
// 取自微信商户平台「API安全」：APIv3 密钥、商户证书序列号、商户私钥（apiclient_key.pem）。
type WechatCredentials struct {
	MchCertSerialNo string `json:"mchCertSerialNo"`
	APIv3Key        string `json:"apiV3Key"`
	PrivateKeyPEM   string `json:"privateKeyPem"`
}

// AlipayCredentials 支付宝电脑网站支付（直连商户）凭据。
// 取自支付宝开放平台：应用私钥、支付宝公钥（注意不是应用公钥）。
type AlipayCredentials struct {
	PrivateKey      string `json:"privateKey"`
	AlipayPublicKey string `json:"alipayPublicKey"`
}

// ParseWechatCredentials 从渠道凭据 JSON 解析微信凭据并校验非空。
func ParseWechatCredentials(raw string) (*WechatCredentials, error) {
	var c WechatCredentials
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		return nil, fmt.Errorf("invalid wechat credentials json: %w", err)
	}
	if c.MchCertSerialNo == "" || c.APIv3Key == "" || c.PrivateKeyPEM == "" {
		return nil, fmt.Errorf("wechat credentials require mchCertSerialNo, apiV3Key, privateKeyPem")
	}
	return &c, nil
}

// ParseAlipayCredentials 从渠道凭据 JSON 解析支付宝凭据并校验非空。
func ParseAlipayCredentials(raw string) (*AlipayCredentials, error) {
	var c AlipayCredentials
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		return nil, fmt.Errorf("invalid alipay credentials json: %w", err)
	}
	if c.PrivateKey == "" || c.AlipayPublicKey == "" {
		return nil, fmt.Errorf("alipay credentials require privateKey, alipayPublicKey")
	}
	return &c, nil
}
