package main

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"log"
	"math/rand"
	"time"
)

var brokers = []string{"localhost:9095", "localhost:9096"}

func newSaramaConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true

	return config
}

func newSyncProducer() (sarama.SyncProducer, error) {
	config := newSaramaConfig()
	producer, err := sarama.NewSyncProducer(brokers, config)

	return producer, err
}

func newAsyncProducer() (sarama.AsyncProducer, error) {
	config := newSaramaConfig()
	producer, err := sarama.NewAsyncProducer(brokers, config)

	return producer, err
}

func prepareMessage(topic string, message []byte) *sarama.ProducerMessage {
	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Partition: -1,
		Value:     sarama.ByteEncoder(message),
	}

	return msg
}

type ExampleInfo struct {
	Uuid      string
	CreatedAt time.Time
}

func main() {

	syncProducer, err := newSyncProducer()
	if err != nil {
		log.Fatal(err)
	}

	asyncProducer, err := newAsyncProducer()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for err := range asyncProducer.Errors() {
			fmt.Printf("Msg async err: %s\n", err)
		}
	}()

	go func() {
		for successMsg := range asyncProducer.Successes() {
			fmt.Printf("Msg async success:  Partition %d, offset %d\n", successMsg.Partition, successMsg.Offset)
		}
	}()

	for {
		exampleModel := ExampleInfo{
			Uuid:      uuid.NewString(),
			CreatedAt: time.Now(),
		}

		exampleModelJson, err := json.Marshal(&exampleModel)
		if err != nil {
			log.Fatal(err)
		}

		message := prepareMessage("uuid_for_import", exampleModelJson)

		if rand.Int()%2 == 0 {
			partition, offset, err := syncProducer.SendMessage(message)
			if err != nil {
				fmt.Printf("Sync SendMessage error %s", err)
			} else {
				fmt.Printf("Sync message sent. Partition %d, offset %d\n", partition, offset)
			}
		} else {
			asyncProducer.Input() <- message
		}

		time.Sleep(time.Second)

	}

}
