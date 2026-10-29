package service

import (
	"testing"

	"github.com/tx7do/go-utils/mapper"
	paymentV1 "go-wind-shop/api/gen/go/payment/service/v1"
	"go-wind-shop/app/core/service/internal/data/ent/paymentchannelconfig"
)

func TestChannelConverter(t *testing.T) {
	c := mapper.NewEnumTypeConverter[paymentV1.PaymentChannelType, paymentchannelconfig.Channel](paymentV1.PaymentChannelType_name, paymentV1.PaymentChannelType_value)

	v := paymentV1.PaymentChannelType_PAYMENT_CHANNEL_MOCK
	out := c.ToEntity(&v)
	t.Logf("ToEntity(MOCK) = %v", *out)

	back := c.ToDTO(out)
	t.Logf("ToDTO(%q) = %v", string(*out), *back)
}
