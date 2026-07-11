package grpcserver

import (
	"net"

	"google.golang.org/grpc"

	"pte_live_im/pkg/setting"
	"pte_live_im/protobuf/imapi"
	"pte_live_im/servers"
	"pte_live_im/servers/grpcapi"
	"pte_live_im/servers/pb"
	"pte_live_im/tools/util"
)

func Init() {
	go serve(":" + setting.CommonSetting.RPCPort)
}

func serve(port string) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(grpcapi.SystemIDInterceptor))
	imapi.RegisterImApiServer(s, grpcapi.NewServer())

	if util.IsCluster() {
		pb.RegisterCommonServiceServer(s, &servers.CommonServiceServer{})
	}

	if err = s.Serve(lis); err != nil {
		panic(err)
	}
}
