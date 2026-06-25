package global

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	ApiVersion = "1.0.0"

	EnvironmentDevelopment = "development"
	EnvironmentProduction  = "production"
	EnvironmentTesting     = "testing"
)

type MysqlConfig struct {
	Host     string `yaml:"host"`
	Port     uint16 `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type JwtConfig struct {
	SecretKey               string `yaml:"secret_key"`
	AccessTokenExpiryHours  int    `yaml:"access_token_expiry_hours"`
	RefreshTokenExpiryHours int    `yaml:"refresh_token_expiry_hours"`
}

type SuperAdminConfig struct {
	Email    string `yaml:"email"`
	Password string `yaml:"password"`
}

type FilesystemConfig struct {
	Type                  string `yaml:"type"`
	BaseUri               string `yaml:"base_uri"`
	PresignedUrlSecretKey string `yaml:"presigned_url_secret_key"`
}

type SupabaseConfig struct {
	URL        string `yaml:"url"`
	ServiceKey string `yaml:"service_key"`
	BucketName string `yaml:"bucket_name"`
}

type MidtransConfig struct {
	ServerKey string `yaml:"server_key"`
	IsSandbox bool   `yaml:"is_sandbox"`
}

type BiteshipConfig struct {
	ApiKey string `yaml:"api_key"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type NotificationConfig struct {
	Enabled     bool   `yaml:"enabled"`
	QueueName   string `yaml:"queue_name"`
	DLQName     string `yaml:"dlq_name"`
	WorkerCount int    `yaml:"worker_count"`
	MaxRetries  int    `yaml:"max_retries"`
	RetryBaseMs int    `yaml:"retry_base_ms"`
}

type ShipmentDeadlineConfig struct {
	BuyerAddressHours         int `yaml:"buyer_address_hours"`
	SellerShipHours           int `yaml:"seller_ship_hours"`
	BuyerReceiveHours         int `yaml:"buyer_receive_hours"`
	TrackingCheckIntervalMins int `yaml:"tracking_check_interval_minutes"`
	DeadlineGraceMinutes      int `yaml:"deadline_grace_minutes"`
}

type FirebaseConfig struct {
	ServiceAccountPath string `yaml:"service_account_path"`
}

type EmailConfig struct {
	Enabled     bool   `yaml:"enabled"`
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	FromAddress string `yaml:"from_address"`
	FromName    string `yaml:"from_name"`
}

type YamlConfig struct {
	timeLocation       *time.Location
	BaseDir            string
	StorageDir         string
	AppName            string                 `yaml:"app_name"`
	Environment        string                 `yaml:"environment"`
	IsDebug            bool                   `yaml:"debug"`
	LogChannel         []string               `yaml:"log_channel"`
	Timezone           string                 `yaml:"timezone"`
	Port               uint                   `yaml:"port"`
	Uri                string                 `yaml:"uri"`
	FeUri              string                 `yaml:"fe_uri"`
	CorsAllowedOrigins []string               `yaml:"cors_allowed_origins"`
	Mysql              MysqlConfig            `yaml:"mysql"`
	JwtConfig          JwtConfig              `yaml:"jwt"`
	SuperAdmin         SuperAdminConfig       `yaml:"super_admin"`
	Filesystem         FilesystemConfig       `yaml:"filesystem"`
	Supabase           *SupabaseConfig        `yaml:"supabase"`
	Midtrans           *MidtransConfig        `yaml:"midtrans"`
	Biteship           *BiteshipConfig        `yaml:"biteship"`
	Redis              RedisConfig            `yaml:"redis"`
	Notification       NotificationConfig     `yaml:"notification"`
	ShipmentDeadline   ShipmentDeadlineConfig `yaml:"shipment_deadline"`
	Firebase           *FirebaseConfig        `yaml:"firebase"`
	Email              EmailConfig            `yaml:"email"`
}

var config YamlConfig

func init() {
	if err := LoadConfig(); err != nil {
		panic(err)
	}
}

func LoadConfig() error {
	baseDir, err := os.Getwd()
	if err != nil {
		return err
	}

	if _, err := os.Stat(fmt.Sprintf("%s/conf.yml", baseDir)); err != nil {
		_, filename, _, _ := runtime.Caller(0)
		baseDir = path.Join(path.Dir(filename), "../")
	}

	config.BaseDir = strings.TrimRight(strings.ReplaceAll(baseDir, "\\\\", "/"), "/")
	config.StorageDir = fmt.Sprintf("%s/storage", config.BaseDir)

	yamlFilePath := fmt.Sprintf("%s/conf.yml", config.BaseDir)
	if _, err := os.Stat(yamlFilePath); err != nil {
		return err
	}

	yamlFile, err := os.ReadFile(yamlFilePath)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(yamlFile, &config); err != nil {
		return err
	}

	config.timeLocation, err = time.LoadLocation(config.Timezone)
	if err != nil {
		return err
	}

	// set defaults
	if config.JwtConfig.AccessTokenExpiryHours == 0 {
		config.JwtConfig.AccessTokenExpiryHours = 24
	}
	if config.JwtConfig.RefreshTokenExpiryHours == 0 {
		config.JwtConfig.RefreshTokenExpiryHours = 168
	}
	if config.Notification.QueueName == "" {
		config.Notification.QueueName = "notif_queue:auction_events"
	}
	if config.Notification.DLQName == "" {
		config.Notification.DLQName = "notif_queue:auction_events:dlq"
	}
	if config.Notification.WorkerCount == 0 {
		config.Notification.WorkerCount = 5
	}
	if config.Notification.MaxRetries == 0 {
		config.Notification.MaxRetries = 3
	}
	if config.Notification.RetryBaseMs == 0 {
		config.Notification.RetryBaseMs = 500
	}
	if config.ShipmentDeadline.BuyerAddressHours == 0 {
		config.ShipmentDeadline.BuyerAddressHours = 24
	}
	if config.ShipmentDeadline.SellerShipHours == 0 {
		config.ShipmentDeadline.SellerShipHours = 72
	}
	if config.ShipmentDeadline.BuyerReceiveHours == 0 {
		config.ShipmentDeadline.BuyerReceiveHours = 168
	}
	if config.ShipmentDeadline.TrackingCheckIntervalMins == 0 {
		config.ShipmentDeadline.TrackingCheckIntervalMins = 60
	}
	if config.ShipmentDeadline.DeadlineGraceMinutes == 0 {
		config.ShipmentDeadline.DeadlineGraceMinutes = 5
	}
	if config.Email.Port == 0 {
		config.Email.Port = 587
	}
	if config.Email.FromName == "" {
		config.Email.FromName = config.AppName
	}

	return nil
}

func GetConfig() YamlConfig {
	return config
}

func GetTimeLocation() *time.Location {
	return config.timeLocation
}

func GetBaseDir() string {
	return config.BaseDir
}

func GetStorageDir() string {
	return config.StorageDir
}

func GetJwtSecretKey() string {
	return config.JwtConfig.SecretKey
}

func GetFilesystem() FilesystemConfig {
	return config.Filesystem
}

func GetSupabaseConfig() *SupabaseConfig {
	return config.Supabase
}

func GetAppName() string {
	return config.AppName
}

func GetLogChannel() []string {
	return config.LogChannel
}

func GetTimezone() string {
	return config.Timezone
}

func GetMysqlConfig() MysqlConfig {
	return config.Mysql
}

func GetSuperAdminConfig() SuperAdminConfig {
	return config.SuperAdmin
}

func IsProduction() bool {
	return config.Environment == EnvironmentProduction
}

func IsDevelopment() bool {
	return config.Environment == EnvironmentDevelopment
}

func IsTesting() bool {
	return config.Environment == EnvironmentTesting
}

func IsDebug() bool {
	return config.IsDebug
}
