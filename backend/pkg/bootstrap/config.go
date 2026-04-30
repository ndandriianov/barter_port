package bootstrap

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DB struct {
		User           string `mapstructure:"user"`
		Password       string `mapstructure:"password"`
		Host           string `mapstructure:"host"`
		Port           string `mapstructure:"port"`
		Name           string `mapstructure:"name"`
		MigrationsPath string `mapstructure:"migrations_path"`
	} `mapstructure:"db"`

	Frontend struct {
		URL string `mapstructure:"url"`
	} `mapstructure:"frontend"`

	Mailer struct {
		Host               string `mapstructure:"host"`
		Port               int    `mapstructure:"port"`
		User               string `mapstructure:"user"`
		Password           string `mapstructure:"password"`
		From               string `mapstructure:"from"`
		TLSMode            string `mapstructure:"tls_mode"`
		InsecureSkipVerify bool   `mapstructure:"insecure_skip_verify"`
		Bypass             bool   `mapstructure:"bypass"`
	} `mapstructure:"mailer"`

	JWT struct {
		AccessSecret  string        `mapstructure:"access_secret"`
		RefreshSecret string        `mapstructure:"refresh_secret"`
		AccessTTL     time.Duration `mapstructure:"access_ttl"`  // В минутах
		RefreshTTL    time.Duration `mapstructure:"refresh_ttl"` // В минутах
	} `mapstructure:"jwt"`

	Kafka struct {
		Brokers                 []string      `mapstructure:"brokers"`
		UserCreationTopic       string        `mapstructure:"user_creation_topic"`
		UserCreationGroup       string        `mapstructure:"user_creation_group"`
		UserCreationResultTopic string        `mapstructure:"user_creation_result_topic"`
		UserCreationResultGroup string        `mapstructure:"user_creation_result_group"`
		OfferReportPenaltyTopic string        `mapstructure:"offer_report_penalty_topic"`
		OfferReportPenaltyGroup string        `mapstructure:"offer_report_penalty_group"`
		ReputationTopic         string        `mapstructure:"reputation_topic"`
		ReputationGroup         string        `mapstructure:"reputation_group"`
		BatchSize               int           `mapstructure:"batch_size"`
		PollInterval            time.Duration `mapstructure:"poll_interval"`
		WriteTimeout            time.Duration `mapstructure:"write_timeout"`
	} `mapstructure:"kafka"`

	Reputation struct {
		DealCompletionRewardPoints int `mapstructure:"deal_completion_reward_points"`
		ReviewCreationRewardPoints int `mapstructure:"review_creation_reward_points"`
	} `mapstructure:"reputation"`

	Admin struct {
		Email    string `mapstructure:"email"`
		Password string `mapstructure:"password"`
	} `mapstructure:"admin"`

	Storage struct {
		Endpoint         string `mapstructure:"endpoint"`
		PublicBaseURL    string `mapstructure:"public_base_url"`
		AvatarBucket     string `mapstructure:"avatar_bucket"`
		OfferPhotoBucket string `mapstructure:"offer_photo_bucket"`
		AccessKeyID      string `mapstructure:"access_key_id"`
		SecretAccessKey  string `mapstructure:"secret_access_key"`
		Region           string `mapstructure:"region"`
	} `mapstructure:"storage"`

	Port                int    `mapstructure:"port"`
	AuthGRPCAddr        string `mapstructure:"auth_grpc_addr"`
	AuthGRPCListenAddr  string `mapstructure:"auth_grpc_listen_addr"`
	UsersGRPCAddr       string `mapstructure:"users_grpc_addr"`
	UsersGRPCListenAddr string `mapstructure:"users_grpc_listen_addr"`
	DealsGRPCAddr       string `mapstructure:"deals_grpc_addr"`
	DealsGRPCListenAddr string `mapstructure:"deals_grpc_listen_addr"`
	ChatsGRPCAddr       string `mapstructure:"chats_grpc_addr"`
	ChatsGRPCListenAddr string `mapstructure:"chats_grpc_listen_addr"`
}

type ConfigOptions struct {
	CommonPath  string
	ServicePath string
	AppEnv      string // local|docker
}

// LoadConfig загружает конфигурацию из нескольких источников с приоритетом: общий конфиг -> сервисный конфиг ->
// конфиг для окружения -> переменные окружения
func LoadConfig(options ConfigOptions) (Config, error) {
	v := viper.New()

	// общий конфиг, обязательный для всех сервисов
	if options.CommonPath == "" {
		return Config{}, fmt.Errorf("common config path is not set")
	}
	v.SetConfigFile(options.CommonPath)
	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %v", err)
	}

	// конфиг конкретного сервиса, может быть не указан, тогда будет использоваться только общий
	if options.ServicePath != "" {
		v.SetConfigFile(options.ServicePath)
		if err := v.MergeInConfig(); err != nil {
			return Config{}, fmt.Errorf("failed to merge service conflict: %v", err)
		}
	}

	// конфиг для конкретного окружения, может быть не указан, тогда будет использоваться только общий и сервисный
	if options.AppEnv != "" {
		overridePath := "./config/" + options.AppEnv + ".yaml"
		if _, err := os.Stat(overridePath); err == nil {
			v.SetConfigFile(overridePath)
			if err := v.MergeInConfig(); err != nil {
				return Config{}, fmt.Errorf("failed to merge env override conflict: %v", err)
			}
		}
	}

	// переопределение через переменные окружения
	bindEnv(v, "db.password")
	bindEnv(v, "jwt.access_secret")
	bindEnv(v, "jwt.refresh_secret")
	bindEnv(v, "kafka.user_creation_topic")
	bindEnv(v, "kafka.reputation_topic")
	bindEnv(v, "kafka.reputation_group")
	bindEnv(v, "mailer.bypass")
	bindEnv(v, "mailer.host")
	bindEnv(v, "mailer.port")
	bindEnv(v, "mailer.user")
	bindEnv(v, "mailer.password")
	bindEnv(v, "mailer.from")
	bindEnv(v, "mailer.tls_mode")
	bindEnv(v, "mailer.insecure_skip_verify")
	applyEnvAlias(v, "mailer.host", "SMTP_HOST")
	applyEnvAlias(v, "mailer.port", "SMTP_PORT")
	applyEnvAlias(v, "mailer.user", "SMTP_USER")
	applyEnvAlias(v, "mailer.password", "SMTP_PASSWORD")
	applyEnvAlias(v, "mailer.from", "SMTP_FROM")
	applyEnvAlias(v, "mailer.tls_mode", "SMTP_TLS_MODE")
	applyEnvAlias(v, "mailer.insecure_skip_verify", "SMTP_INSECURE_SKIP_VERIFY")
	bindEnv(v, "admin.email")
	bindEnv(v, "admin.password")
	bindEnv(v, "storage.endpoint")
	bindEnv(v, "storage.public_base_url")
	bindEnv(v, "storage.avatar_bucket")
	bindEnv(v, "storage.offer_photo_bucket")
	bindEnv(v, "storage.access_key_id")
	bindEnv(v, "storage.secret_access_key")
	bindEnv(v, "storage.region")
	bindEnv(v, "reputation.deal_completion_reward_points")
	bindEnv(v, "reputation.review_creation_reward_points")

	// десериализация в структуру
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return config, nil
}

func bindEnv(v *viper.Viper, key string) {
	_ = v.BindEnv(key, strings.ToUpper(strings.ReplaceAll(key, ".", "_")))
}

func applyEnvAlias(v *viper.Viper, key, alias string) {
	if value, ok := os.LookupEnv(alias); ok {
		v.Set(key, value)
	}
}
