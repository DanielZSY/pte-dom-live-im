package grpcapi

import (
	"context"

	"pte_live_im/define/retcode"
	"pte_live_im/pkg/appid"
	"pte_live_im/protobuf/imapi"
	"pte_live_im/servers"

	"google.golang.org/grpc"
)

type ctxKey string

const appIDKey ctxKey = "appId"

func AppIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(appIDKey).(string)
	return v
}

// SystemIDFromContext 兼容旧调用
func SystemIDFromContext(ctx context.Context) string {
	return AppIDFromContext(ctx)
}

func AppIDInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if info.FullMethod == "/imapi.ImApi/Register" {
		return handler(ctx, req)
	}

	appId := appid.FromContext(ctx)
	if err := servers.ValidateAppID(appId); err != nil {
		return &imapi.ApiReply{
			Code: retcode.FAIL,
			Msg:  err.Error(),
		}, nil
	}

	return handler(context.WithValue(ctx, appIDKey, appId), req)
}

// SystemIDInterceptor 兼容旧名称
func SystemIDInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	return AppIDInterceptor(ctx, req, info, handler)
}
