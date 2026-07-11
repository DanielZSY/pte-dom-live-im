package routers

import (
	"net/http"

	"pte_live_im/api"
	"pte_live_im/define/retcode"
	"pte_live_im/pkg/appid"
	"pte_live_im/servers"
)

func AccessTokenMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		appId := appid.FromHTTP(r)
		if err := servers.ValidateAppID(appId); err != nil {
			code := retcode.FAIL
			switch err.Error() {
			case "appId无效", "appId不能为空":
				code = retcode.APP_ID_ERROR
			}
			api.Render(w, code, err.Error(), []string{})
			return
		}

		next.ServeHTTP(w, r)
	})
}
