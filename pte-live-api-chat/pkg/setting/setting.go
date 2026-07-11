package setting

import (
	"flag"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type serverSetting struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	Mode string `yaml:"mode"`
}

type imSetting struct {
	WsURL            string `yaml:"wsUrl"`
	HTTPURL          string `yaml:"httpUrl"`
	DeliverPath      string `yaml:"deliverPath"`
	DeliverBackend   string `yaml:"deliverBackend"`
	PulsarServiceURL string `yaml:"pulsarServiceUrl"`
	PulsarTopic      string `yaml:"pulsarTopic"`
	OutboxEnabled    bool   `yaml:"outboxEnabled"`
	OutboxWorkers    int    `yaml:"outboxWorkers"`
	OutboxBatchSize  int    `yaml:"outboxBatchSize"`
	OutboxInterval   int    `yaml:"outboxInterval"`
	OutboxLockTTL    int    `yaml:"outboxLockTtl"`
	OutboxMaxRetries int    `yaml:"outboxMaxRetries"`
}

type sceneSetting struct {
	TimeoutWorkerEnabled bool `yaml:"timeoutWorkerEnabled"`
	TimeoutInterval      int  `yaml:"timeoutInterval"`
	TimeoutBatchSize     int  `yaml:"timeoutBatchSize"`
	MicRequestTTL        int  `yaml:"micRequestTtl"`
	PKInviteTTL          int  `yaml:"pkInviteTtl"`
}

type ClusterConf struct {
	Enabled string `yaml:"enabled"`
	PodHost string `yaml:"podHost"`
}

type MySQLConf struct {
	DSN             string   `yaml:"dsn"`
	WriteDSN        string   `yaml:"writeDsn"`
	ReadDSNs        []string `yaml:"readDsns"`
	MaxOpenConns    int      `yaml:"maxOpenConns"`
	MaxIdleConns    int      `yaml:"maxIdleConns"`
	ConnMaxLifetime int      `yaml:"connMaxLifetime"`
}

type RedisConf struct {
	Addr     string   `yaml:"addr"`
	Addrs    []string `yaml:"addrs"`
	Mode     string   `yaml:"mode"`
	Password string   `yaml:"password"`
	DB       int      `yaml:"db"`
}

type appConfig struct {
	Cluster ClusterConf   `yaml:"cluster"`
	Server  serverSetting `yaml:"server"`
	MySQL   MySQLConf     `yaml:"mysql"`
	Redis   RedisConf     `yaml:"redis"`
	IM      imSetting     `yaml:"im"`
	Scene   sceneSetting  `yaml:"scene"`
}

var (
	Server = serverSetting{
		Port: "11504",
		Mode: "debug",
	}
	IM = imSetting{
		WsURL:            "ws://127.0.0.1:11510/ws",
		HTTPURL:          "http://127.0.0.1:11511",
		DeliverPath:      "/api/chat/deliver",
		DeliverBackend:   "http",
		PulsarTopic:      "persistent://pte_live/live/chat-events",
		OutboxEnabled:    true,
		OutboxWorkers:    2,
		OutboxBatchSize:  20,
		OutboxInterval:   2,
		OutboxLockTTL:    60,
		OutboxMaxRetries: 10,
	}
	Scene = sceneSetting{
		TimeoutWorkerEnabled: true,
		TimeoutInterval:      5,
		TimeoutBatchSize:     100,
		MicRequestTTL:        60,
		PKInviteTTL:          60,
	}
	MySQL = MySQLConf{
		MaxOpenConns:    20,
		MaxIdleConns:    5,
		ConnMaxLifetime: 300,
	}
	Redis = RedisConf{
		Addr: "pte_live_redis:6379",
		Mode: "single",
	}
	Cluster    = ClusterConf{Enabled: "auto"}
	ConfigPath string
)

func Setup() {
	flag.StringVar(&ConfigPath, "c", "", "config file path")
	flag.Parse()

	loadConfigFile()
	normalizeMySQL()

	if v := strings.TrimSpace(os.Getenv("CHAT_API_PORT")); v != "" {
		Server.Port = v
	}
	if v := strings.TrimSpace(os.Getenv("SERVER_PORT")); v != "" {
		Server.Port = v
	}
	if v := strings.TrimSpace(os.Getenv("GIN_MODE")); v != "" {
		Server.Mode = v
	}
	if v := strings.TrimRight(strings.TrimSpace(os.Getenv("IM_WS_URL")), "/"); v != "" {
		IM.WsURL = ensureWSPath(v)
	}
	if v := strings.TrimRight(strings.TrimSpace(os.Getenv("CHAT_IM_WS_URL")), "/"); v != "" {
		IM.WsURL = ensureWSPath(v)
	}
	applyIMEnv()
	normalizeIM()
	applySceneEnv()
	normalizeScene()
	applyMySQLEnv()
	normalizeMySQL()
	normalizeRedis()
}

