package config

import (
	"time"
)

type Config struct {
	APIServerHost  string
	KubeConfigFile string
	ResyncPeriod   time.Duration
	SyncRateLimit  float64
	AlertSpeaker   string
	//InfluxDB Config
	InfluxDBAddr      string
	InfluxDBUsernName string
	InfluxDBPassword  string
	InfluxDBName      string
	//RabbitMQ config
	RabbitMQHosts      []string
	RabbitMQTopicName  string
	RabbitExchangeType string
	RabbitDurable      bool
	RabbitRouteKey     string
	RabbitUser         string
	RabbitPassword     string
	RabbitVhost        string

	HttpPort        int
	EnableProfiling bool

	LogLevel string
	LogFile  string
}
