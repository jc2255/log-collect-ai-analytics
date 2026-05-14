package kafka

import (
	"context"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

// Producer Kafka生产者
type Producer struct {
	writer *kafkago.Writer
}

// NewProducer 创建生产者
func NewProducer(brokers []string, topic string) *Producer {
	w := &kafkago.Writer{
		Addr:         kafkago.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafkago.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafkago.RequireOne,
	}
	return &Producer{writer: w}
}

// NewProducerWithoutTopic 创建不指定topic的生产者(写入时指定)
func NewProducerWithoutTopic(brokers []string) *Producer {
	w := &kafkago.Writer{
		Addr:         kafkago.TCP(brokers...),
		Balancer:     &kafkago.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafkago.RequireOne,
	}
	return &Producer{writer: w}
}

// SendMessage 发送消息
func (p *Producer) SendMessage(ctx context.Context, key, value []byte) error {
	return p.writer.WriteMessages(ctx, kafkago.Message{
		Key:   key,
		Value: value,
	})
}

// SendMessages 批量发送消息
func (p *Producer) SendMessages(ctx context.Context, messages []kafkago.Message) error {
	return p.writer.WriteMessages(ctx, messages...)
}

// SendToTopic 发送到指定topic
func (p *Producer) SendToTopic(ctx context.Context, topic string, key, value []byte) error {
	return p.writer.WriteMessages(ctx, kafkago.Message{
		Topic: topic,
		Key:   key,
		Value: value,
	})
}

// Close 关闭生产者
func (p *Producer) Close() error {
	return p.writer.Close()
}

// Consumer Kafka消费者
type Consumer struct {
	reader *kafkago.Reader
}

// NewConsumer 创建消费者
func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	r := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafkago.LastOffset,
	})
	return &Consumer{reader: r}
}

// ReadMessage 读取消息(阻塞)
func (c *Consumer) ReadMessage(ctx context.Context) (kafkago.Message, error) {
	return c.reader.ReadMessage(ctx)
}

// FetchMessage 获取消息(需要手动commit)
func (c *Consumer) FetchMessage(ctx context.Context) (kafkago.Message, error) {
	return c.reader.FetchMessage(ctx)
}

// CommitMessages 提交offset
func (c *Consumer) CommitMessages(ctx context.Context, msgs ...kafkago.Message) error {
	return c.reader.CommitMessages(ctx, msgs...)
}

// Close 关闭消费者
func (c *Consumer) Close() error {
	return c.reader.Close()
}
