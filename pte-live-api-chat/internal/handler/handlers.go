package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"pte_live_api_chat/internal/service"
	"pte_live_api_chat/pkg/response"
)

type Handlers struct {
	tokens      *service.IMTokenService
	shopMessage *service.ShopMessageService
	chat        *service.ChatService
	scene       *service.SceneService
	ops         *service.OpsService
}

func NewHandlers(tokens *service.IMTokenService, shopMessage *service.ShopMessageService, chat *service.ChatService, scene *service.SceneService, ops *service.OpsService) *Handlers {
	return &Handlers{tokens: tokens, shopMessage: shopMessage, chat: chat, scene: scene, ops: ops}
}

func (h *Handlers) Ping(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}
	response.Success(w, map[string]string{"service": "api-chat", "status": "ok"})
}

func (h *Handlers) Healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}
	response.Success(w, h.ops.Health(r.Context(), false))
}

func (h *Handlers) Readyz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}
	res := h.ops.Health(r.Context(), true)
	if !res.Ready {
		response.JSON(w, http.StatusServiceUnavailable, response.Body{Code: 0, Msg: res.Error, Data: res})
		return
	}
	response.Success(w, res)
}

func (h *Handlers) OpsMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	res, err := h.ops.Metrics(r.Context())
	if err != nil {
		response.Error(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	response.Success(w, res)
}

func (h *Handlers) PrometheusMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}
	body, err := h.ops.Prometheus(r.Context())
	if err != nil {
		response.Error(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	_, _ = w.Write([]byte(body))
}

func (h *Handlers) IMToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}
	var req service.IMTokenRequest
	if r.Method == http.MethodPost {
		if err := decodeJSON(r, &req); err != nil {
			response.Error(w, http.StatusOK, "参数错误")
			return
		}
	} else {
		q := r.URL.Query()
		req.AppID = q.Get("app_id")
		req.UserID = q.Get("user_id")
		req.Scene = q.Get("scene")
		req.RoomID = q.Get("room_id")
		req.DeviceID = q.Get("device_id")
		req.Platform = q.Get("platform")
	}
	if req.AppID == "" {
		req.AppID = firstHeader(r, "AppId", "appid")
	}
	if req.Token == "" && req.IMToken == "" {
		req.Token = bearerToken(r)
	}
	res, err := h.tokens.Issue(req, bearerToken(r))
	if err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	response.Success(w, res)
}

func (h *Handlers) IMUserSig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}
	var req service.IMTokenRequest
	if r.Method == http.MethodPost {
		if err := decodeJSON(r, &req); err != nil {
			response.Error(w, http.StatusOK, "参数错误")
			return
		}
	} else {
		q := r.URL.Query()
		req.AppID = q.Get("app_id")
		req.SDKAppID = q.Get("sdk_app_id")
		req.UserID = q.Get("user_id")
		req.Identifier = q.Get("identifier")
		req.UserType = q.Get("user_type")
		req.Scene = q.Get("scene")
		req.RoomID = q.Get("room_id")
		req.DeviceID = q.Get("device_id")
		req.Platform = q.Get("platform")
		req.Expire = int64(atoi(q.Get("expire")))
	}
	if req.AppID == "" {
		req.AppID = firstHeader(r, "AppId", "appid")
	}
	res, err := h.tokens.IssueUserSig(r.Context(), req, bearerToken(r), clientIP(r))
	if err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	response.Success(w, res)
}

func (h *Handlers) IMUserSigVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.UserSigVerifyRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	res, err := h.tokens.VerifyUserSig(r.Context(), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) ShopMessageSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.ShopMessageSendRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	res, err := h.shopMessage.Send(r.Context(), requestAppID(r, req.AppID), req, isShadowRequest(r))
	if err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	response.Success(w, res)
}

