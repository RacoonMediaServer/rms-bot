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
	addr := flag.String("ip", "127.0.0.1", "Host of the bot server")
	port := flag.Int("port", 8080, "Port of the bot server")
	flag.Parse()

	u := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%d", *addr, *port),
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
	msg.Buttons = append(msg.Buttons, &communication.Button{
		Title:   "One",
		Command: "1",
	})
	msg.Buttons = append(msg.Buttons, &communication.Button{
		Title:   "Two",
		Command: "2",
	})
	msg.KeyboardStyle = communication.KeyboardStyle_Message
	msg.Attachment = &communication.Attachment{
		Type:     communication.Attachment_PhotoURL,
		MimeType: "",
		Content:  []byte("https://external-content.duckduckgo.com/iu/?u=https%3A%2F%2Fimages.radio.com%2Faiu-media%2Fgettyimages-91702654-23d3a585-1cd4-4051-9536-729eaba3e179.jpg&f=1&nofb=1&ipt=7b9e64fb346221340dd35e3f728fc6df90b10f4cb85fc7ab91f1c01da446d2ac&ipo=images"),
	}
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
