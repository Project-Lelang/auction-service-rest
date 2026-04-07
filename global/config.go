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

type PostgresConfig struct {
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

type YamlConfig struct {
	timeLocation       *time.Location
	BaseDir            string
	StorageDir         string
	AppName            string           `yaml:"app_name"`
	Environment        string           `yaml:"environment"`
	IsDebug            bool             `yaml:"debug"`
	LogChannel         []string         `yaml:"log_channel"`
	Timezone           string           `yaml:"timezone"`
	Port               uint             `yaml:"port"`
	Uri                string           `yaml:"uri"`
	FeUri              string           `yaml:"fe_uri"`
	CorsAllowedOrigins []string         `yaml:"cors_allowed_origins"`
	Postgres           PostgresConfig   `yaml:"postgres"`
	JwtConfig          JwtConfig        `yaml:"jwt"`
	SuperAdmin         SuperAdminConfig `yaml:"super_admin"`
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

func GetAppName() string {
	return config.AppName
}

func GetLogChannel() []string {
	return config.LogChannel
}

func GetTimezone() string {
	return config.Timezone
}

func GetPostgresConfig() PostgresConfig {
	return config.Postgres
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
