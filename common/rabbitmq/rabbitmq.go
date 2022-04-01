package rabbitmq

import (
  "fmt"
  "log"
  "os"

  "github.com/apolunin/cs-demo/common"
  "github.com/streadway/amqp"
)

const (
  RabbitUser     = "RABBIT_USERNAME"
  RabbitPassword = "RABBIT_PASSWORD"
  RabbitHost     = "RABBIT_HOST"
  RabbitPort     = "RABBIT_PORT"
  RabbitQueue    = "RABBIT_QUEUE"
)

func getConnectionURL() string {
  usr, pwd, host, port :=
      mustGetEnvVar(RabbitUser),
      mustGetEnvVar(RabbitPassword),
      mustGetEnvVar(RabbitHost),
      mustGetEnvVar(RabbitPort)

  return fmt.Sprintf("amqp://%s:%s@%s:%s/", usr, pwd, host, port)
}

func GetConnection() (*amqp.Connection, error) {
  return amqp.Dial(getConnectionURL())
}

func DeclareQueue(ch *amqp.Channel) amqp.Queue {
  q, err := ch.QueueDeclare(
    mustGetEnvVar(RabbitQueue), // name
    false,                      // durable
    false,                      // delete when unused
    false,                      // exclusive
    false,                      // no-wait
    nil,                        // arguments
  )
  common.FailOnError(err, "Failed to declare a queue")
  return q
}

func mustGetEnvVar(name string) string {
  value := os.Getenv(name)
  if value == "" {
    log.Fatalf("%s environment variable is not set", name)
  }
  return value
}
