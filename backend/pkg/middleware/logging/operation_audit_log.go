package logging

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/tx7do/go-utils/timeutil"
	"github.com/tx7do/go-utils/trans"
	"google.golang.org/protobuf/proto"

	auditV1 "go-wind-shop/api/gen/go/audit/service/v1"

	appViewer "go-wind-shop/pkg/entgo/viewer"
)

type OperationAuditLogMiddleware struct {
	op *options
}

func NewOperationAuditLogMiddleware(op *options) *OperationAuditLogMiddleware {
	return &OperationAuditLogMiddleware{
		op: op,
	}
}

func (a *OperationAuditLogMiddleware) Name() string {
	return "OperationAuditLogMiddleware"
}

// httpMethodToAction 将 HTTP 方法映射到操作审计的 ActionType。
// GET → READ，POST → CREATE，PUT/PATCH → UPDATE，DELETE → DELETE，
// 其它（OPTIONS/HEAD 等）→ OTHER。
func httpMethodToAction(method string) auditV1.OperationAuditLog_ActionType {
	switch method {
	case "GET":
		return auditV1.OperationAuditLog_READ
	case "POST":
		return auditV1.OperationAuditLog_CREATE
	case "PUT", "PATCH":
		return auditV1.OperationAuditLog_UPDATE
	case "DELETE":
		return auditV1.OperationAuditLog_DELETE
	default:
		return auditV1.OperationAuditLog_OTHER
	}
}

func (a *OperationAuditLogMiddleware) Handle(ctx context.Context, htr *http.Transport, middleErr error) {
	if a.op == nil || htr == nil {
		return
	}

	operationAuditLog := &auditV1.OperationAuditLog{}

	// 动作类型由 HTTP 方法派生。资源类型/资源ID/前后数据需各业务服务在
	// 自身写路径中补充（通用中间件无法解析业务实体归属），此处留空。
	operationAuditLog.Action = trans.Ptr(httpMethodToAction(htr.Request().Method))

	clientIp := getClientRealIP(htr.Request())

	operationAuditLog.IpAddress = trans.Ptr(clientIp)
	operationAuditLog.CreatedAt = timeutil.TimeToTimestamppb(trans.Ptr(time.Now()))
	operationAuditLog.GeoLocation = fillGeoLocation(clientIp)
	operationAuditLog.RequestId = trans.Ptr(getRequestId(htr.Request()))

	ut := extractAuthToken(htr)
	if ut != nil {
		operationAuditLog.UserId = trans.Ptr(ut.UserId)
		operationAuditLog.TenantId = ut.TenantId
		operationAuditLog.Username = ut.Username
	}

	// 获取错误码和是否成功
	statusCode, reason, success := getStatusCode(middleErr)

	operationAuditLog.Success = trans.Ptr(success)
	operationAuditLog.FailureReason = trans.Ptr(reason)
	_ = statusCode

	// 计算哈希和签名
	operationAuditLog.LogHash = trans.Ptr(a.hashLog(operationAuditLog))
	operationAuditLog.Signature = a.signature(operationAuditLog)

	// 写入日志
	if a.op.writeOperationLogFunc != nil {
		ctx = appViewer.NewSystemViewerContext(ctx)
		_ = a.op.writeOperationLogFunc(ctx, operationAuditLog)
	}
}

// hashLog 计算日志的 SHA256 哈希（十六进制小写字符串）
// 规则：排除 log_hash 和 signature 字段，Protobuf 确定性序列化后哈希
func (a *OperationAuditLogMiddleware) hashLog(operationAuditLog *auditV1.OperationAuditLog) string {
	if operationAuditLog == nil {
		return ""
	}

	operationAuditLog.LogHash = nil
	operationAuditLog.Signature = nil

	rawBytes, err := proto.Marshal(operationAuditLog)
	if err != nil {
		fmt.Printf("marshal log failed: %v\n", err)
		return ""
	}

	hash := sha256.Sum256(rawBytes)
	return hex.EncodeToString(hash[:])
}

// signature 生成日志的 ECDSA 数字签名
// 签名内容：tenant_id + user_id + created_at（原始时间戳） + log_hash
// 返回：ECDSA 签名字节数组（r+s 拼接，DER 格式）
func (a *OperationAuditLogMiddleware) signature(operationAuditLog *auditV1.OperationAuditLog) []byte {
	if operationAuditLog == nil || a.op.ecPrivateKey == nil {
		return nil
	}

	tenantID := operationAuditLog.GetTenantId()
	userID := operationAuditLog.GetUserId()
	logHash := operationAuditLog.GetLogHash()
	createdAt := operationAuditLog.GetCreatedAt()

	type signContent struct {
		TenantID uint32 `json:"tenant_id"`
		UserID   uint32 `json:"user_id"`
		Sec      int64  `json:"sec"`   // createdAt 秒数
		Nanos    int32  `json:"nanos"` // createdAt 纳秒数
		LogHash  string `json:"log_hash"`
	}
	sc := signContent{
		TenantID: tenantID,
		UserID:   userID,
		LogHash:  logHash,
	}
	if createdAt != nil {
		sc.Sec = createdAt.Seconds
		sc.Nanos = createdAt.Nanos
	}

	scBytes, err := json.Marshal(sc)
	if err != nil {
		fmt.Printf("marshal sign content failed: %v\n", err)
		return nil
	}

	scHash := sha256.Sum256(scBytes)

	r, s, err := ecdsa.Sign(rand.Reader, a.op.ecPrivateKey, scHash[:])
	if err != nil {
		fmt.Printf("ECDSA sign failed: %v\n", err)
		return nil
	}

	signBytes, err := encodeDER(r, s)
	if err != nil {
		fmt.Printf("encode DER failed: %v\n", err)
		return nil
	}

	return signBytes
}
