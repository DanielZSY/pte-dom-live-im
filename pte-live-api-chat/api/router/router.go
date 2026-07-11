package router

import (
	"net/http"

	"pte_live_api_chat/internal/handler"
	iredis "pte_live_api_chat/internal/redis"
	"pte_live_api_chat/pkg/ratelimit"
	"pte_live_api_chat/pkg/response"
)

func New(h *handler.Handlers) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", h.Ping)
	mux.HandleFunc("/healthz", h.Healthz)
	mux.HandleFunc("/readyz", h.Readyz)
	mux.HandleFunc("/metrics", h.PrometheusMetrics)
	mux.HandleFunc("/api/internal/ops/metrics", h.OpsMetrics)
	mux.HandleFunc("/api/v1/im/usersig", h.IMUserSig)
	mux.HandleFunc("/api/internal/im/usersig/verify", h.IMUserSigVerify)
	mux.HandleFunc("/api/v1/chat/conversation/open-single", h.ChatOpenSingle)
	mux.HandleFunc("/api/v1/chat/conversation/create-group", h.ChatCreateGroup)
	mux.HandleFunc("/api/v1/chat/conversation/list", h.ChatConversationList)
	mux.HandleFunc("/api/v1/chat/conversation/detail", h.ChatConversationDetail)
	mux.HandleFunc("/api/v1/chat/conversation/read", h.ChatConversationRead)
	mux.HandleFunc("/api/v1/chat/member/list", h.ChatMemberList)
	mux.HandleFunc("/api/v1/chat/member/add", h.ChatMemberAdd)
	mux.HandleFunc("/api/v1/chat/member/remove", h.ChatMemberRemove)
	mux.HandleFunc("/api/v1/chat/message/send", h.ChatMessageSend)
	mux.HandleFunc("/api/v1/chat/message/history", h.ChatMessageHistory)
	mux.HandleFunc("/api/v1/chat/message/sync", h.ChatMessageSync)
	mux.HandleFunc("/api/v1/chat/message/ack", h.ChatMessageAck)
	mux.HandleFunc("/api/v1/chat/message/recall", h.ChatMessageRecall)
	mux.HandleFunc("/api/v1/chat/message/delete", h.ChatMessageDelete)
	mux.HandleFunc("/api/v1/scene/shop/message/send", h.ShopMessageSend)
	mux.HandleFunc("/api/v1/scene/shop/message/recent", h.ShopMessageList)
	mux.HandleFunc("/api/v1/scene/shop/message/history", h.ShopMessageList)
	mux.HandleFunc("/api/v1/scene/shop/message/audit/list", h.ShopMessageList)
	mux.HandleFunc("/api/v1/scene/shop/message/audit/count", h.ShopMessagePendingCount)
	mux.HandleFunc("/api/v1/scene/shop/message/audit/submit", h.ShopMessageAuditSubmit)
	registerSceneRoutes(mux, "/api/v1/scene", h)
	registerSceneRoutes(mux, "/api/v1/scene/show", h)
	registerSceneRoutes(mux, "/api/v1/scene/voice", h)
	return withCORS(ratelimit.Middleware("api-chat", iredis.NewClient(), notFound(mux)))
}

func registerSceneRoutes(mux *http.ServeMux, prefix string, h *handler.Handlers) {
	mux.HandleFunc(prefix+"/room/open", h.SceneRoomOpen)
	mux.HandleFunc(prefix+"/room/close", h.SceneRoomClose)
	mux.HandleFunc(prefix+"/room/list", h.SceneRoomList)
	mux.HandleFunc(prefix+"/room/detail", h.SceneRoomDetail)
	mux.HandleFunc(prefix+"/room/enter", h.SceneRoomEnter)
	mux.HandleFunc(prefix+"/room/leave", h.SceneRoomLeave)
	mux.HandleFunc(prefix+"/member/list", h.SceneMemberList)
	mux.HandleFunc(prefix+"/seat/action", h.SceneSeatAction)
	mux.HandleFunc(prefix+"/moderation/action", h.SceneModerationAction)
	mux.HandleFunc(prefix+"/pk/invite", h.ScenePKInvite)
	mux.HandleFunc(prefix+"/pk/start", h.ScenePKStart)
	mux.HandleFunc(prefix+"/pk/action", h.ScenePKAction)
	mux.HandleFunc(prefix+"/pk/end", h.ScenePKEnd)
	mux.HandleFunc(prefix+"/event/send", h.SceneEventSend)
	mux.HandleFunc(prefix+"/event/list", h.SceneEventList)
}

func notFound(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rr, r)
		if rr.wrote {
			return
		}
		response.Error(w, http.StatusNotFound, "接口不存在")
	})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, authori-zation, Authorization, Token, AppId")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.wrote = true
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	r.wrote = true
	return r.ResponseWriter.Write(b)
}
