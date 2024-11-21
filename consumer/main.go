package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"log"
	"time"
)

var brokers = []string{"localhost:9095", "localhost:9096"}

type ExampleInfo struct {
	Uuid      string
	CreatedAt time.Time
}

type ExampleConsumer struct{}

func (c *ExampleConsumer) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c *ExampleConsumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c *ExampleConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		var exampleModel ExampleInfo
		err := json.Unmarshal(message.Value, &exampleModel)
		if err != nil {
			fmt.Printf("Error unmarshaling message: %s\n", err)
		}

		fmt.Printf("Msg received. uuid %v created_at: %v\n", exampleModel.Uuid, exampleModel.CreatedAt)
		session.MarkMessage(message, "")
	}

	return nil
}

func subscribe(ctx context.Context, topic string, consumerGroup sarama.ConsumerGroup) error {
	var consumer sarama.ConsumerGroupHandler
	consumer = &ExampleConsumer{}

	go func() {
		for {
			if err := consumerGroup.Consume(ctx, []string{topic}, consumer); err != nil {
				fmt.Printf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				fmt.Printf("Error from consumer: %v", ctx.Err())
				return
			}
		}
	}()

	return nil
}

func main() {

	config := sarama.NewConfig()

	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumerGroup, err := sarama.NewConsumerGroup(brokers, "cons", config)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	errSubscribe := subscribe(ctx, "uuid_for_import", consumerGroup)
	if errSubscribe != nil {
		log.Fatal(err)
	}

	<-ctx.Done()

}