func ensureWSPath(raw string) string {
	if strings.HasSuffix(strings.ToLower(raw), "/ws") {
		return raw
	}
	return raw + "/ws"
}

func loadConfigFile() {
	path := strings.TrimSpace(ConfigPath)
	if path == "" {
		path = strings.TrimSpace(os.Getenv("CONFIG_FILE"))
	}
	if path == "" {
		for _, candidate := range []string{"conf/app.yaml", "conf/app.yaml.example"} {
			if _, err := os.Stat(candidate); err == nil {
				path = candidate
				break
			}
		}
	}
	if path == "" {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	cfg := appConfig{IM: IM, Scene: Scene, Redis: Redis}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return
	}
	if strings.TrimSpace(cfg.Server.Port) != "" {
		Server = cfg.Server
	}
	if strings.TrimSpace(cfg.Cluster.Enabled) != "" {
		Cluster = cfg.Cluster
	}
	if strings.TrimSpace(cfg.IM.WsURL) != "" {
		IM = cfg.IM
		IM.WsURL = ensureWSPath(IM.WsURL)
	}
	normalizeIM()
	Scene = cfg.Scene
	normalizeScene()
	MySQL = cfg.MySQL
	Redis = cfg.Redis
}

func applyIMEnv() {
	if v := strings.TrimRight(strings.TrimSpace(os.Getenv("IM_HTTP_URL")), "/"); v != "" {
		IM.HTTPURL = v
	}
	if v := strings.TrimRight(strings.TrimSpace(os.Getenv("CHAT_IM_HTTP_URL")), "/"); v != "" {
		IM.HTTPURL = v
	}
	if v := strings.TrimSpace(os.Getenv("CHAT_IM_DELIVER_PATH")); v != "" {
		IM.DeliverPath = v
	}
	if v := strings.TrimSpace(os.Getenv("CHAT_IM_DELIVER_BACKEND")); v != "" {
		IM.DeliverBackend = v
	}
	if v := strings.TrimSpace(os.Getenv("CHAT_PULSAR_SERVICE_URL")); v != "" {
		IM.PulsarServiceURL = v
	}
	if v := strings.TrimSpace(os.Getenv("CHAT_PULSAR_TOPIC")); v != "" {
		IM.PulsarTopic = v
	}
	if v := strings.TrimSpace(os.Getenv("CHAT_OUTBOX_ENABLED")); v != "" {
		IM.OutboxEnabled = envBool(v, IM.OutboxEnabled)
	}
	if v := strings.TrimSpace(os.Getenv("CHAT_OUTBOX_WORKERS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			IM.OutboxWorkers = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("CHAT_OUTBOX_BATCH_SIZE")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			IM.OutboxBatchSize = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("CHAT_OUTBOX_INTERVAL")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			IM.OutboxInterval = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("CHAT_OUTBOX_LOCK_TTL")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			IM.OutboxLockTTL = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("CHAT_OUTBOX_MAX_RETRIES")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			IM.OutboxMaxRetries = n
		}
	}
}

func normalizeIM() {
	IM.WsURL = ensureWSPath(strings.TrimRight(strings.TrimSpace(IM.WsURL), "/"))
	IM.HTTPURL = strings.TrimRight(strings.TrimSpace(IM.HTTPURL), "/")
	if IM.HTTPURL == "" {
		IM.HTTPURL = "http://127.0.0.1:11511"
	}
	IM.DeliverPath = strings.TrimSpace(IM.DeliverPath)
	if IM.DeliverPath == "" {
		IM.DeliverPath = "/api/chat/deliver"
	}
	if !strings.HasPrefix(IM.DeliverPath, "/") {
		IM.DeliverPath = "/" + IM.DeliverPath
	}
	switch strings.ToLower(strings.TrimSpace(IM.DeliverBackend)) {
	case "pulsar":
		IM.DeliverBackend = "pulsar"
	case "both", "dual", "dual-write", "dualwrite":
		IM.DeliverBackend = "both"
	default:
		IM.DeliverBackend = "http"
	}
	IM.PulsarServiceURL = strings.TrimSpace(IM.PulsarServiceURL)
	IM.PulsarTopic = strings.TrimSpace(IM.PulsarTopic)
	if IM.PulsarTopic == "" {
		IM.PulsarTopic = "persistent://pte_live/live/chat-events"
	}
	if IM.OutboxWorkers <= 0 {
		IM.OutboxWorkers = 2
	}
	if IM.OutboxBatchSize <= 0 || IM.OutboxBatchSize > 100 {
		IM.OutboxBatchSize = 20
	}
	if IM.OutboxInterval <= 0 {
		IM.OutboxInterval = 2
	}
	if IM.OutboxLockTTL <= 0 {
		IM.OutboxLockTTL = 60
	}
	if IM.OutboxMaxRetries <= 0 {
		IM.OutboxMaxRetries = 10
	}
}

