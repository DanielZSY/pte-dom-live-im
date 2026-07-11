package setting

import (
	"flag"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

type CORSConf struct {
	Enabled          bool     `yaml:"enabled"`
	AllowOrigins     []string `yaml:"allowOrigins"`
	AllowCredentials bool     `yaml:"allowCredentials"`
	AllowMethods     string   `yaml:"allowMethods"`
	AllowHeaders     string   `yaml:"allowHeaders"`
	ExposeHeaders    string   `yaml:"exposeHeaders"`
}

type commonConf struct {
	HttpPort      string `yaml:"httpPort"`
	WebSocketPort string `yaml:"webSocketPort"`
	RPCPort       string `yaml:"rpcPort"`
	Cluster       bool   `yaml:"cluster"`
	CryptoKey     string `yaml:"cryptoKey"`
}

type redisConf struct {
	Addr     string   `yaml:"addr"`
	Addrs    []string `yaml:"addrs"`
	Mode     string   `yaml:"mode"`
	Password string   `yaml:"password"`
	DB       int      `yaml:"db"`
}

type liveConf struct {
	SkipTokenValidate  bool   `yaml:"skipTokenValidate"`
	TokenKeyPrefix     string `yaml:"tokenKeyPrefix"`
	QueueWorkers       int    `yaml:"queueWorkers"`
	JwtSalt            string `yaml:"jwtSalt"`
	JwtType            string `yaml:"jwtType"`
	JwtLeeway          int    `yaml:"jwtLeeway"`
	JwtBlacklistPrefix string `yaml:"jwtBlacklistPrefix"`
}

type authConf struct {
	UserSigVerifyURL     string `yaml:"userSigVerifyUrl"`
	UserSigVerifyTimeout int    `yaml:"userSigVerifyTimeout"`
	LegacyTokenEnabled   bool   `yaml:"legacyTokenEnabled"`
}

// QueueConf IM 消息队列后端。默认使用 Pulsar；Redis 只作为兼容可选项。
type QueueConf struct {
	Backend     string `yaml:"backend"`     // pulsar | redis | both
	ConsumeFrom string `yaml:"consumeFrom"` // pulsar | redis（backend=both 时生效）
}

// PulsarConf Pulsar 连接（IM 事件管道）。
type PulsarConf struct {
	Enabled          bool   `yaml:"enabled"`
	ServiceURL       string `yaml:"serviceURL"`
	Topic            string `yaml:"topic"`
	Subscription     string `yaml:"subscription"`
	ChatTopic        string `yaml:"chatTopic"`
	ChatSubscription string `yaml:"chatSubscription"`
	ChatWorkers      int    `yaml:"chatWorkers"`
}

var CommonSetting = &commonConf{}
var CORSSetting = &CORSConf{}
var RedisSetting = &redisConf{}
var QueueSetting = &QueueConf{
	Backend:     "pulsar",
	ConsumeFrom: "pulsar",
}
var PulsarSetting = &PulsarConf{
	Enabled:          true,
	Topic:            "persistent://pte_live/live/im-events",
	Subscription:     "pte-live-im",
	ChatTopic:        "persistent://pte_live/live/chat-events",
	ChatSubscription: "pte-live-im-chat",
	ChatWorkers:      2,
}
var LiveSetting = &liveConf{
	SkipTokenValidate:  false,
	TokenKeyPrefix:     "live:token:",
	QueueWorkers:       2,
	JwtType:            "user",
	JwtLeeway:          60,
	JwtBlacklistPrefix: "auth:user:blacklist:",
}
var AuthSetting = &authConf{
	UserSigVerifyTimeout: 3,
	LegacyTokenEnabled:   false,
}

type etcdConf struct {
	Endpoints []string `yaml:"endpoints"`
}

var EtcdSetting = &etcdConf{}

type global struct {
	LocalHost      string //本机内网IP
	ServerList     map[string]string
	ServerListLock sync.RWMutex
}

var GlobalSetting = &global{}

type appConfig struct {
	Cluster clusterConf `yaml:"cluster"`
	Common  commonConf  `yaml:"common"`
	Etcd    etcdConf    `yaml:"etcd"`
	Redis   redisConf   `yaml:"redis"`
	Queue   QueueConf   `yaml:"queue"`
	Pulsar  PulsarConf  `yaml:"pulsar"`
	Live    liveConf    `yaml:"live"`
	Auth    *authConf   `yaml:"auth"`
	CORS    CORSConf    `yaml:"cors"`
}

func Setup() {
	configFile := flag.String("c", "conf/app.yaml", "-c conf/app.yaml")
	flag.Parse()

	data, err := os.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("setting.Setup, fail to read %q: %v (提示: 本地 cp conf/app.yaml.example conf/app.yaml；Docker 挂载 release/config/*/app.yaml)", *configFile, err)
	}

	var cfg appConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("setting.Setup, fail to parse %q: %v", *configFile, err)
	}

	clusterCfg = cfg.Cluster
	*CommonSetting = cfg.Common
	*CORSSetting = cfg.CORS
	*EtcdSetting = cfg.Etcd
	if cfg.Redis.Addr != "" {
		*RedisSetting = cfg.Redis
	}
	if strings.TrimSpace(cfg.Queue.Backend) != "" || strings.TrimSpace(cfg.Queue.ConsumeFrom) != "" {
		*QueueSetting = cfg.Queue
	}
	if cfg.Pulsar.Enabled || strings.TrimSpace(cfg.Pulsar.ServiceURL) != "" {
		*PulsarSetting = cfg.Pulsar
	}
	if cfg.Live.QueueWorkers > 0 || cfg.Live.TokenKeyPrefix != "" {
		*LiveSetting = cfg.Live
	}
	if cfg.Auth != nil {
		*AuthSetting = *cfg.Auth
	}
	applyQueueEnv()
	applyPulsarEnv()
	applyAuthEnv()
	applyCorsEnv()
	normalizeCORS()
	normalizeRedisAddrs()
	normalizeCluster()
	normalizeQueue()
	normalizePulsar()
	normalizeAuth()
	if LiveSetting.QueueWorkers <= 0 {
		LiveSetting.QueueWorkers = 2
	}
	if LiveSetting.TokenKeyPrefix == "" {
		LiveSetting.TokenKeyPrefix = "live:token:"
	}
	if LiveSetting.JwtType == "" {
		LiveSetting.JwtType = "user"
	}
	if LiveSetting.JwtLeeway <= 0 {
		LiveSetting.JwtLeeway = 60
	}
	if LiveSetting.JwtBlacklistPrefix == "" {
		LiveSetting.JwtBlacklistPrefix = "auth:user:blacklist:"
	}

	GlobalSetting = &global{
		LocalHost:  resolveLocalHost(),
		ServerList: make(map[string]string),
	}
}

