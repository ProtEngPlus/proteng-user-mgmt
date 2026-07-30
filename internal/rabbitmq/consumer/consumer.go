package consumer

import (
	"encoding/json"
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

type Payload struct {
	UserId    string `json:"user_id"`
	JobName   string `json:"job_name"`
	JobState  string `json:"job_state"`
	StageId   string `json:"stage_id"`
	StageName string `json:"stage_name"`
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
				logger.Infof("Received a message: %v from %v", string(d.Body), q.Name)

				// Decode the incoming message
				var payload Payload
				if err := json.Unmarshal([]byte(string(d.Body)), &payload); err != nil {
					logger.Errorf("Error unmarshal mq payload: %v", err)
					return
				}

				// Send email
				user, err := GetUser(c.userRepository, payload.UserId)
				if err != nil {
					logger.Errorf("Failed to get user information: %v", err)
					continue
				}

				SendJobNotificationEmail(user, payload.JobName, payload.JobState, payload.StageId, payload.StageName, c.temp)
			}
		}()

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

		<-sig
		logger.Zap.Info("Shutting down consumer...")
		os.Exit(0)
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

func SendJobNotificationEmail(user *models.User, jobName string, jobState string, stageID string, stageName string, temp *template.Template) {
	emailData := utils.EmailData{
		FirstName: user.Name,
		Subject:   "Job status notification",
		JobName:   jobName,
		JobState:  jobState,
		StageID:   stageID,
		StageName: stageName,
	}

	err := utils.SendEmail(user, &emailData, temp, "jobStatusNotification")
	if err != nil {
		logger.Errorf("Failed to send email: %v", err)
		return
	}
}
