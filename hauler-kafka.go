package main

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/logiqai/hauler-common"
	"github.com/logiqai/utils/log"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
)

var (
	logger   = log.NewPackageLogger(logrus.InfoLevel)
	logEntry = log.NewPackageLoggerEntry(logger, "hauler-kafka")
)

func main() {
	db := new(hauler_common.LogiqJsonBatchBackend)
	host, exists := os.LookupEnv("LOGIQ_HOST")
	if exists {
		db.LogiqHost = host
	} else {
		db.LogiqHost = "logiq-flash"
	}
	port, exists := os.LookupEnv("LOGIQ_PORT")
	if exists {
		db.LogiqPort = port
	} else {
		db.LogiqPort = "443"
	}

	bootstrapservers, exists := os.LookupEnv("BOOTSTRAP_SERVERS")

	if !exists {
		for {
			logEntry.ContextLogger().Errorf("BOOTSTRAP_SERVERS not set")
			time.Sleep(time.Minute)
		}
	}

	topics, exists := os.LookupEnv("TOPIC")

	if !exists {
		for {
			logEntry.ContextLogger().Errorf("TOPIC not set")
			time.Sleep(time.Minute)
		}
	}

	groupid, exists := os.LookupEnv("GROUP_ID")

	if !exists {
		for {
			logEntry.ContextLogger().Errorf("GROUP_ID not set")
			time.Sleep(time.Minute)
		}
	}

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  bootstrapservers,
		"group.id":           groupid,
		"enable.auto.commit": false,
		"auto.offset.reset":  "smallest"})

	if err != nil {
		for {
			logEntry.ContextLogger().Errorf("Error creating consumer: %v", err)
			time.Sleep(time.Minute)
		}
	}

	topicsList := strings.Split(topics, ",")

	err = consumer.SubscribeTopics(topicsList, nil)

	cluster, clusterExists := os.LookupEnv("service_name")
	namespace := "kafka"
	if !clusterExists {
		cluster = "kafka-ingest"
	}
	messageBatch := []map[string]interface{}{}
	ticker := time.NewTicker(time.Second * 1)

	for {
		select {
		case <-ticker.C:
			if len(messageBatch) > 0 {
				if err := db.PublishBatchSizeLimitOrFlushTimeout(messageBatch); err != nil {
					logEntry.ContextLogger().Errorf("Error publishing batch: %v", err)
				} else {
					consumer.Commit()
				}
				messageBatch = []map[string]interface{}{}
			}
		default:
			ev := consumer.Poll(100)
			switch e := ev.(type) {
			case *kafka.Message:
				// application-specific processing
				m := map[string]interface{}{
					"timestamp":   e.Timestamp.Format(time.RFC3339),
					"namespace":   namespace,
					"cluster_id":  cluster,
					"application": (*e.TopicPartition.Topic),
					"message":     fmt.Sprintf("%s", e.Value),
				}
				messageBatch = append(messageBatch, m)
			case kafka.PartitionEOF:
				break
			case kafka.Error:
				logEntry.ContextLogger().Errorf("%% Error: %v\n", e.Error())
				errorMessage := map[string]interface{}{
					"timestamp":  time.Now().Format(time.RFC3339),
					"namespace":  namespace,
					"cluster_id": cluster,
					"message":    e.Error(),
					"severity":   "error",
				}
				messageBatch = append(messageBatch, errorMessage)
			default:
				logEntry.ContextLogger().Debugf("Ignored %v\n", e)
			}
			if len(messageBatch) >= 25 {
				if err := db.PublishBatchSizeLimitOrFlushTimeout(messageBatch); err != nil {
					logEntry.ContextLogger().Errorf("Error publishing batch: %v", err)
				} else {
					consumer.Commit()
				}
				messageBatch = []map[string]interface{}{}
			}
		}
	}
	consumer.Close()
}
