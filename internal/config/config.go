package config

import "time"

type Config struct {
	Stage       string `env:"STAGE" envDefault:"DEV"`
	WorkerCount int    `env:"WORKER_COUNT" envDefault:"5"`
	AddressFile string `env:"ADDRESS_FILE" envDefault:"addresses.json"`
	InstanceID  string `env:"INSTANCE_ID" envDefault:"local-instance-1"`

	HTTP     HTTPConfig
	Database DatabaseConfig
	Ethereum EthereumConfig
	Kafka    KafkaConfig
	Redis    RedisConfig
}

type DatabaseConfig struct {
	Host     string `env:"DB_HOST" envDefault:"localhost"`
	Port     int    `env:"DB_PORT" envDefault:"5432"`
	User     string `env:"DB_USER" envDefault:"deblock"`
	Password string `env:"DB_PASSWORD" envDefault:"password"`
	DBName   string `env:"DB_NAME" envDefault:"deblock_monitoring"`
	SSLMode  string `env:"DB_SSLMODE" envDefault:"disable"`
	MaxConns int    `env:"DB_MAX_CONNS" envDefault:"25"`
}

type EthereumConfig struct {
	RPCURL    string `env:"ETH_RPC_URL" envDefault:"wss://boldest-ultra-brook.quiknode.pro/66e2faa15f6ac8c352251544ef668501ceba0c81"`
	ChainID   int64  `env:"ETH_CHAIN_ID" envDefault:"1"`
	BatchSize int    `env:"ETH_BATCH_SIZE" envDefault:"100"`
}

type KafkaConfig struct {
	Brokers []string `env:"KAFKA_BROKERS" envSeparator:"," envDefault:"localhost:9092"`
	Topic   string   `env:"KAFKA_TOPIC" envDefault:"ethereum-transactions"`
	GroupID string   `env:"KAFKA_GROUP_ID" envDefault:"deblock-monitor"`
}

type RedisConfig struct {
	Addr     string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	Password string `env:"REDIS_PASSWORD" envDefault:""`
	DB       int    `env:"REDIS_DB" envDefault:"0"`
}

type HTTPConfig struct {
	Address      string        `env:"HTTP_ADDRESS" envDefault:":8080"`
	ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"30s"`
	WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"30s"`
	IdleTimeout  time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"120s"`
}
