package util

import (
	"errors"
	uuid "github.com/satori/go.uuid"
	"pte_live_im/pkg/setting"
	"pte_live_im/tools/crypto"
	"strings"
)

//GenUUID 生成uuid
func GenUUID() string {
	uuidFunc := uuid.NewV4()
	uuidStr := uuidFunc.String()
	uuidStr = strings.Replace(uuidStr, "-", "", -1)
	uuidByt := []rune(uuidStr)
	return string(uuidByt[8:24])
}

// GenClientId 对称加密 本机地址+会话ID，用于集群路由且保证每连接唯一
func GenClientId() string {
	raw := []byte(setting.GlobalSetting.LocalHost + ":" + setting.CommonSetting.RPCPort + ":" + GenUUID())
	str, err := crypto.Encrypt(raw, []byte(setting.CommonSetting.CryptoKey))
	if err != nil {
		panic(err)
	}

	return str
}

// ParseClientAddr 解析解密后的 clientId 载荷：host:port 或 host:port:sessionId
func ParseClientAddr(decrypted string) (host string, port string, err error) {
	if decrypted == "" {
		return "", "", errors.New("解析地址错误")
	}
	parts := strings.Split(decrypted, ":")
	if len(parts) < 2 {
		return "", "", errors.New("解析地址错误")
	}
	if len(parts) == 2 {
		return parts[0], parts[1], nil
	}
	port = parts[len(parts)-2]
	host = strings.Join(parts[:len(parts)-2], ":")
	return host, port, nil
}

// ParseRedisAddrValue 兼容旧格式 host:port
func ParseRedisAddrValue(redisValue string) (host string, port string, err error) {
	return ParseClientAddr(redisValue)
}

//判断地址是否为本机
func IsAddrLocal(host string, port string) bool {
	return host == setting.GlobalSetting.LocalHost && port == setting.CommonSetting.RPCPort
}

//是否集群
func IsCluster() bool {
	return setting.CommonSetting.Cluster
}

//获取client key地址信息
func GetAddrInfoAndIsLocal(clientId string) (addr string, host string, port string, isLocal bool, err error) {
	//解密ClientId
	addr, err = crypto.Decrypt(clientId, []byte(setting.CommonSetting.CryptoKey))
	if err != nil {
		return
	}

	host, port, err = ParseClientAddr(addr)
	if err != nil {
		return
	}

	isLocal = IsAddrLocal(host, port)
	return
}

func GenGroupKey(systemId, groupName string) string {
	return systemId + ":" + groupName
}
