package service

import (
	"context"

	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	authenticationV1 "go-wind-shop/api/gen/go/authentication/service/v1"
	"go-wind-shop/pkg/logx"

)

type MfaService struct {
	authenticationV1.UnimplementedMFAServiceServer

	log             *logx.Helper
	mfaServiceClient authenticationV1.MFAServiceClient
}

func NewMfaService(ctx *bootstrap.Context, mfaServiceClient authenticationV1.MFAServiceClient) *MfaService {
	return &MfaService{
		log:              logx.New(ctx.NewLoggerHelper("mfa/service/admin-service")),
		mfaServiceClient: mfaServiceClient,
	}
}

func (s *MfaService) GetMFAStatus(ctx context.Context, req *authenticationV1.GetMFAStatusRequest) (*authenticationV1.GetMFAStatusResponse, error) {
	return s.mfaServiceClient.GetMFAStatus(ctx, req)
}

func (s *MfaService) ListEnrolledMethods(ctx context.Context, req *authenticationV1.ListEnrolledMethodsRequest) (*authenticationV1.ListEnrolledMethodsResponse, error) {
	return s.mfaServiceClient.ListEnrolledMethods(ctx, req)
}

func (s *MfaService) StartEnrollMethod(ctx context.Context, req *authenticationV1.StartEnrollMethodRequest) (*authenticationV1.StartEnrollMethodResponse, error) {
	return s.mfaServiceClient.StartEnrollMethod(ctx, req)
}

func (s *MfaService) ConfirmEnrollMethod(ctx context.Context, req *authenticationV1.ConfirmEnrollMethodRequest) (*authenticationV1.ConfirmEnrollMethodResponse, error) {
	return s.mfaServiceClient.ConfirmEnrollMethod(ctx, req)
}

func (s *MfaService) DisableMFA(ctx context.Context, req *authenticationV1.DisableMFARequest) (*emptypb.Empty, error) {
	return s.mfaServiceClient.DisableMFA(ctx, req)
}

func (s *MfaService) RevokeMFADevice(ctx context.Context, req *authenticationV1.RevokeMFADeviceRequest) (*emptypb.Empty, error) {
	return s.mfaServiceClient.RevokeMFADevice(ctx, req)
}

// VerifyMFAChallenge 登录时 MFA 挑战验证（转发 core；core 通过后签发真 token）。
func (s *MfaService) VerifyMFAChallenge(ctx context.Context, req *authenticationV1.VerifyMFAChallengeRequest) (*authenticationV1.LoginResponse, error) {
	return s.mfaServiceClient.VerifyMFAChallenge(ctx, req)
}
