package service

import (
	"testing"

	"github.com/tx7do/go-utils/mapper"
	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
	"go-wind-shop/app/core/service/internal/data/ent/pointsledger"
)

func TestPointsTypeConverter(t *testing.T) {
	c := mapper.NewEnumTypeConverter[catalogV1.PointsLedgerType, pointsledger.Type](catalogV1.PointsLedgerType_name, catalogV1.PointsLedgerType_value)

	v := pointsledger.TypePointsTypeEarn
	out := c.ToDTO(&v)
	if out == nil {
		t.Fatal("ToDTO returned nil")
	}
	t.Logf("ToDTO(EARN) = %v", *out)
	if *out != catalogV1.PointsLedgerType_POINTS_TYPE_EARN {
		t.Errorf("mismatch: %v", *out)
	}
}
