package data

import (
	"testing"

	field_mask "github.com/tx7do/go-utils/fieldmaskutil"
	identityV1 "go-wind-shop/api/gen/go/identity/service/v1"
)

func TestZZDebugMask(t *testing.T) {
	u := &identityV1.User{}
	updateUserReq := &identityV1.UpdateUserRequest{
		UpdateMask: &field_mask.FieldMask{
			Paths: []string{"userName", "realName", "avatar", "roleId"},
		},
	}
	updateUserReq.UpdateMask.Normalize()
	t.Logf("normalized: %v", updateUserReq.UpdateMask.Paths)
	if !updateUserReq.UpdateMask.IsValid(u) {
		t.Logf("IsValid false for user proto")
	}
}
