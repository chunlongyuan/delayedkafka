package config

var Cfg Config

type Config struct {
	Env            string   `env:"ENV"`
	LogLevel       string   `env:"LOG_LEVEL"`
	Port           string   `env:"PORT" envDefault:"8000"`
	KafkaHost      []string `env:"KAFKA_SERVER"`
	PProfPort      string   `env:"PPROF_PORT" envDefault:"6060"`
	HealthPort     string   `env:"HEALTH_PORT" envDefault:"6062"`
	PrometheusPort string   `env:"PROMETHEUS_PORT" envDefault:"8080"`
	RedisCacheHost string   `env:"REDIS_CACHE_HOST"`
	RedisCachePort string   `env:"REDIS_CACHE_PORT" envDefault:"6379"`
	DBHostname     string   `env:"DB_HOSTNAME"`
	DBPort         string   `env:"DB_PORT" envDefault:"3306"`
	DBUsername     string   `env:"DB_USERNAME" envDefault:"test"`
	DBPassword     string   `env:"DB_PASSWORD" envDefault:"test123"`
	DBDatabase     string   `env:"DB_DATABASE"`
	QueueKeyword   string   `env:"KDQueueKeyword" envDefault:"kdqueue"` // 涉及到缓存的关键字
}
