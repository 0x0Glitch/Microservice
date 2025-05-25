package main

import (
	"encoding/json"

	"github.com/0x0Glitch/toll-calculator/types"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type DataProducer interface {
	ProduceData(types.OBUData) error
}

type kafkaProducer struct {
	producer *kafka.Producer
	topic string
}

func NewKafkaProducer(topic string) (DataProducer, error) {

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
	if err != nil {
		panic(err)
	}
	// start another go routine to check if we have delivered the data
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					// fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					// fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	return &kafkaProducer{
		producer: p,
		topic: topic,
	}, nil
}
func (p *kafkaProducer) ProduceData(data types.OBUData) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &p.topic,
			Partition: kafka.PartitionAny},
		Value: b,
	}, nil)

}
