package main

import (
	"flag"
	"fmt"
	"github.com/RacoonMediaServer/rms-packages/pkg/communication"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net/http"
	"net/url"
	"time"
)

func main() {
	token := flag.String("token", "", "Device token")
	needCode := flag.Bool("code", false, "Need to acquire linkage code")
	flag.Parse()

	u := url.URL{
		Scheme: "ws",
		Host:   "127.0.0.1:8080",
		Path:   "/bot",
	}

	h := make(http.Header)
	h.Add("X-Token", *token)

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), h)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if *needCode {
		acquireCode(conn)
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		cnt := 0

		for {
			<-ticker.C
			msg := communication.BotMessage{Text: fmt.Sprintf("%d", cnt)}
			cnt++
			send(conn, &msg)
		}
	}()

	for {
		msg := receive(conn)
		fmt.Println(msg.String())
	}
}

func acquireCode(conn *websocket.Conn) {
	send(conn, &communication.BotMessage{Type: communication.MessageType_AcquiringCode})
	resp := receive(conn)
	fmt.Println("Linkage Code", resp.Text)
}

func send(conn *websocket.Conn, msg *communication.BotMessage) {
	msg.Timestamp = timestamppb.Now()
	buf, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	if err = conn.WriteMessage(websocket.BinaryMessage, buf); err != nil {
		panic(err)
	}
}

func receive(conn *websocket.Conn) *communication.UserMessage {
	_, buf, err := conn.ReadMessage()
	if err != nil {
		panic(err)
	}
	var msg communication.UserMessage
	if err = proto.Unmarshal(buf, &msg); err != nil {
		panic(err)
	}

	return &msg
}
