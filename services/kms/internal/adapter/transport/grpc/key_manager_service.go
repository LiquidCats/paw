package grpc

import (
	"context"
	"fmt"

	"github.com/LiquidCats/paw/services/litehsm/internal/app/domain/entities"
	"github.com/LiquidCats/paw/services/litehsm/internal/app/usecase/hsm"
	"github.com/LiquidCats/paw/services/litehsm/pkg/unsafe"
	v1 "github.com/LiquidCats/paw/protos/gen/go/services/litehsm/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type KeyManagerServiceServer struct {
	*v1.UnimplementedKeyManagerServiceServer

	createKeyUseCase     *hsm.KeyManagerCreateKey
	setExpirationUseCase *hsm.KeyManagerSetExpiration
	setStatusUseCase     *hsm.KeyManagerSetStatus
}

func NewKeyManagerServiceServer(
	createKeyUseCase *hsm.KeyManagerCreateKey,
	setExpirationUseCase *hsm.KeyManagerSetExpiration,
	setStatusUseCase *hsm.KeyManagerSetStatus,
) *KeyManagerServiceServer {
	return &KeyManagerServiceServer{
		createKeyUseCase:     createKeyUseCase,
		setExpirationUseCase: setExpirationUseCase,
		setStatusUseCase:     setStatusUseCase,
	}
}

func (srv *KeyManagerServiceServer) AttachToGRPC(s grpc.ServiceRegistrar) {
	v1.RegisterKeyManagerServiceServer(s, srv)
}

func (srv *KeyManagerServiceServer) CreateKey(
	ctx context.Context,
	req *v1.CreateKeyRequest,
) (*v1.CreateKeyResponse, error) {
	curve, err := entities.CurveTypeFromProto(req.GetCurve())
	if err != nil {
		return nil, fmt.Errorf("struct=Server, method=CreateKey, call=CurveTypeFromProto: %w", err)
	}

	algorithm, err := entities.AlgorithmTypeFromProto(req.GetAlgorithm())
	if err != nil {
		return nil, fmt.Errorf("struct=Server, method=CreateKey, call=AlgorithmTypeFromProto: %w", err)
	}

	entry := entities.KeyEntry{
		Alias:     req.GetAlias(),
		Curve:     curve,
		Algorithm: algorithm,
	}
	if req.GetExpiration() != nil {
		entry.ExpiresAt = new(req.GetExpiration().AsTime())
	}

	newKey, err := srv.createKeyUseCase.Handle(ctx, entry)
	if err != nil {
		return nil, fmt.Errorf("struct=Server, method=CreateKey, call=createKeyUseCase.Handle: %w", err)
	}

	resp := v1.CreateKeyResponse{
		KeyId:     newKey.KeyID.String(),
		Alias:     newKey.Alias,
		Curve:     newKey.Curve.ToProto(),
		Algorithm: newKey.Algorithm.ToProto(),
		Status:    newKey.Status.ToProto(),
	}

	if newKey.ExpiresAt != nil {
		resp.Expiration = timestamppb.New(*newKey.ExpiresAt)
	}

	return &resp, nil
}

func (srv *KeyManagerServiceServer) SetKeyExpiration(
	ctx context.Context,
	req *v1.SetKeyExpirationRequest,
) (*v1.SetKeyExpirationResponse, error) {
	if err := srv.setExpirationUseCase.Handle(
		ctx,
		entities.KeyID(unsafe.StringToBytes(req.GetKeyId())),
		req.GetExpiration().AsTime(),
	); err != nil {
		return nil, fmt.Errorf("struct=Server, method=SetKeyExpiration, call=setExpirationUseCase.Handle: %w", err)
	}

	return &v1.SetKeyExpirationResponse{}, nil
}

func (srv *KeyManagerServiceServer) EnableKey(
	ctx context.Context,
	req *v1.EnableKeyRequest,
) (*v1.EnableKeyResponse, error) {
	if err := srv.setStatusUseCase.Handle(
		ctx,
		entities.KeyID(unsafe.StringToBytes(req.GetKeyId())),
		entities.KeyStatusEnabled,
	); err != nil {
		return nil, fmt.Errorf("struct=Server, method=DeleteKey, call=deleteKeyUseCase.Handle: %w", err)
	}

	return &v1.EnableKeyResponse{}, nil
}

func (srv *KeyManagerServiceServer) DisableKey(
	ctx context.Context,
	req *v1.DisableKeyRequest,
) (*v1.DisableKeyResponse, error) {
	if err := srv.setStatusUseCase.Handle(
		ctx,
		entities.KeyID(unsafe.StringToBytes(req.GetKeyId())),
		entities.KeyStatusDisabled,
	); err != nil {
		return nil, fmt.Errorf("struct=Server, method=DeleteKey, call=deleteKeyUseCase.Handle: %w", err)
	}

	return &v1.DisableKeyResponse{}, nil
}

func (srv *KeyManagerServiceServer) DeleteKey(
	ctx context.Context,
	req *v1.DeleteKeyRequest,
) (*v1.DeleteKeyResponse, error) {
	if err := srv.setStatusUseCase.Handle(
		ctx,
		entities.KeyID(unsafe.StringToBytes(req.GetKeyId())),
		entities.KeyStatusDeleted,
	); err != nil {
		return nil, fmt.Errorf("struct=Server, method=DeleteKey, call=deleteKeyUseCase.Handle: %w", err)
	}

	return &v1.DeleteKeyResponse{}, nil
}
