package di

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/DelayedNotifier/internal/application"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/DelayedNotifier/internal/domain/interfaces"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/DelayedNotifier/internal/infrastructure/messaging/email"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/DelayedNotifier/internal/infrastructure/repository/redis/notifier"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/DelayedNotifier/internal/logger"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/DelayedNotifier/internal/presentation"

	"go.uber.org/zap"
)

type Container struct {
	Config      *Config
	Logger      *zap.Logger
	HTTPServer  *http.Server
	RabbitConn  *rabbitmq.Connection
	RedisClient interfaces.NotifierRepository
}

func NewContainer(ctx context.Context) *Container {
	config := ReadConfig()
	log := logger.New()

	redisRepo := notifier.NewRedisClient(config.Redis.Addr, config.Redis.Password, config.Redis.DB, log)

	conn, err := rabbitmq.Connect(config.RabbitMQ.URL, config.RabbitMQ.Retries, config.RabbitMQ.Pause)
	if err != nil {
		log.Fatal("failed to connect to RabbitMQ", zap.Error(err))
	}

	channel, err := conn.Channel()
	if err != nil {
		log.Fatal("failed to create channel", zap.Error(err))
	}

	exchange := rabbitmq.NewExchange(config.RabbitMQ.Exchange, "fanout")
	exchange.Durable = true
	if err := exchange.BindToChannel(channel); err != nil {
		log.Fatal("failed to declare exchange", zap.Error(err))
	}

	qManager := rabbitmq.NewQueueManager(channel)
	queue, err := qManager.DeclareQueue(config.RabbitMQ.Queue, rabbitmq.QueueConfig{Durable: true})
	if err != nil {
		log.Fatal("failed to declare queue", zap.Error(err))
	}

	err = channel.QueueBind(queue.Name, "", exchange.Name(), false, nil)
	if err != nil {
		log.Fatal("failed to bind queue", zap.Error(err))
	}

	producer := rabbitmq.NewPublisher(channel, exchange.Name())
	emailClient, _ := email.NewEmailSender(config.SMTPClient.Host, config.SMTPClient.Email, config.SMTPClient.Password, config.SMTPClient.Port)

	notifierService := application.NewNotifierService(redisRepo, emailClient, log)
	workerChannel, err := conn.Channel()
	if err != nil {
		log.Fatal("failed to create channel", zap.Error(err))
	}

	sched := application.NewScheduler(redisRepo, producer, log)
	go sched.Run(ctx)

	worker := application.NewNotifierWorker(notifierService, workerChannel, config.RabbitMQ.Queue, log)
	go worker.Run()

	httpServer := InitHTTPServer(config.HTTP, notifierService, log)

	return &Container{
		Config:      &config,
		Logger:      log,
		HTTPServer:  httpServer,
		RabbitConn:  conn,
		RedisClient: redisRepo,
	}
}

func InitHTTPServer(cfg HTTPConfig, notifierService interfaces.NotifierService, logger *zap.Logger) *http.Server {
	router := ginext.New()
	presentation.InitRoutes(router, notifierService, logger)

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	return &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
}
