package config

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

var (
	globalConfig *AppConfig
	once         sync.Once
)

// AppConfig 应用全局配置
type AppConfig struct {
	Server    ServerConfig  `mapstructure:"server"`
	MySQL     MySQLConfig   `mapstructure:"mysql"`
	Redis     RedisConfig   `mapstructure:"redis"`
	Kafka     KafkaConfig   `mapstructure:"kafka"`
	ES        ESConfig      `mapstructure:"elasticsearch"`
	JWT       JWTConfig     `mapstructure:"jwt"`
	Log       LogConfig     `mapstructure:"log"`
	License   LicenseConfig `mapstructure:"license"`
	LCATopURL string        `mapstructure:"lca_top_url"`
}

type ServerConfig struct {
	Name string `mapstructure:"name"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug, release, test
}

type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

func (m *MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		m.User, m.Password, m.Host, m.Port, m.DBName)
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
	GroupID string   `mapstructure:"group_id"`
	Topic   string   `mapstructure:"topic"`
}

type ESConfig struct {
	Addresses []string `mapstructure:"addresses"`
	Username  string   `mapstructure:"username"`
	Password  string   `mapstructure:"password"`
	Sniff     bool     `mapstructure:"sniff"`
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpireHour int    `mapstructure:"expire_hour"`
	Issuer     string `mapstructure:"issuer"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"` // MB
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"` // days
	Compress   bool   `mapstructure:"compress"`
}

type LicenseConfig struct {
	PublicKey string `mapstructure:"public_key"`
}

// Init 初始化配置
func Init(configPath string) error {
	var err error
	once.Do(func() {
		viper.SetConfigFile(configPath)
		viper.AutomaticEnv()

		if e := viper.ReadInConfig(); e != nil {
			err = fmt.Errorf("read config file failed: %w", e)
			return
		}

		globalConfig = &AppConfig{}
		if e := viper.Unmarshal(globalConfig); e != nil {
			err = fmt.Errorf("unmarshal config failed: %w", e)
			return
		}
	})
	return err
}

// Get 获取全局配置
func Get() *AppConfig {
	return globalConfig
}
