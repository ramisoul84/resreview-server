package config

import "time"

type Config struct {
	App      AppConfig
	Logging  LoggingConfig
	Server   ServerConfig
	JWT      JWTConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	MinIO    MinIOConfig
}

type AppConfig struct {
	Env     string
	Name    string
	Version string
}
type LoggingConfig struct {
	Level  string
	Output string
	File   string
}

type ServerConfig struct {
	Port           string
	WSPort         string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	AllowedOrigins string
	BodyLimitMB    int
}

type JWTConfig struct {
	Secret             string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	Issuer             string
}

type PostgresConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	Name         string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	ConnLifetime time.Duration
	ConnIdleTime time.Duration
}

type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type MinIOConfig struct {
	Endpoint          string
	AccessKey         string
	SecretKey         string
	Bucket            string
	UseSSL            bool
	PublicURL         string
	InitBucketOnStart bool
	MaxSizeMB         int
}

func Load(env string) *Config {
	loadDotenv(env)

	cfg := &Config{
		App: AppConfig{
			Env:     getEnv("APP_ENV", "development"),
			Name:    getEnv("APP_NAME", "resview-server"),
			Version: getEnv("APP_VERSION", "1.0.0"),
		},

		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Output: getEnv("LOG_OUTPUT", "stdout"),
			File:   getEnv("LOG_FILE", "logs/resview-server.log"),
		},

		Server: ServerConfig{
			Port:           getEnv("PORT", "8080"),
			WSPort:         getEnv("WS_PORT", "8001"),
			ReadTimeout:    getEnvDuration("READ_TIMEOUT", 15*time.Second),
			WriteTimeout:   getEnvDuration("WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:    getEnvDuration("IDLE_TIMEOUT", 120*time.Second),
			AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:4200"),
			BodyLimitMB:    getEnvInt("BODY_LIMIT_MB", 10),
		},

		JWT: JWTConfig{
			Secret:             mustGetEnv("JWT_SECRET"),
			AccessTokenExpiry:  getEnvDuration("JWT_ACCESS_EXPIRY", 15*time.Minute),
			RefreshTokenExpiry: getEnvDuration("JWT_REFRESH_EXPIRY", 168*time.Hour),
			Issuer:             getEnv("JWT_ISSUER", "resreview-server"),
		},

		Postgres: PostgresConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnv("DB_PORT", "5432"),
			User:         mustGetEnv("DB_USER"),
			Password:     mustGetEnv("DB_PASSWORD"),
			Name:         mustGetEnv("DB_NAME"),
			SSLMode:      getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 10),
			ConnLifetime: getEnvDuration("DB_CONN_LIFETIME", 5*time.Minute),
			ConnIdleTime: getEnvDuration("DB_CONN_IDLE_TIME", 1*time.Minute),
		},

		Redis: RedisConfig{
			Addr:         getEnv("REDIS_ADDR", "localhost:6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getEnvInt("REDIS_DB", 0),
			DialTimeout:  getEnvDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:  getEnvDuration("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout: getEnvDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
		},

		MinIO: MinIOConfig{
			Endpoint:          mustGetEnv("MINIO_ENDPOINT"),
			AccessKey:         mustGetEnv("MINIO_ACCESS_KEY"),
			SecretKey:         mustGetEnv("MINIO_SECRET_KEY"),
			Bucket:            getEnv("MINIO_BUCKET", "dating-photos"),
			UseSSL:            getEnvBool("MINIO_USE_SSL", false),
			PublicURL:         mustGetEnv("MINIO_PUBLIC_URL"),
			InitBucketOnStart: getEnvBool("MINIO_INIT_BUCKET_ON_START", true),
			MaxSizeMB:         getEnvInt("PHOTO_MAX_SIZE_MB", 10),
		},
	}

	return cfg
}

func (c MinIOConfig) MaxSizeBytes() int64 {
	return int64(c.MaxSizeMB) * 1024 * 1024
}

func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}
