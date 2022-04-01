package main

import (
  "bufio"
  "encoding/json"
  "flag"
  "fmt"
  "os"
  "strings"
  "time"

  "github.com/apolunin/cs-demo/common"
  "github.com/apolunin/cs-demo/common/rabbitmq"
  "github.com/joho/godotenv"
  "github.com/streadway/amqp"
)

var (
  env = flag.String("e", ".env", "specify file with environment variables")
  fp  = flag.String("f", "", "specify input file, if omitted then data is read from console")
  dly = flag.Int(
    "d", 0,
    "specify delay in milliseconds between adjacent message sends, if omitted then delay is 0",
  )
)

func main() {
  flag.Parse()
  common.FailOnError(
    godotenv.Load(*env),
    fmt.Sprintf("cannot load file %q", *env),
  )

  var err error
  var input *os.File
  if *fp != "" {
    fmt.Printf("reading message data from file: %q\n", *fp)
    input, err = os.Open(*fp)
    common.FailOnError(
      err,
      fmt.Sprintf("cannot open input file: %s\n", *fp),
    )

    defer input.Close()
  } else {
    input = os.Stdin
  }

  delay := time.Duration(*dly)

  conn, err := rabbitmq.GetConnection()
  common.FailOnError(err, "failed to connect to RabbitMQ")
  defer conn.Close()

  ch, err := conn.Channel()
  common.FailOnError(err, "failed to open a channel")
  defer ch.Close()

  q := rabbitmq.DeclareQueue(ch)

  for s := bufio.NewScanner(input); s.Scan(); {
    fmt.Printf("processing message: %q\n", s.Text())

    time.Sleep(delay * time.Millisecond)

    msg, err := parse(s.Text())
    if err != nil {
      fmt.Printf("error: %s\n", err)
      continue
    }

    if err = send(msg, ch, q); err != nil {
      fmt.Printf("error: %s\n", err)
      continue
    }

    fmt.Printf("message sent: %q\n", s.Text())
  }
}

func parse(message string) (*common.Message, error) {
  s := strings.Split(message, " ")
  if len(s) == 0 {
    return nil, fmt.Errorf("cannot parse message %q", message)
  }

  at := common.ActionType(s[0])
  if !at.IsValid() {
    return nil, fmt.Errorf("cannot parse message %q: action type %q is invalid", message, at)
  }

  var msg *common.Message

  switch {
  case (at == common.AddItem) && (len(s) == 3):
    msg = &common.Message{Action: &at, Key: s[1], Value: s[2]}
  case (at == common.RemoveItem || at == common.GetItem) && (len(s) == 2):
    msg = &common.Message{Action: &at, Key: s[1]}
  case (at == common.GetAllItems) && (len(s) == 1):
    msg = &common.Message{Action: &at}
  default:
    return nil, fmt.Errorf("cannot parse message %q", message)
  }

  return msg, nil
}

func send(message *common.Message, ch *amqp.Channel, q amqp.Queue) error {
  body, err := json.Marshal(message)
  if err != nil {
    return fmt.Errorf("cannot marshal message: %s", err)
  }

  err = ch.Publish(
    "",     // exchange
    q.Name, // routing key
    false,  // mandatory
    false,  // immediate
    amqp.Publishing{
      ContentType: "application/json",
      Body:        body,
    },
  )

  if err != nil {
    return fmt.Errorf("cannot send message: %s", err)
  }

  return nil
}
