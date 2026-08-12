package task

const (
	CouponExpireSweepTaskType = "coupon_expire_sweep"
)

// CouponExpireSweepTaskData 优惠券过期清扫任务的载荷（周期性任务，无实体 ID）。
type CouponExpireSweepTaskData struct {
	Name string `json:"name"`
}
