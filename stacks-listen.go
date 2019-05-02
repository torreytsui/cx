package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/cloud66-oss/cloud66"
	"github.com/cloud66/cli"
	"github.com/cloud66/wray"
	"github.com/mgutz/ansi"
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

func runListen(c *cli.Context) {
	stack := mustStack(c)

	StartListen(stack)
}

func StartListen(stack *cloud66.Stack) {
	if debugMode {
		fmt.Printf("Connecting to Faye on %s\n", profile.FayeEndpoint)
	}

	//	sigChan := make(chan os.Signal, 1)
	//  signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	channel := "/realtime/" + stack.Uid + "/*"

	wray.RegisterTransports([]wray.Transport{&wray.HttpTransport{}})

	fc := wray.NewFayeClient(profile.FayeEndpoint)
	sub := fc.Subscribe(channel, true, handleMessage)
	if debugMode {
		fmt.Printf("Subscribed to %s\n", sub)
	}
	go fc.Listen()

	// handle interrupts
	hupChan := make(chan os.Signal, 1)
	termChan := make(chan os.Signal, 1)
	signal.Notify(hupChan, syscall.SIGHUP)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-termChan:
			return
		case <-hupChan:
			return
		}
	}
}

func handleMessage(msg wray.Message) {
	redColor := ansi.ColorFunc("red+h")
	capColor := ansi.ColorFunc("yellow")
	infoColor := ansi.ColorFunc("white")

	s, err := strconv.Unquote(msg.Data)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	var m logMessage
	err = json.Unmarshal([]byte(s), &m)
	if err != nil {
		fmt.Println("Error:", err)
	}

	var level string
	var colorFunc func(string) string

	switch m.Severity {
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

	fmt.Println(colorFunc(fmt.Sprintf("%s [%s] - %s", m.Time, level, m.Message)))

}
