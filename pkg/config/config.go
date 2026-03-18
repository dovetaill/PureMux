package config

// Config 定义应用阶段一需要的强类型配置入口。
type Config struct {
	App   AppConfig   `yaml:"app"`
	MySQL MySQLConfig `yaml:"mysql"`
	Redis RedisConfig `yaml:"redis"`
	Log   LogConfig   `yaml:"log"`
}

// AppConfig 定义应用基础监听配置。
type AppConfig struct {
	Name string `yaml:"name" env:"APP_NAME" env-required:"true"`
	Env  string `yaml:"env" env:"APP_ENV" env-default:"local"`
	Host string `yaml:"host" env:"APP_HOST" env-default:"0.0.0.0"`
	Port int    `yaml:"port" env:"APP_PORT" env-default:"8080"`
}

// MySQLConfig 定义 MySQL 连接与连接池参数。
type MySQLConfig struct {
	Host                   string `yaml:"host" env:"MYSQL_HOST" env-required:"true"`
	Port                   int    `yaml:"port" env:"MYSQL_PORT" env-default:"3306"`
	User                   string `yaml:"user" env:"MYSQL_USER" env-required:"true"`
	Password               string `yaml:"password" env:"MYSQL_PASSWORD" env-required:"true"`
	DBName                 string `yaml:"dbname" env:"MYSQL_DBNAME" env-required:"true"`
	Charset                string `yaml:"charset" env:"MYSQL_CHARSET" env-default:"utf8mb4"`
	ParseTime              bool   `yaml:"parse_time" env:"MYSQL_PARSE_TIME" env-default:"true"`
	Loc                    string `yaml:"loc" env:"MYSQL_LOC" env-default:"Local"`
	MaxOpenConns           int    `yaml:"max_open_conns" env:"MYSQL_MAX_OPEN_CONNS" env-default:"20"`
	MaxIdleConns           int    `yaml:"max_idle_conns" env:"MYSQL_MAX_IDLE_CONNS" env-default:"10"`
	ConnMaxLifetimeMinutes int    `yaml:"conn_max_lifetime_minutes" env:"MYSQL_CONN_MAX_LIFETIME_MINUTES" env-default:"60"`
}

// RedisConfig 定义 Redis 连接与连接池参数。
type RedisConfig struct {
	Addr         string `yaml:"addr" env:"REDIS_ADDR" env-required:"true"`
	Password     string `yaml:"password" env:"REDIS_PASSWORD"`
	DB           int    `yaml:"db" env:"REDIS_DB" env-default:"0"`
	PoolSize     int    `yaml:"pool_size" env:"REDIS_POOL_SIZE" env-default:"10"`
	MinIdleConns int    `yaml:"min_idle_conns" env:"REDIS_MIN_IDLE_CONNS" env-default:"2"`
}

// LogConfig 定义结构化日志输出参数。
type LogConfig struct {
	Level       string `yaml:"level" env:"LOG_LEVEL" env-default:"info"`
	Format      string `yaml:"format" env:"LOG_FORMAT" env-default:"json"`
	Output      string `yaml:"output" env:"LOG_OUTPUT" env-default:"both"`
	Dir         string `yaml:"dir" env:"LOG_DIR" env-default:"logs"`
	Filename    string `yaml:"filename" env:"LOG_FILENAME" env-default:"app.log"`
	MaxSizeMB   int    `yaml:"max_size_mb" env:"LOG_MAX_SIZE_MB" env-default:"100"`
	MaxBackups  int    `yaml:"max_backups" env:"LOG_MAX_BACKUPS" env-default:"14"`
	MaxAgeDays  int    `yaml:"max_age_days" env:"LOG_MAX_AGE_DAYS" env-default:"30"`
	Compress    bool   `yaml:"compress" env:"LOG_COMPRESS" env-default:"false"`
	RotateDaily bool   `yaml:"rotate_daily" env:"LOG_ROTATE_DAILY" env-default:"true"`
}
