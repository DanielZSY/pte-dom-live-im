package grpcapi

import (
	"context"

	"pte_live_im/define/retcode"
	"pte_live_im/protobuf/imapi"
	"pte_live_im/servers"
)

type Server struct {
	imapi.UnimplementedImApiServer
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Register(_ context.Context, req *imapi.RegisterReq) (*imapi.ApiReply, error) {
	if req.GetSystemId() == "" {
		return fail("appId不能为空"), nil
	}

	if err := servers.Register(req.GetSystemId()); err != nil {
		return fail(err.Error()), nil
	}

	return success(""), nil
}

func (s *Server) SendToClient(ctx context.Context, req *imapi.SendToClientReq) (*imapi.ApiReply, error) {
	if req.GetClientId() == "" {
		return fail("clientId不能为空"), nil
	}

	messageID := servers.SendMessage2Client(
		req.GetClientId(),
		req.GetSendUserId(),
		int(req.GetCode()),
		req.GetMsg(),
		strPtr(req.GetData()),
	)

	return success(messageID), nil
}

func (s *Server) SendToClients(ctx context.Context, req *imapi.SendToClientsReq) (*imapi.ApiReply, error) {
	if len(req.GetClientIds()) == 0 {
		return fail("clientIds不能为空"), nil
	}

	data := req.GetData()
	for _, clientID := range req.GetClientIds() {
		_ = servers.SendMessage2Client(clientID, req.GetSendUserId(), int(req.GetCode()), req.GetMsg(), &data)
	}

	return success(""), nil
}

func (s *Server) BindToGroup(ctx context.Context, req *imapi.BindToGroupReq) (*imapi.ApiReply, error) {
	if req.GetClientId() == "" {
		return fail("clientId不能为空"), nil
	}
	if req.GetGroupName() == "" {
		return fail("groupName不能为空"), nil
	}

	systemID := SystemIDFromContext(ctx)
	servers.AddClient2Group(systemID, req.GetGroupName(), req.GetClientId(), req.GetUserId(), req.GetExtend())

	return success(""), nil
}

func (s *Server) SendToGroup(ctx context.Context, req *imapi.SendToGroupReq) (*imapi.ApiReply, error) {
	if req.GetGroupName() == "" {
		return fail("groupName不能为空"), nil
	}

	systemID := SystemIDFromContext(ctx)
	messageID := servers.SendMessage2Group(
		systemID,
		req.GetSendUserId(),
		req.GetGroupName(),
		int(req.GetCode()),
		req.GetMsg(),
		strPtr(req.GetData()),
	)

	return success(messageID), nil
}

func (s *Server) GetOnlineList(ctx context.Context, req *imapi.GetOnlineListReq) (*imapi.ApiReply, error) {
	if req.GetGroupName() == "" {
		return fail("groupName不能为空"), nil
	}

	systemID := SystemIDFromContext(ctx)
	groupName := req.GetGroupName()
	ret := servers.GetOnlineList(&systemID, &groupName)

	list, _ := ret["list"].([]string)
	count, _ := ret["count"].(int)

	return &imapi.ApiReply{
		Code:  retcode.SUCCESS,
		Msg:   "success",
		Count: int32(count),
		List:  list,
	}, nil
}

func (s *Server) CloseClient(ctx context.Context, req *imapi.CloseClientReq) (*imapi.ApiReply, error) {
	if req.GetClientId() == "" {
		return fail("clientId不能为空"), nil
	}

	systemID := SystemIDFromContext(ctx)
	servers.CloseClient(req.GetClientId(), systemID)

	return success(""), nil
}

func success(messageID string) *imapi.ApiReply {
	return &imapi.ApiReply{
		Code:      retcode.SUCCESS,
		Msg:       "success",
		MessageId: messageID,
	}
}

func fail(msg string) *imapi.ApiReply {
	return &imapi.ApiReply{
		Code: retcode.FAIL,
		Msg:  msg,
	}
}

func strPtr(s string) *string {
	return &s
}
