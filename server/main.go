package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/apolunin/cs-demo/common"
	"github.com/apolunin/cs-demo/common/rabbitmq"
	"github.com/apolunin/cs-demo/server/storage"
	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
)

var (
	env = flag.String("e", ".env", "specify file with environment variables")
	num = flag.Int("g", runtime.NumCPU(), "specify the number of goroutines to use for processing")
	out = flag.String("o", "log.txt", "specify file name for logging server activity")
)

func main() {
	flag.Parse()
	common.FailOnError(
		godotenv.Load(*env),
		fmt.Sprintf("cannot load file %q", *env),
	)

	conn, err := rabbitmq.GetConnection()
	common.FailOnError(err, "failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	common.FailOnError(err, "failed to open a channel")
	defer ch.Close()

	q := rabbitmq.DeclareQueue(ch)

	msgs, err := consume(ch, q)
	common.FailOnError(err, "failed to register a consumer")

	logFile, err := os.OpenFile(*out, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	defer logFile.Close()
	log := bufio.NewWriter(logFile)

	go func() {
		items := storage.New()
		sem := make(chan struct{}, *num)
		processed := make(chan string, *num)

		go func() {
			for s := range processed {
				if _, err := log.WriteString(fmt.Sprintf("%s\n", s)); err != nil {
					fmt.Printf("cannot write log record to file: %s", err)
				}
				if err := log.Flush(); err != nil {
					fmt.Printf("cannot flush data to log file: %s", err)
				}
			}
		}()

		for m := range msgs {
			fmt.Printf("received a message: %s\n", m.Body)
			sem <- struct{}{}
			go func(body []byte) {
				processed <- handleMessage(items, body)
				<-sem
			}(m.Body)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println(" [*] Waiting for messages. To exit press CTRL+C")
	<-sigs
}

func consume(ch *amqp.Channel, q amqp.Queue) (<-chan amqp.Delivery, error) {
	return ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
}

func handleMessage(items *storage.ItemStorage, body []byte) string {
	var msg common.Message

	if err := json.Unmarshal(body, &msg); err != nil {
		return fmt.Sprintf("cannot unmarshal message: %s", err)
	}

	var res string

	switch *msg.Action {
	case common.AddItem:
		items.AddItem(msg.Key, msg.Value)
		res = fmt.Sprintf("AddItem - item added: [key: %q, val: %q]", msg.Key, msg.Value)
	case common.RemoveItem:
		items.RemoveItem(msg.Key)
		res = fmt.Sprintf("RemoveItem - item removed: [key: %q]", msg.Key)
	case common.GetItem:
		val, ok := items.GetItem(msg.Key)
		if ok {
			res = fmt.Sprintf("GetItem - item retrieved: [key: %q, val: %q]", msg.Key, val)
		} else {
			res = fmt.Sprintf("GetItem - item is missing: [key: %q]", msg.Key)
		}
	case common.GetAllItems:
		var b strings.Builder
		b.WriteString("GetAllItems:\n")
		for i, item := range items.GetAllItems() {
			b.WriteString(
				fmt.Sprintf(
					"\titem-%d:\t[key: %q, val: %q]\n", i, item.Key, item.Value,
				),
			)
		}
		res = b.String()
	}

	return res
}
