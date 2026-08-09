package privacy

import (
	"context"
	"fmt"
	"reflect"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"

	"github.com/tx7do/go-crud/viewer"
)

// UserPrivacy 是一个行级隔离规则，强制普通用户只能查询/变更 user_id = 自身 userID 的记录。
//
// 与 TenantPrivacy 的区别：
//   - TenantPrivacy 由 TenantID mixin 注入，覆盖 tenant_id 字段
//   - UserPrivacy 由具体实体（Order/Cart）在 Policy() 中显式声明，覆盖 user_id 字段
//
// 系统/平台视图放行（不做 user 过滤），无 viewer 拒绝。
//
// 适用实体：Order、Cart（购物车与订单按用户隔离，同租户内不可互看）。
type UserPrivacy struct {
	decision error
	// ColumnName 指定行级隔离所用的列名。默认空串表示 "user_id"
	// （覆盖 Order/Cart/PaymentTransaction/Shipment 等多数实体）。
	// 列名不同（如 InternalMessageRecipient 的 recipient_user_id）的实体
	// 在 Policy() 中显式传入，使 EvalQuery/EvalMutation 的 WHERE 注入
	// 落到正确列。
	ColumnName string
}

// userColumn 返回行级隔离所用的列名，空串回退到 "user_id"。
func (u UserPrivacy) userColumn() string {
	if u.ColumnName != "" {
		return u.ColumnName
	}
	return "user_id"
}

func (u UserPrivacy) EvalQuery(ctx context.Context, query ent.Query) error {
	vc, exist := viewer.FromContext(ctx)
	if !exist {
		return fmt.Errorf("security: missing ViewerContext in context")
	}
	// 平台/系统视图放行：允许查看全量数据
	if vc.IsPlatformContext() || vc.IsSystemContext() {
		return nil
	}
	uid := vc.UserID()
	if uid == 0 {
		return fmt.Errorf("security: viewer has no user id")
	}
	if err := injectUserWhere(query, uid, u.userColumn()); err != nil {
		return err
	}
	return nil
}

func (u UserPrivacy) EvalMutation(ctx context.Context, m ent.Mutation) error {
	vc, exist := viewer.FromContext(ctx)
	if !exist {
		return fmt.Errorf("missing ViewerContext in context")
	}
	// 平台/系统视图放行
	if vc.IsPlatformContext() || vc.IsSystemContext() {
		return nil
	}
	uid := vc.UserID()
	if uid == 0 {
		return fmt.Errorf("security: viewer has no user id")
	}
	op := m.Op()
	if !op.Is(ent.OpCreate) {
		// 非 Create 的 mutation（Update/Delete）同样按归属列过滤
		if err := injectUserWhereMutation(m, uid, u.userColumn()); err != nil {
			return err
		}
		return nil
	}
	// Create：仅对 user_id 列的实体强制覆盖归属（防伪造）。
	// 列名不同的实体（如 InternalMessageRecipient 的 recipient_user_id）
	// 不走客户端 Create（系统消息投递由系统视图在上方放行），此处直接返回。
	if u.userColumn() != "user_id" {
		return nil
	}
	if s, ok := m.(interface{ SetUserID(uint32) }); ok {
		s.SetUserID(uint32(uid))
		return nil
	}
	// 兜底：反射 SetField
	rv := reflect.ValueOf(m)
	if mf := rv.MethodByName("SetField"); mf.IsValid() && mf.Kind() == reflect.Func && mf.Type().NumIn() == 2 {
		mf.Call([]reflect.Value{reflect.ValueOf("user_id"), reflect.ValueOf(uid)})
		return nil
	}
	return fmt.Errorf("unable to set user_id on mutation")
}

// injectUserWhere 在查询上注入 columnName = uid 过滤（通过反射调用 query.Where）。
func injectUserWhere(query ent.Query, uid uint64, columnName string) error {
	rv := reflect.ValueOf(query)
	mf := rv.MethodByName("Where")
	if !mf.IsValid() || mf.Kind() != reflect.Func {
		return nil
	}
	mt := mf.Type()
	if !mt.IsVariadic() || mt.NumIn() != 1 {
		return nil
	}
	elem := mt.In(0).Elem()
	selPtrType := reflect.TypeOf((*sql.Selector)(nil))
	if elem.Kind() != reflect.Func || elem.NumIn() < 1 || elem.In(0) != selPtrType {
		return nil
	}
	fn := func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(columnName), uid))
	}
	valFn := reflect.ValueOf(fn)
	if valFn.Type() != elem {
		if valFn.Type().ConvertibleTo(elem) {
			valFn = valFn.Convert(elem)
		} else {
			valFn = reflect.MakeFunc(elem, func(in []reflect.Value) []reflect.Value {
				s := in[0].Interface().(*sql.Selector)
				fn(s)
				return nil
			})
		}
	}
	_ = mf.Call([]reflect.Value{valFn})
	return nil
}

// injectUserWhereMutation 在 mutation（Update/Delete）上注入 columnName = uid 过滤。
func injectUserWhereMutation(m ent.Mutation, uid uint64, columnName string) error {
	rv := reflect.ValueOf(m)
	mf := rv.MethodByName("Where")
	if !mf.IsValid() || mf.Kind() != reflect.Func {
		return nil
	}
	mt := mf.Type()
	if !mt.IsVariadic() || mt.NumIn() != 1 {
		return nil
	}
	elem := mt.In(0).Elem()
	selPtrType := reflect.TypeOf((*sql.Selector)(nil))
	if elem.Kind() != reflect.Func || elem.NumIn() < 1 || elem.In(0) != selPtrType {
		return nil
	}
	fn := func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(columnName), uid))
	}
	valFn := reflect.ValueOf(fn)
	if valFn.Type() != elem {
		if valFn.Type().ConvertibleTo(elem) {
			valFn = valFn.Convert(elem)
		} else {
			valFn = reflect.MakeFunc(elem, func(in []reflect.Value) []reflect.Value {
				s := in[0].Interface().(*sql.Selector)
				fn(s)
				return nil
			})
		}
	}
	_ = mf.Call([]reflect.Value{valFn})
	return nil
}

// UserPrivacy 实现 ent.Policy（EvalQuery/EvalMutation），由 Order/Cart schema 的 Policy() 返回。
