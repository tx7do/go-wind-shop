// Package logx 是 kratos-bootstrap/logger v0.1.3 Helper 的升级兼容层。
//
// 背景：script 域引入的依赖链强制 bootstrap/logger 升级到 v0.1.3，其 Helper
// 的方法签名从"无 ctx"改为"ctx-first"（Errorf(ctx, format, ...)）。全库约
// 9000 处旧签名调用逐点迁移代价过高，且旧调用本就不携带 ctx（老 Helper
// 内部同样无 ctx），语义等价——故提供本适配器保留旧签名，内部以
// context.Background() 调新 API。
//
// 新代码请直接使用 logger.Helper 的 ctx-first API；本层仅为存量平滑过渡。
package logx

import (
	"context"
	"os"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	bLogger "github.com/tx7do/kratos-bootstrap/logger"
)

// Helper 兼容适配器：保留 v0.1.2 时代的无 ctx 调用签名。
type Helper struct {
	h *bLogger.Helper
}

// New 包装 bootstrap 的 Helper。
func New(h *bLogger.Helper) *Helper {
	if h == nil {
		return nil
	}
	return &Helper{h: h}
}

// NewFromLogger 由 bootstrap Logger 构造。
func NewFromLogger(l bLogger.Logger) *Helper {
	return New(bLogger.NewHelper(l))
}

// WithLogger 派生带键值的新 Helper。
func (l *Helper) WithLogger(kv ...any) *Helper {
	return New(bLogger.NewHelper(l.h.With(kv...)))
}

// AsKratosLogger 适配为 kratos log.Logger（ent 等组件以该接口消费）。
func (l *Helper) AsKratosLogger() log.Logger {
	return &kratosAdapter{h: l.h}
}

func (l *Helper) Debug(args ...any) {
	l.h.Debug(context.Background(), fmt.Sprint(args...))
}

func (l *Helper) Info(args ...any) {
	l.h.Info(context.Background(), fmt.Sprint(args...))
}

func (l *Helper) Warn(args ...any) {
	l.h.Warn(context.Background(), fmt.Sprint(args...))
}

func (l *Helper) Error(args ...any) {
	l.h.Error(context.Background(), fmt.Sprint(args...))
}

func (l *Helper) Infow(msg string, kv ...any) {
	l.h.Info(context.Background(), msg, kv...)
}

func (l *Helper) Warnw(msg string, kv ...any) {
	l.h.Warn(context.Background(), msg, kv...)
}

func (l *Helper) Errorw(msg string, kv ...any) {
	l.h.Error(context.Background(), msg, kv...)
}

func (l *Helper) Debugw(msg string, kv ...any) {
	l.h.Debug(context.Background(), msg, kv...)
}

func (l *Helper) Debugf(format string, args ...any) {
	l.h.Debugf(context.Background(), format, args...)
}

func (l *Helper) Infof(format string, args ...any) {
	l.h.Infof(context.Background(), format, args...)
}

func (l *Helper) Warnf(format string, args ...any) {
	l.h.Warnf(context.Background(), format, args...)
}

func (l *Helper) Errorf(format string, args ...any) {
	l.h.Errorf(context.Background(), format, args...)
}

func (l *Helper) Fatal(args ...any) {
	l.Error(args...)
	os.Exit(1)
}

func (l *Helper) Fatalf(format string, args ...any) {
	l.Errorf(format, args...)
	os.Exit(1)
}

// kratosAdapter 把 bootstrap Helper 适配为 kratos log.Logger。
type kratosAdapter struct {
	h *bLogger.Helper
}

func (a *kratosAdapter) Log(level log.Level, keyvals ...any) error {
	msg := fmt.Sprint(keyvals...)
	switch level {
	case log.LevelDebug:
		a.h.Debug(context.Background(), msg)
	case log.LevelWarn:
		a.h.Warn(context.Background(), msg)
	case log.LevelError:
		a.h.Error(context.Background(), msg)
	default:
		a.h.Info(context.Background(), msg)
	}
	return nil
}

// AsKratos 把 bootstrap Logger 适配为 kratos log.Logger（kratos 中间件/三方组件消费）。
func AsKratos(l bLogger.Logger) log.Logger {
	return &bToK{l: l}
}

type bToK struct {
	l bLogger.Logger
}

func (a *bToK) Log(level log.Level, keyvals ...any) error {
	msg := fmt.Sprint(keyvals...)
	switch level {
	case log.LevelDebug:
		a.l.Debug(context.Background(), msg)
	case log.LevelWarn:
		a.l.Warn(context.Background(), msg)
	case log.LevelError:
		a.l.Error(context.Background(), msg)
	default:
		a.l.Info(context.Background(), msg)
	}
	return nil
}
