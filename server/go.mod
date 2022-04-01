module github.com/apolunin/cs-demo/server

go 1.18

replace github.com/apolunin/cs-demo/common => ../common

require (
	github.com/apolunin/cs-demo/common v0.0.0-00010101000000-000000000000
	github.com/joho/godotenv v1.4.0
	github.com/streadway/amqp v1.0.0
)

require github.com/apolunin/orderedmap v0.1.0 // indirect