func applySceneEnv() {
	if v := strings.TrimSpace(os.Getenv("CHAT_SCENE_TIMEOUT_WORKER_ENABLED")); v != "" {
		Scene.TimeoutWorkerEnabled = envBool(v, Scene.TimeoutWorkerEnabled)
	}
	if v := strings.TrimSpace(os.Getenv("CHAT_SCENE_TIMEOUT_INTERVAL")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			Scene.TimeoutInterval = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("CHAT_SCENE_TIMEOUT_BATCH_SIZE")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			Scene.TimeoutBatchSize = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("CHAT_SCENE_MIC_REQUEST_TTL")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			Scene.MicRequestTTL = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("CHAT_SCENE_PK_INVITE_TTL")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			Scene.PKInviteTTL = n
		}
	}
}

func normalizeScene() {
	if Scene.TimeoutInterval <= 0 {
		Scene.TimeoutInterval = 5
	}
	if Scene.TimeoutBatchSize <= 0 || Scene.TimeoutBatchSize > 500 {
		Scene.TimeoutBatchSize = 100
	}
	if Scene.MicRequestTTL <= 0 {
		Scene.MicRequestTTL = 60
	}
	if Scene.PKInviteTTL <= 0 {
		Scene.PKInviteTTL = 60
	}
}

func applyMySQLEnv() {
	if v := strings.TrimSpace(os.Getenv("MYSQL_WRITE_DSN")); v != "" {
		MySQL.WriteDSN = v
		MySQL.DSN = v
	}
	if v := strings.TrimSpace(os.Getenv("MYSQL_DSN")); v != "" && strings.TrimSpace(MySQL.WriteDSN) == "" {
		MySQL.WriteDSN = v
		MySQL.DSN = v
	}
	if v := strings.TrimSpace(os.Getenv("MYSQL_READ_DSNS")); v != "" {
		MySQL.ReadDSNs = splitCSV(v)
	}
	if v := strings.TrimSpace(os.Getenv("MYSQL_MAX_OPEN_CONNS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			MySQL.MaxOpenConns = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("MYSQL_MAX_IDLE_CONNS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			MySQL.MaxIdleConns = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("MYSQL_CONN_MAX_LIFETIME")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			MySQL.ConnMaxLifetime = n
		}
	}
}

func normalizeMySQL() {
	MySQL.WriteDSN = strings.TrimSpace(MySQL.WriteDSN)
	MySQL.DSN = strings.TrimSpace(MySQL.DSN)
	if MySQL.WriteDSN == "" {
		MySQL.WriteDSN = MySQL.DSN
	}
	if MySQL.WriteDSN != "" {
		MySQL.DSN = MySQL.WriteDSN
	}
	if len(MySQL.ReadDSNs) == 0 && MySQL.WriteDSN != "" {
		MySQL.ReadDSNs = []string{MySQL.WriteDSN}
	}
	MySQL.ReadDSNs = trimNonEmpty(MySQL.ReadDSNs)
	if MySQL.MaxOpenConns <= 0 {
		MySQL.MaxOpenConns = 20
	}
	if MySQL.MaxIdleConns <= 0 {
		MySQL.MaxIdleConns = 5
	}
	if MySQL.ConnMaxLifetime <= 0 {
		MySQL.ConnMaxLifetime = 300
	}
}

func MySQLConfigured() bool {
	return strings.TrimSpace(MySQL.WriteDSN) != "" || strings.TrimSpace(MySQL.DSN) != ""
}

func (m *MySQLConf) HasReadReplica() bool {
	write := strings.TrimSpace(m.WriteDSN)
	if write == "" {
		write = strings.TrimSpace(m.DSN)
	}
	for _, dsn := range m.ReadDSNs {
		if r := strings.TrimSpace(dsn); r != "" && r != write {
			return true
		}
	}
	return false
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	return trimNonEmpty(parts)
}

func trimNonEmpty(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if item := strings.TrimSpace(v); item != "" {
			out = append(out, item)
		}
	}
	return out
}

func envBool(v string, def bool) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}
