package config

import (
	"crypto/tls"
	"github.com/Shopify/sarama"
	"gocloud.dev/pubsub/kafkapubsub"
	"os"
)

func GetKafkaConfig() *sarama.Config {
	conf := kafkapubsub.MinimalConfig()

	conf.Net.SASL.User = os.Getenv("KAFKA_USERNAME")
	conf.Net.SASL.Password = os.Getenv("KAFKA_PASSWORD")

	conf.Producer.Retry.Max = 5
	conf.Producer.RequiredAcks = sarama.WaitForAll
	conf.Metadata.Full = true
	conf.Metadata.Full = true
	conf.Net.SASL.Enable = true
	conf.Net.SASL.Handshake = true
	conf.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	conf.Net.TLS.Enable = true
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ClientAuth:         0,
	}
	conf.Net.TLS.Config = tlsConfig
	conf.ClientID = "group"

	return conf
}
