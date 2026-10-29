package gateway

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"testing"
)

// TestCentsYuanRoundTrip 分↔元字符串互转口径。
func TestCentsYuanRoundTrip(t *testing.T) {
	cases := []struct {
		cents int64
		yuan  string
	}{
		{12800, "128.00"},
		{9900, "99.00"},
		{1, "0.01"},
		{123456789, "1234567.89"},
	}
	for _, c := range cases {
		if got := CentsToYuanString(c.cents); got != c.yuan {
			t.Errorf("CentsToYuanString(%d) = %q, want %q", c.cents, got, c.yuan)
		}
		back, err := YuanStringToCents(c.yuan)
		if err != nil {
			t.Fatalf("YuanStringToCents(%q) error: %v", c.yuan, err)
		}
		if back != c.cents {
			t.Errorf("YuanStringToCents(%q) = %d, want %d", c.yuan, back, c.cents)
		}
	}
}

// generateTestRSAKeyPair 生成测试用密钥对：私钥 PKCS8 PEM（支付宝格式）+
// 对应公钥 PEM（充当「支付宝公钥」）。
func generateTestRSAKeyPair(t *testing.T) (privPEM, pubPEM string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}

	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("marshal pkcs8: %v", err)
	}
	privPEM = strings.TrimSpace(string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})))

	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("marshal pkix: %v", err)
	}
	pubPEM = strings.TrimSpace(string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})))

	return privPEM, pubPEM
}

// signAlipayNotification 按支付宝验签口径对通知字段签名：
// 去掉 sign/sign_type 后按 key 字典序拼 k=v& 串，RSA-SHA256 + Base64。
func signAlipayNotification(t *testing.T, privPEM string, fields map[string]string) map[string]string {
	t.Helper()
	block, _ := pem.Decode([]byte(privPEM))
	if block == nil {
		t.Fatal("decode private key pem failed")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("parse pkcs8: %v", err)
	}
	rsaKey := key.(*rsa.PrivateKey)

	keys := make([]string, 0, len(fields))
	for k := range fields {
		if k == "sign" || k == "sign_type" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, k+"="+fields[k])
	}
	content := strings.Join(pairs, "&")

	digest := sha256.Sum256([]byte(content))
	sig, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	out := make(map[string]string, len(fields)+2)
	for k, v := range fields {
		out[k] = v
	}
	out["sign"] = base64.StdEncoding.EncodeToString(sig)
	out["sign_type"] = "RSA2"
	return out
}

// TestMockGateway 模拟渠道行为边界。
func TestMockGateway(t *testing.T) {
	g := NewMockGateway()
	res, err := g.CreatePayment(context.Background(), PayRequest{OutTradeNo: "1", AmountCents: 100})
	if err != nil {
		t.Fatalf("mock create: %v", err)
	}
	if res.RedirectURL != "" || res.QRCode != "" {
		t.Errorf("mock should not produce pay url/qr: %+v", res)
	}
	if _, err = g.ParseNotify(context.Background(), &http.Request{}); err == nil {
		t.Errorf("mock parse notify should be unsupported")
	}
}

// TestAlipayNotifySignatureVerification 用测试密钥对走一遍
// 「签发通知 → 验签解析」往返：合法签名通过；篡改金额必须被拒绝。
func TestAlipayNotifySignatureVerification(t *testing.T) {
	privPEM, pubPEM := generateTestRSAKeyPair(t)

	gw, err := NewAlipayPageGateway("2021000000000000", false,
		`{"privateKey":`+quoteJSON(privPEM)+`,"alipayPublicKey":`+quoteJSON(pubPEM)+`}`)
	if err != nil {
		t.Fatalf("new alipay gateway: %v", err)
	}

	signedForm := signAlipayNotification(t, privPEM, map[string]string{
		"app_id":       "2021000000000000",
		"trade_no":     "2026102922001400001234567890",
		"out_trade_no": "42",
		"total_amount": "128.00",
		"trade_status": "TRADE_SUCCESS",
	})

	form := url.Values{}
	for k, v := range signedForm {
		form.Set(k, v)
	}

	req, err := http.NewRequest(http.MethodPost, "http://local/notify", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	nr, err := gw.ParseNotify(context.Background(), req)
	if err != nil {
		t.Fatalf("legit notification should verify: %v", err)
	}
	if nr.OutTradeNo != "42" || !nr.Success || nr.PaidCents != 12800 || nr.GatewayTradeNo == "" {
		t.Errorf("unexpected notify result: %+v", nr)
	}

	// 篡改金额 → 验签必须失败
	tampered := url.Values{}
	for k, v := range signedForm {
		tampered.Set(k, v)
	}
	tampered.Set("total_amount", "1.00")
	if _, err = gw.client.DecodeNotification(context.Background(), tampered); err == nil {
		t.Errorf("tampered notification must fail verification")
	}
}

func quoteJSON(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}
