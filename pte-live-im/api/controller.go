package api

import (
	"encoding/json"
	zhongwen "github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"pte_live_im/define/retcode"
	"pte_live_im/pkg/appid"
	"gopkg.in/go-playground/validator.v9"
	zh2 "gopkg.in/go-playground/validator.v9/translations/zh"
	"io"
	"net/http"
)

type RetData struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func ConnRender(conn *websocket.Conn, data interface{}) (err error) {
	err = conn.WriteJSON(RetData{
		Code: retcode.SUCCESS,
		Msg:  "success",
		Data: data,
	})

	return
}

func ConnRenderError(conn *websocket.Conn, code int, msg string) error {
	return conn.WriteJSON(RetData{
		Code: code,
		Msg:  msg,
		Data: []string{},
	})
}

func Render(w http.ResponseWriter, code int, msg string, data interface{}) (str string) {
	var retData RetData

	retData.Code = code
	retData.Msg = msg
	retData.Data = data

	retJson, _ := json.Marshal(retData)
	str = string(retJson)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = io.WriteString(w, str)
	return
}

// AppID 从 Header 解析租户 appId（兼容旧 SystemId）
func AppID(r *http.Request) string {
	return appid.FromHTTP(r)
}

func Validate(inputData interface{}) error {
	validate := validator.New()
	zh := zhongwen.New()
	uni := ut.New(zh, zh)
	trans, _ := uni.GetTranslator("zh")

	_ = zh2.RegisterDefaultTranslations(validate, trans)

	err := validate.Struct(inputData)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			return errors.New(err.Translate(trans))
		}
	}

	return nil
}
