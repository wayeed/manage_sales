package configs

import (
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	Log      LogConfig      `yaml:"log"`
	CORS     CORSConfig     `yaml:"cors"`
	Upload   UploadConfig   `yaml:"upload"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type DatabaseConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	DBName       string `yaml:"dbname"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
	MaxOpenConns int    `yaml:"max_open_conns"`
}

type JWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
}

type LogConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
}

type UploadConfig struct {
	Dir     string `yaml:"dir"`
	BaseURL string `yaml:"base_url"`
}

var GlobalConfig *Config

// LoadConfig 从指定路径加载配置文件
// 自动展开配置值中的环境变量引用（如 ${DB_PASSWORD} 或 $DB_PASSWORD）
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// 展开环境变量，支持 ${VAR:-default} 默认值语法以及 ${VAR} / $VAR
	expanded := expandEnvWithDefault(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, err
	}

	GlobalConfig = &cfg
	return &cfg, nil
}

// expandEnvWithDefault 支持 ${VAR:-default} 和 ${VAR} / $VAR 语法
// 当环境变量未设置时使用 default 值
func expandEnvWithDefault(s string) string {
	result := s

	// 处理 ${VAR:-default} 模式
	patternWithDefault := regexp.MustCompile(`\$\{([^:}]+):-([^}]*)\}`)
	result = patternWithDefault.ReplaceAllStringFunc(result, func(match string) string {
		parts := patternWithDefault.FindStringSubmatch(match)
		if len(parts) == 3 {
			varName := parts[1]
			defaultVal := parts[2]
			if val := os.Getenv(varName); val != "" {
				return val
			}
			return defaultVal
		}
		return match
	})

	// 处理 ${VAR} 模式（不带默认值）
	patternNoDefault := regexp.MustCompile(`\$\{([^}]+)\}`)
	result = patternNoDefault.ReplaceAllStringFunc(result, func(match string) string {
		parts := patternNoDefault.FindStringSubmatch(match)
		if len(parts) == 2 {
			varName := parts[1]
			if val := os.Getenv(varName); val != "" {
				return val
			}
		}
		return match
	})

	// 最后使用 os.ExpandEnv 处理 $VAR 模式
	result = os.ExpandEnv(result)
	return result
}
