package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bitly/go-nsq"
	"github.com/bitly/nsq/util"
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

const (
	CLR_0 = "\x1b[30;1m"
	CLR_R = "\x1b[31;1m"
	CLR_G = "\x1b[32;1m"
	CLR_Y = "\x1b[33;1m"
	CLR_B = "\x1b[34;1m"
	CLR_M = "\x1b[35;1m"
	CLR_C = "\x1b[36;1m"
	CLR_W = "\x1b[37;1m"
	CLR_N = "\x1b[0m"
)

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

	consumer.AddHandler(&tailHandler{totalMessages: totalMessages})

	lookupdHTTPAddrs.Set("localhost:4161")

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
	th.messagesShown++

	var message logMessage
	err := json.Unmarshal(m.Body, &message)
	if err != nil {
		fmt.Println("error:", err)
	}

	if message.Severity == 4 {
		fmt.Printf("%s%s\n", CLR_R, message.Message)
		fmt.Print(CLR_W)
	} else {
		fmt.Printf("%s%s\n", CLR_W, message.Message)
	}

	if th.totalMessages > 0 && th.messagesShown >= th.totalMessages {
		os.Exit(0)
	}
	return nil
}
