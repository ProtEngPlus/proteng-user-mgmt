package consumer

import (
	"errors"
	"html/template"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/protengplus/proteng-user-mgmt/internal/logger"
	"github.com/protengplus/proteng-user-mgmt/models"
	"github.com/protengplus/proteng-user-mgmt/repositories"
	"github.com/protengplus/proteng-user-mgmt/utils"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	userRepository repositories.UserRepository
	temp           *template.Template
}

func NewConsumer(userRepository repositories.UserRepository, temp *template.Template) *Consumer {
	return &Consumer{
		userRepository: userRepository,
		temp:           temp,
	}
}

func (c *Consumer) RunConsumer(amqpURL string, queueName string) error {
	attempts := 0
	for {
		attempts++
		conn, err := amqp.Dial(amqpURL)
		if err != nil {
			logger.Errorf("Failed to connect to RabbitMQ: %v", err)
			if attempts >= 5 {
				return errors.New("failed to connect to RabbitMQ after 5 attempts")
			}
			time.Sleep(10 * time.Second)
			continue
		}
		attempts = 0

		ch, err := conn.Channel()
		if err != nil {
			logger.Errorf("Failed to open a channel: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		q, err := ch.QueueDeclare(
			queueName, // name
			true,      // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // no-wait
			nil,       // arguments
		)
		if err != nil {
			logger.Errorf("Failed to declare a queue: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		msgs, err := ch.Consume(
			q.Name, // queue
			"",     // consumer
			true,   // auto-ack
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // arguments
		)
		if err != nil {
			logger.Errorf("Failed to register a consumer: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		logger.Infof(" [*] Waiting for messages from %s", q.Name)

		go func() {
			for d := range msgs {
				logger.Infof("JobConsumer: Received a message: %v from %v", string(d.Body), q.Name)

				// Send Email
				user, err := GetUser(c.userRepository, string(d.Body)[1:len(string(d.Body))-1])
				if err != nil {
					logger.Errorf("Failed to get user information: %v", err)
					continue
				}

				SendJobNotificationEmail(user, c.temp)
			}
		}()

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

		<-sig
		logger.Zap.Info("Shutting down consumer...")
		return nil
	}
}

func GetUser(userRepository repositories.UserRepository, userId string) (*models.User, error) {
	user, err := userRepository.FindById(userId)
	if err != nil {
		logger.Errorf("Failed to get user information: %v", err)
		return nil, err
	}
	return user, nil
}

func SendJobNotificationEmail(user *models.User, temp *template.Template) {
	emailData := utils.EmailData{
		FirstName: user.Name,
		Subject:   "Job status notification",
	}

	err := utils.SendEmail(user, &emailData, temp, "jobStatusNotification")
	if err != nil {
		logger.Errorf("Failed to send email: %v", err)
		return
	}
}