func (h *Handlers) ShopMessageList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}
	req, ok := decodeShopMessageListRequest(w, r)
	if !ok {
		return
	}
	var (
		res *service.ListResult
		err error
	)
	switch r.URL.Path {
	case "/api/v1/scene/shop/message/recent":
		res, err = h.shopMessage.Recent(r.Context(), requestAppID(r, req.AppID), req)
	case "/api/v1/scene/shop/message/history":
		res, err = h.shopMessage.History(r.Context(), requestAppID(r, req.AppID), req)
	default:
		res, err = h.shopMessage.AuditList(r.Context(), requestAppID(r, req.AppID), req)
	}
	if err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	response.Success(w, res)
}

func (h *Handlers) ShopMessagePendingCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}
	req, ok := decodeShopMessageListRequest(w, r)
	if !ok {
		return
	}
	res, err := h.shopMessage.PendingCount(r.Context(), requestAppID(r, req.AppID), req)
	if err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	response.Success(w, res)
}

func (h *Handlers) ShopMessageAuditSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.ShopMessageAuditRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	res, err := h.shopMessage.AuditSubmit(r.Context(), requestAppID(r, req.AppID), req)
	if err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	response.Success(w, res)
}

func (h *Handlers) ChatOpenSingle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.OpenSingleRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	res, err := h.chat.OpenSingle(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) ChatCreateGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.CreateGroupRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	res, err := h.chat.CreateGroup(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) ChatConversationList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.ConversationListRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	res, err := h.chat.ListConversations(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) ChatConversationDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.ConversationDetailRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	res, err := h.chat.ConversationDetail(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) ChatConversationRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.ConversationReadRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	res, err := h.chat.MarkRead(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) ChatMessageSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.MessageSendRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	res, err := h.chat.SendMessage(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) ChatMessageHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.MessageHistoryRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	res, err := h.chat.History(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) ChatMessageSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.MessageSyncRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	res, err := h.chat.SyncMessages(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) ChatMessageAck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.MessageAckRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	res, err := h.chat.AckMessages(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) ChatMessageRecall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.MessageActionRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	res, err := h.chat.RecallMessage(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) ChatMessageDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.MessageActionRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	res, err := h.chat.DeleteMessage(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) ChatMemberList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.MemberListRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	res, err := h.chat.ListMembers(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) ChatMemberAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.MemberAddRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	res, err := h.chat.AddMembers(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) ChatMemberRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.MemberRemoveRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	res, err := h.chat.RemoveMember(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) SceneRoomOpen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.SceneRoomRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	applySceneTypeFromPath(r, &req.SceneType)
	res, err := h.scene.OpenRoom(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) SceneRoomClose(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.SceneRoomRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	applySceneTypeFromPath(r, &req.SceneType)
	res, err := h.scene.CloseRoom(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) SceneRoomList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.SceneListRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	applySceneTypeFromPath(r, &req.SceneType)
	res, err := h.scene.ListRooms(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) SceneRoomDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.SceneRoomRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	applySceneTypeFromPath(r, &req.SceneType)
	res, err := h.scene.Detail(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) SceneRoomEnter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.SceneMemberRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	applySceneTypeFromPath(r, &req.SceneType)
	res, err := h.scene.EnterRoom(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) SceneRoomLeave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.SceneMemberRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	applySceneTypeFromPath(r, &req.SceneType)
	res, err := h.scene.LeaveRoom(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) SceneMemberList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.SceneMemberRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	applySceneTypeFromPath(r, &req.SceneType)
	res, err := h.scene.ListMembers(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) SceneSeatAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.SceneSeatActionRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	applySceneTypeFromPath(r, &req.SceneType)
	res, err := h.scene.SeatAction(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) SceneModerationAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.SceneModerationRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	applySceneTypeFromPath(r, &req.SceneType)
	res, err := h.scene.ModerationAction(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) ScenePKStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.ScenePKRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	applySceneTypeFromPath(r, &req.SceneType)
	res, err := h.scene.StartPK(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) ScenePKInvite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.ScenePKRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	applySceneTypeFromPath(r, &req.SceneType)
	res, err := h.scene.InvitePK(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) ScenePKAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.ScenePKRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	applySceneTypeFromPath(r, &req.SceneType)
	res, err := h.scene.PKAction(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) ScenePKEnd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.ScenePKRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	applySceneTypeFromPath(r, &req.SceneType)
	res, err := h.scene.EndPK(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) SceneEventSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.SceneEventRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	applySceneTypeFromPath(r, &req.SceneType)
	res, err := h.scene.SendEvent(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func (h *Handlers) SceneEventList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	var req service.SceneEventListRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusOK, "参数错误")
		return
	}
	applySceneTypeFromPath(r, &req.SceneType)
	res, err := h.scene.ListEvents(r.Context(), requestAppID(r, req.AppID), req)
	writeServiceResult(w, res, err)
}

