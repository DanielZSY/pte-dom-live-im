package routers

import (
	"net/http"

	adminConnection "pte_live_im/api/admin/connection"
	"pte_live_im/api/bind2group"
	"pte_live_im/api/chat/deliver"
	"pte_live_im/api/closeclient"
	"pte_live_im/api/getonlinelist"
	"pte_live_im/api/live/disconnectuser"
	"pte_live_im/api/live/giftcount"
	"pte_live_im/api/live/giftlist"
	"pte_live_im/api/live/linkmiclist"
	"pte_live_im/api/live/mutelist"
	"pte_live_im/api/live/roominfo"
	"pte_live_im/api/live/sendmessage"
	"pte_live_im/api/live/sessioncountsreset"
	"pte_live_im/api/register"
	"pte_live_im/api/send2client"
	"pte_live_im/api/send2clients"
	"pte_live_im/api/send2group"
	"pte_live_im/pkg/security"
	"pte_live_im/servers"
)

func apiHandler(h http.HandlerFunc) http.HandlerFunc {
	return CORSMiddleware(security.HTTPMiddleware("api", h))
}

func Init() {
	registerHandler := &register.Controller{}
	sendToClientHandler := &send2client.Controller{}
	sendToClientsHandler := &send2clients.Controller{}
	sendToGroupHandler := &send2group.Controller{}
	bindToGroupHandler := &bind2group.Controller{}
	getGroupListHandler := &getonlinelist.Controller{}
	closeClientHandler := &closeclient.Controller{}
	chatDeliverHandler := &deliver.Controller{}
	liveSendMessageHandler := &sendmessage.Controller{}
	liveLinkMicListHandler := &linkmiclist.Controller{}
	liveMuteListHandler := &mutelist.Controller{}
	liveGiftListHandler := &giftlist.Controller{}
	liveGiftCountHandler := &giftcount.Controller{}
	liveRoomInfoHandler := &roominfo.Controller{}
	liveSessionCountsResetHandler := &sessioncountsreset.Controller{}
	liveDisconnectUserHandler := &disconnectuser.Controller{}
	adminConnectionHandler := &adminConnection.Controller{}
	adminLocalConnectionHandler := &adminConnection.Controller{Local: true}

	http.HandleFunc("/ping", CORSMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.Method == http.MethodOptions {
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	}))

	http.HandleFunc("/api/register", apiHandler(registerHandler.Run))
	http.HandleFunc("/api/send_to_client", apiHandler(AccessTokenMiddleware(sendToClientHandler.Run)))
	http.HandleFunc("/api/send_to_clients", apiHandler(AccessTokenMiddleware(sendToClientsHandler.Run)))
	http.HandleFunc("/api/send_to_group", apiHandler(AccessTokenMiddleware(sendToGroupHandler.Run)))
	http.HandleFunc("/api/bind_to_group", apiHandler(AccessTokenMiddleware(bindToGroupHandler.Run)))
	http.HandleFunc("/api/get_online_list", apiHandler(AccessTokenMiddleware(getGroupListHandler.Run)))
	http.HandleFunc("/api/close_client", apiHandler(AccessTokenMiddleware(closeClientHandler.Run)))
	http.HandleFunc("/api/chat/deliver", apiHandler(AccessTokenMiddleware(chatDeliverHandler.Run)))
	http.HandleFunc("/api/admin/connection/list", apiHandler(AccessTokenMiddleware(adminConnectionHandler.List)))
	http.HandleFunc("/api/admin/connection/local-list", apiHandler(AccessTokenMiddleware(adminLocalConnectionHandler.List)))
	http.HandleFunc("/api/admin/connection/kick", apiHandler(AccessTokenMiddleware(adminConnectionHandler.Kick)))

	http.HandleFunc("/api/live/send_message", apiHandler(AccessTokenMiddleware(liveSendMessageHandler.Run)))
	http.HandleFunc("/api/live/linkmic_list", apiHandler(AccessTokenMiddleware(liveLinkMicListHandler.Run)))
	http.HandleFunc("/api/live/mute_list", apiHandler(AccessTokenMiddleware(liveMuteListHandler.Run)))
	http.HandleFunc("/api/live/gift_list", apiHandler(AccessTokenMiddleware(liveGiftListHandler.Run)))
	http.HandleFunc("/api/live/gift_count", apiHandler(AccessTokenMiddleware(liveGiftCountHandler.Run)))
	http.HandleFunc("/api/live/room_info", apiHandler(AccessTokenMiddleware(liveRoomInfoHandler.Run)))
	http.HandleFunc("/api/live/session_counts_reset", apiHandler(AccessTokenMiddleware(liveSessionCountsResetHandler.Run)))
	http.HandleFunc("/api/live/disconnect_user", apiHandler(AccessTokenMiddleware(liveDisconnectUserHandler.Run)))

	servers.StartWebSocket()

	go servers.WriteMessage()
}
