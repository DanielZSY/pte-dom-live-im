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

type authSetting struct {
	RequestHeader string `yaml:"requestHeader"`
	AdminUsername string `yaml:"adminUsername"`
	AdminPassword string `yaml:"adminPassword"`
	TokenSecret   string `yaml:"tokenSecret"`
	TokenTTL      int    `yaml:"tokenTtl"`
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

type IMConf struct {
	BaseURLs []string `yaml:"baseUrls"`
}

type appConfig struct {
	Cluster ClusterConf   `yaml:"cluster"`
	Server  serverSetting `yaml:"server"`
	Auth    authSetting   `yaml:"auth"`
	MySQL   MySQLConf     `yaml:"mysql"`
	Redis   RedisConf     `yaml:"redis"`
	IM      IMConf        `yaml:"im"`
}

var (
	Server = serverSetting{
		Port: "11505",
		Mode: "debug",
	}
	Auth = authSetting{
		RequestHeader: "authori-zation",
		AdminUsername: "imadmin",
		AdminPassword: "pte123321",
		TokenSecret:   "pte-live-api-chat-admin-dev-secret",
		TokenTTL:      86400,
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
	IM = IMConf{
		BaseURLs: []string{"http://pte_live_im:11511"},
	}
	Cluster    = ClusterConf{Enabled: "auto"}
	ConfigPath string
)

func Setup() {
	flag.StringVar(&ConfigPath, "c", "", "config file path")
	flag.Parse()

	loadConfigFile()
	normalizeMySQL()

	if v := strings.TrimSpace(os.Getenv("CHAT_ADMIN_PORT")); v != "" {
		Server.Port = v
	}
	if v := strings.TrimSpace(os.Getenv("SERVER_PORT")); v != "" {
		Server.Port = v
	}
	if v := strings.TrimSpace(os.Getenv("GIN_MODE")); v != "" {
		Server.Mode = v
	}
	if v := strings.TrimSpace(os.Getenv("AUTH_REQUEST_HEADER")); v != "" {
		Auth.RequestHeader = v
	}
	if v := strings.TrimSpace(os.Getenv("CHAT_ADMIN_USERNAME")); v != "" {
		Auth.AdminUsername = v
	}
	if v := strings.TrimSpace(os.Getenv("CHAT_ADMIN_PASSWORD")); v != "" {
		Auth.AdminPassword = v
	}
	if v := strings.TrimSpace(os.Getenv("CHAT_ADMIN_TOKEN_SECRET")); v != "" {
		Auth.TokenSecret = v
	}
	if v := strings.TrimSpace(os.Getenv("CHAT_ADMIN_TOKEN_TTL")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			Auth.TokenTTL = n
		}
	}
	normalizeAuth()
	applyMySQLEnv()
	normalizeRedis()
	applyIMEnv()
	normalizeMySQL()
	normalizeIM()
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
	cfg := appConfig{Auth: Auth, MySQL: MySQL, Redis: Redis, IM: IM}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return
	}
	if strings.TrimSpace(cfg.Server.Port) != "" {
		Server = cfg.Server
	}
	if strings.TrimSpace(cfg.Auth.RequestHeader) != "" {
		Auth = cfg.Auth
	}
	if strings.TrimSpace(cfg.Cluster.Enabled) != "" {
		Cluster = cfg.Cluster
	}
	MySQL = cfg.MySQL
	Redis = cfg.Redis
	if len(cfg.IM.BaseURLs) > 0 {
		IM = cfg.IM
	}
	normalizeAuth()
}

func normalizeAuth() {
	Auth.RequestHeader = strings.TrimSpace(Auth.RequestHeader)
	if Auth.RequestHeader == "" {
		Auth.RequestHeader = "authori-zation"
	}
	if strings.TrimSpace(Auth.AdminUsername) == "" {
		Auth.AdminUsername = "imadmin"
	}
	if strings.TrimSpace(Auth.AdminPassword) == "" {
		Auth.AdminPassword = "pte123321"
	}
	if strings.TrimSpace(Auth.TokenSecret) == "" {
		Auth.TokenSecret = "pte-live-api-chat-admin-dev-secret"
	}
	if Auth.TokenTTL <= 0 {
		Auth.TokenTTL = 86400
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

func applyIMEnv() {
	if v := strings.TrimSpace(os.Getenv("IM_BASE_URLS")); v != "" {
		IM.BaseURLs = splitCSV(v)
	}
	if v := strings.TrimSpace(os.Getenv("IM_BASE_URL")); v != "" && len(IM.BaseURLs) == 0 {
		IM.BaseURLs = []string{v}
	}
}

func normalizeIM() {
	IM.BaseURLs = trimNonEmpty(IM.BaseURLs)
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
