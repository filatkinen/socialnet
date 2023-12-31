package producer

import (
	"context"
	"errors"
	"log"

	"github.com/filatkinen/socialnet/internal/rabbit"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	log     *log.Logger
	config  rabbit.Config
	conn    *amqp.Connection
	channel *amqp.Channel
	chExit  chan struct{}
}

func NewProducer(config rabbit.Config, log *log.Logger) (*Producer, error) {
	p := Producer{log: log, config: config}

	configAmpq := amqp.Config{
		Vhost:      "/",
		Properties: amqp.NewConnectionProperties(),
	}
	configAmpq.Properties.SetClientConnectionName("posts_deliver")

	conn, err := amqp.Dial(config.GetDSN())
	if err != nil {
		e := p.Close()
		return nil, errors.Join(err, e)
	}
	p.conn = conn

	channel, err := conn.Channel()
	if err != nil {
		e := p.Close()
		return nil, errors.Join(err, e)
	}
	p.channel = channel

	err = channel.ExchangeDeclare(
		config.ExchangeName, // name
		"direct",            // type
		true,                // durable
		false,               // auto-deleted
		false,               // internal
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		e := p.Close()
		return nil, errors.Join(err, e)
	}
	p.channel = channel

	p.chExit = make(chan struct{})

	return &p, nil
}

func (p *Producer) Stop() {
	p.log.Println("Stopping sender posts")
	p.chExit <- struct{}{}
}

func (p *Producer) Close() (err error) {
	p.log.Println("Closing sender posts")
	if p.channel != nil {
		if e := p.channel.Close(); e != nil {
			err = errors.Join(err, e)
		}
	}
	if p.conn != nil {
		if e := p.conn.Close(); e != nil {
			err = errors.Join(err, e)
		}
	}
	return err
}

func (p *Producer) SendMessages(message []byte, clientID string) {
	err := p.channel.PublishWithContext(context.Background(),
		p.config.ExchangeName, // exchange
		clientID,              // routing key
		false,                 // mandatory
		false,                 // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        message,
		})
	if err != nil {
		p.log.Println("Error publishing: " + err.Error())
		return
	}
}