func applyAuthEnv() {
	if v := strings.TrimSpace(os.Getenv("IM_USERSIG_VERIFY_URL")); v != "" {
		AuthSetting.UserSigVerifyURL = v
	}
	if v := strings.TrimSpace(os.Getenv("IM_USERSIG_VERIFY_TIMEOUT")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			AuthSetting.UserSigVerifyTimeout = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("IM_LEGACY_TOKEN_ENABLED")); v != "" {
		AuthSetting.LegacyTokenEnabled = envBool(v, AuthSetting.LegacyTokenEnabled)
	}
}

func applyQueueEnv() {
	if v := strings.TrimSpace(os.Getenv("QUEUE_BACKEND")); v != "" {
		QueueSetting.Backend = v
	}
	if v := strings.TrimSpace(os.Getenv("QUEUE_CONSUME_FROM")); v != "" {
		QueueSetting.ConsumeFrom = v
	}
}

func applyPulsarEnv() {
	if v := strings.TrimSpace(os.Getenv("PULSAR_SERVICE_URL")); v != "" {
		PulsarSetting.ServiceURL = v
	}
	if v := strings.TrimSpace(os.Getenv("PULSAR_TOPIC")); v != "" {
		PulsarSetting.Topic = v
	}
	if v := strings.TrimSpace(os.Getenv("PULSAR_SUBSCRIPTION")); v != "" {
		PulsarSetting.Subscription = v
	}
	if v := strings.TrimSpace(os.Getenv("CHAT_PULSAR_TOPIC")); v != "" {
		PulsarSetting.ChatTopic = v
	}
	if v := strings.TrimSpace(os.Getenv("CHAT_PULSAR_SUBSCRIPTION")); v != "" {
		PulsarSetting.ChatSubscription = v
	}
	if v := strings.TrimSpace(os.Getenv("CHAT_PULSAR_WORKERS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			PulsarSetting.ChatWorkers = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("PULSAR_ENABLED")); v != "" {
		PulsarSetting.Enabled = envBool(v, PulsarSetting.Enabled)
	}
}

func normalizeQueue() {
	switch strings.ToLower(strings.TrimSpace(QueueSetting.Backend)) {
	case "pulsar":
		QueueSetting.Backend = "pulsar"
	case "both", "dual", "dual-write", "dualwrite":
		QueueSetting.Backend = "both"
	default:
		QueueSetting.Backend = "pulsar"
	}
	switch strings.ToLower(strings.TrimSpace(QueueSetting.ConsumeFrom)) {
	case "redis":
		QueueSetting.ConsumeFrom = "redis"
	case "pulsar":
		QueueSetting.ConsumeFrom = "pulsar"
	default:
		QueueSetting.ConsumeFrom = "pulsar"
	}
}

func normalizePulsar() {
	if strings.TrimSpace(PulsarSetting.Topic) == "" {
		PulsarSetting.Topic = "persistent://pte_live/live/im-events"
	}
	if strings.TrimSpace(PulsarSetting.Subscription) == "" {
		PulsarSetting.Subscription = "pte-live-im"
	}
	if strings.TrimSpace(PulsarSetting.ChatTopic) == "" {
		PulsarSetting.ChatTopic = "persistent://pte_live/live/chat-events"
	}
	if strings.TrimSpace(PulsarSetting.ChatSubscription) == "" {
		PulsarSetting.ChatSubscription = "pte-live-im-chat"
	}
	if PulsarSetting.ChatWorkers <= 0 {
		PulsarSetting.ChatWorkers = 2
	}
	if QueueSetting.Backend == "pulsar" || QueueSetting.Backend == "both" {
		PulsarSetting.Enabled = true
	}
}

func normalizeAuth() {
	if AuthSetting.UserSigVerifyTimeout <= 0 {
		AuthSetting.UserSigVerifyTimeout = 3
	}
	if AuthSetting.UserSigVerifyTimeout > 10 {
		AuthSetting.UserSigVerifyTimeout = 10
	}
}

func applyCorsEnv() {
	if v := strings.TrimSpace(os.Getenv("CORS_ALLOW_ORIGINS")); v != "" {
		parts := strings.Split(v, ",")
		origins := make([]string, 0, len(parts))
		for _, p := range parts {
			if o := strings.TrimSpace(p); o != "" {
				origins = append(origins, o)
			}
		}
		if len(origins) > 0 {
			CORSSetting.AllowOrigins = origins
		}
	}
	if v := strings.TrimSpace(os.Getenv("CORS_ENABLED")); v != "" {
		CORSSetting.Enabled = envBool(v, CORSSetting.Enabled)
	}
}

func normalizeCORS() {
	if !CORSSetting.Enabled {
		return
	}
	if CORSSetting.AllowMethods == "" {
		CORSSetting.AllowMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	}
	if CORSSetting.AllowHeaders == "" {
		CORSSetting.AllowHeaders = "Content-Type, X-Requested-With, app_id, Accept, Origin, Cache-Control, Pragma, authori-zation, AppID, App-Id, appId, AppId, User-Id, userId"
	}
	if CORSSetting.ExposeHeaders == "" {
		CORSSetting.ExposeHeaders = "authori-zation"
	}
}

func envBool(v string, def bool) bool {
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}

func Default() {
	CommonSetting = &commonConf{
		HttpPort:      "11511",
		WebSocketPort: "11510",
		RPCPort:       "11512",
		Cluster:       false,
		CryptoKey:     "Adba723b7fe06819",
	}

	EtcdSetting = &etcdConf{}

	RedisSetting = &redisConf{}
	LiveSetting = &liveConf{
		SkipTokenValidate:  true,
		TokenKeyPrefix:     "live:token:",
		QueueWorkers:       2,
		JwtType:            "user",
		JwtLeeway:          60,
		JwtBlacklistPrefix: "auth:user:blacklist:",
	}
	AuthSetting = &authConf{
		UserSigVerifyTimeout: 3,
		LegacyTokenEnabled:   false,
	}

	GlobalSetting = &global{
		LocalHost:  resolveLocalHost(),
		ServerList: make(map[string]string),
	}
}

// 获取本机内网IP
func resolveLocalHost() string {
	if v := strings.TrimSpace(os.Getenv("IM_LOCAL_HOST")); v != "" {
		return v
	}
	return getIntranetIp()
}

func getIntranetIp() string {
	addrs, _ := net.InterfaceAddrs()

	for _, addr := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}

		}
	}

	return ""
}
