package servers

import (
	"encoding/json"
	"errors"
	"pte_live_im/define"
	"pte_live_im/pkg/etcd"
	"pte_live_im/tools/util"
	"sync"
	"time"
)

type accountInfo struct {
	SystemId     string `json:"systemId"`
	RegisterTime int64  `json:"registerTime"`
}

var SystemMap sync.Map

func ValidateAppID(appId string) error {
	if len(appId) == 0 {
		return errors.New("appId不能为空")
	}

	if util.IsCluster() {
		resp, err := etcd.Get(define.ETCD_PREFIX_ACCOUNT_INFO + appId)
		if err != nil {
			return errors.New("etcd服务器错误")
		}
		if resp.Count == 0 {
			return errors.New("appId无效")
		}
	} else if _, ok := SystemMap.Load(appId); !ok {
		return errors.New("appId无效")
	}

	return nil
}

// ValidateSystemID 兼容旧调用
func ValidateSystemID(appId string) error {
	return ValidateAppID(appId)
}

func Register(appId string) (err error) {
	if len(appId) == 0 {
		return errors.New("appId不能为空")
	}

	accountInfo := accountInfo{
		SystemId:     appId,
		RegisterTime: time.Now().Unix(),
	}

	if util.IsCluster() {
		resp, err := etcd.Get(define.ETCD_PREFIX_ACCOUNT_INFO + appId)
		if err != nil {
			return err
		}

		if resp.Count > 0 {
			return nil
		}

		jsonBytes, _ := json.Marshal(accountInfo)

		err = etcd.Put(define.ETCD_PREFIX_ACCOUNT_INFO+appId, string(jsonBytes))
		if err != nil {
			panic(err)
			return err
		}
	} else {
		if _, ok := SystemMap.Load(appId); ok {
			return nil
		}

		SystemMap.Store(appId, accountInfo)
	}

	return nil
}