func decodeJSON(r *http.Request, out interface{}) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	return dec.Decode(out)
}

func writeServiceResult(w http.ResponseWriter, data interface{}, err error) {
	if err != nil {
		response.Error(w, http.StatusOK, err.Error())
		return
	}
	response.Success(w, data)
}

func bearerToken(r *http.Request) string {
	for _, name := range []string{"authori-zation", "Authorization", "Token"} {
		v := strings.TrimSpace(r.Header.Get(name))
		if v == "" {
			continue
		}
		if strings.HasPrefix(strings.ToLower(v), "bearer ") {
			return strings.TrimSpace(v[7:])
		}
		return v
	}
	return ""
}

func firstHeader(r *http.Request, names ...string) string {
	for _, name := range names {
		if v := strings.TrimSpace(r.Header.Get(name)); v != "" {
			return v
		}
	}
	return ""
}

func decodeShopMessageListRequest(w http.ResponseWriter, r *http.Request) (service.ShopMessageListRequest, bool) {
	var req service.ShopMessageListRequest
	if r.Method == http.MethodPost {
		if err := decodeJSON(r, &req); err != nil {
			response.Error(w, http.StatusOK, "参数错误")
			return req, false
		}
		return req, true
	}
	q := r.URL.Query()
	req.AppID = atoi(q.Get("app_id"))
	req.LiveID = int64(atoi(q.Get("live_id")))
	req.RoomID = q.Get("room_id")
	req.SessionID = q.Get("session_id")
	req.Keyword = q.Get("keyword")
	req.UserID = int64(atoi(q.Get("user_id")))
	req.Page = atoi(q.Get("page"))
	req.PageSize = atoi(q.Get("page_size"))
	req.Limit = atoi(q.Get("limit"))
	req.SensitiveOnly = q.Get("sensitive_only") == "1" || strings.EqualFold(q.Get("sensitive_only"), "true")
	return req, true
}

func requestAppID(r *http.Request, bodyAppID int) int {
	if bodyAppID > 0 {
		return bodyAppID
	}
	for _, name := range []string{"AppId", "appid", "App-ID", "X-App-Id"} {
		if v := atoi(r.Header.Get(name)); v > 0 {
			return v
		}
	}
	return 10001
}

func clientIP(r *http.Request) string {
	for _, name := range []string{"X-Forwarded-For", "X-Real-IP"} {
		if v := strings.TrimSpace(r.Header.Get(name)); v != "" {
			if idx := strings.Index(v, ","); idx >= 0 {
				return strings.TrimSpace(v[:idx])
			}
			return v
		}
	}
	return r.RemoteAddr
}

func applySceneTypeFromPath(r *http.Request, sceneType *string) {
	if sceneType == nil || strings.TrimSpace(*sceneType) != "" {
		return
	}
	path := r.URL.Path
	if strings.Contains(path, "/scene/show/") {
		*sceneType = "show"
		return
	}
	if strings.Contains(path, "/scene/voice/") {
		*sceneType = "voice"
	}
}

func isShadowRequest(r *http.Request) bool {
	return strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Chat-Proxy-Mode")), "shadow")
}

func atoi(raw string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(raw))
	return n
}
