package server

import (
	"log"
	"meta/meta-server/client"
	"net"
	"net/url"

	packet "github.com/surgemq/message"
)

type MQServer struct {
	cm       *client.ClientManager
	listener *net.TCPListener
	limiter  string
}

func NewMQServer() *MQServer {
	cm := client.NewClientManager()
	return &MQServer{
		cm: cm,
	}
}

func (this *MQServer) Start(uri string) {
	u, err := url.Parse(uri)
	if err != nil {
		panic(err)
	}

	addr, err := net.ResolveTCPAddr(u.Scheme, u.Host)
	if err != nil {
		panic(err)
	}
	this.listener, _ = net.ListenTCP(u.Scheme, addr)
	defer this.listener.Close()
	log.Println("Accepting connections at:", uri)
	for {
		conn, err := this.listener.AcceptTCP()
		if err != nil {
			continue
		}
		log.Println("Handle connection ", conn.RemoteAddr().String(), this.cm.Size())
		go this.handleConnection(conn)
	}
}

func (this *MQServer) Stop() {
	this.listener.Close()
}

func (this *MQServer) handleConnection(conn *net.TCPConn) {
	client := client.NewMetaClient(conn, func(producer *client.MetaClient, message packet.Message) {
		msg := message.(*packet.PublishMessage)
		for _, consumer := range this.cm.CloneMap() {
			if consumer == nil || consumer.IsClosed {
				continue
			}
			for _, topic := range consumer.Topics {
				if topic == string(msg.Topic()) {
					consumer.OutChan <- msg
				}
			}
		}
	})
	client.Start()
	this.cm.AddClient(client)
}

func HeartBeat(conn net.TCPConn, timeout int) {

}