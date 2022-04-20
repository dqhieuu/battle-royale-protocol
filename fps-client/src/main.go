package main

import (
	"bytes"
	_ "embed"
	g "github.com/AllenDang/giu"
	"github.com/xtaci/kcp-go"
	"log"
	"math/rand"
	"time"
)

var clientsToAdd int32

func loop() {
	g.SingleWindow().Layout(
		g.Style().SetFontSize(32).To(g.Label("Client spawner cho server PWNBG")),
		g.Row(
			g.Label("Số lượng client thêm vào"),
			g.InputInt(&clientsToAdd).Size(60),
			g.Button("Thêm client mới"),
		),
	)
}

func clientToGameServer(url string) {
	time.Sleep(time.Second)

	// dial to the echo server

	if sess, err := kcp.Dial(url); err == nil {
		buffer := make([]byte, 1024)

		sess.Write([]byte("hello"))

		for {
			if _, err := sess.Read(buffer); err == nil {
				var byteBuffer bytes.Buffer
				byteBuffer.Write(buffer)
				log.Println("recv:", byteBuffer.String())
			} else {
				log.Fatal(err)
			}
		}
	} else {
		log.Println(err)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	go clientToGameServer("192.168.1.23:19001")
	wnd := g.NewMasterWindow("Canvas", 500, 300, 0)

	wnd.Run(loop)
}
