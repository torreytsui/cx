package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bitly/go-nsq"
	"github.com/bitly/nsq/util"
	"github.com/mgutz/ansi"
)

var cmdListen = &Command{
	Run:        runListen,
	Usage:      "listen",
	NeedsStack: true,
	Category:   "stack",
	Short:      "tails all deployment logs",
	Long: `This acts as a log tail for deployment of a stack so you don't have to follow the deployment on the web.

Examples:
$ cx listen
$ cx listen -s mystack
`,
}

type logMessage struct {
	Severity   int       `json:"severity"`
	Message    string    `json:"message"`
	Time       time.Time `json:"time"`
	Raw        bool      `json:"is_raw"`
	Deployment bool      `json:"is_cap"`
}

type tailHandler struct {
	totalMessages int
	messagesShown int
}

var (
	lookupdHTTPAddrs = util.StringArray{}
)

const (
	totalMessages = 0
)

func runListen(cmd *Command, args []string) {
	stack := mustStack()

	if len(args) > 0 {
		cmd.printUsage()
		os.Exit(2)
	}

	maxInFlight := 200

	// build a ephemeral channel
	rand.Seed(time.Now().UnixNano())
	channel := fmt.Sprintf("listen%06d#ephemeral", rand.Int()%999999)
	topic := stack.Uid

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Don't ask for more messages than we want
	if totalMessages > 0 && totalMessages < maxInFlight {
		maxInFlight = totalMessages
	}

	cfg := nsq.NewConfig()
	cfg.UserAgent = fmt.Sprintf("cx/%s go-nsq/%s", VERSION, nsq.VERSION)
	cfg.MaxInFlight = maxInFlight

	consumer, err := nsq.NewConsumer(topic, channel, cfg)
	if err != nil {
		printFatal(err.Error())
	}

	if !debugMode {
		nullLogger := log.New(ioutil.Discard, "", log.LstdFlags)
		consumer.SetLogger(nullLogger, nsq.LogLevelDebug)
	}

	consumer.AddHandler(&tailHandler{totalMessages: totalMessages})

	lookupdHTTPAddrs.Set(nsqLookup)

	err = consumer.ConnectToNSQLookupds(lookupdHTTPAddrs)
	if err != nil {
		printFatal(err.Error())
	}

	for {
		select {
		case <-consumer.StopChan:
			return
		case <-sigChan:
			consumer.Stop()
		}
	}
}

func (th *tailHandler) HandleMessage(m *nsq.Message) error {
	redColor := ansi.ColorFunc("red+h")
	capColor := ansi.ColorFunc("yellow")
	infoColor := ansi.ColorFunc("white")

	th.messagesShown++

	var message logMessage
	err := json.Unmarshal(m.Body, &message)
	if err != nil {
		fmt.Println("error:", err)
	}

	var level string
	var colorFunc func(string) string

	switch message.Severity {
	case 0:
		level = "TRACE"
		colorFunc = infoColor
	case 1:
		level = "DEBUG"
		colorFunc = infoColor
	case 2:
		level = "INFO"
		colorFunc = infoColor
	case 3:
		level = "WARN"
		colorFunc = capColor
	case 4:
		level = "ERROR"
		colorFunc = redColor
	case 5:
		level = "IMPORTANT"
		colorFunc = redColor
	case 6:
		level = "FATAL"
		colorFunc = redColor
	}

	fmt.Println(colorFunc(fmt.Sprintf("%s [%s] - %s", message.Time, level, message.Message)))

	if th.totalMessages > 0 && th.messagesShown >= th.totalMessages {
		os.Exit(0)
	}
	return nil
}
