package main

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	//   "math/rand"
	"net"
	"os"
	"os/signal"
	"time"

	proto "github.com/huin/mqtt"
	"github.com/nagae-memooff/config"
	"github.com/nagae-memooff/mqtt"
)

var host = flag.String("host", "192.168.100.236:1884", "服务器IP")
var laddr = flag.String("laddr", "192.168.100.217", "使用这个本地地址建立连接")
var user = flag.String("user", "server", "user")
var pass = flag.String("pass", "minxing123", "password")
var list_file = flag.String("list_file", "channel.list", "读取topic文件名")
var dump = flag.Bool("dump", false, "dump messages?")
var wait = flag.Int("wait", 100, "每个客户端建立连接前等待的毫秒数")
var pace = flag.Int("pace", 1, "每个客户端每发布一条消息时，平均等待的秒数")
var start_wait = flag.Int("start_wait", 50, "每个客户端发布第一条消息前等待的秒数")
var t = flag.Bool("t", false, "是否连接特定master客户端")
var tpace = flag.Int("tpace", 200, "特定master客户端每隔多少毫秒发送一条消息")
var ping_only = flag.Bool("ping_only", false, "只ping，不发消息")
var ssl = flag.Bool("ssl", true, "使用ssl")
var raddr *net.TCPAddr
var start_port = flag.Int("start_port", 1024, "本地起始端口")

var now_port = uint16(0)

// var publish = flag.Bool("publish", true, "是否发消息")
var close_wait = flag.Int("close_wait", 5, "中断测试时，等待mqtt服务器推送完毕的秒数")

var broadcast_topic = "/b/272d56f8515874d695f1a1a4db899b38"

// var broadcast_message_temp = `{"clients"}`
var publish = true

var teding_client_id = "fake_teding_recv"
var teding_topic = "/u/teding_topic"

// var payload proto.Payload
// var topic string

var recv_count uint64 = 0
var will_subscribe_count uint64 = 0
var subscribe_count uint64 = 0
var start_wait_time = time.Duration(*start_wait) * time.Second
var pace_time = (time.Duration)(*pace) * time.Second

func main() {
	flag.Parse()

	fmt.Println(*list_file)
	err := config.Parse(*list_file)
	if err != nil {
		log.Println("load config failed.\n" + err.Error())
		os.Exit(1)
	}

	raddr, err = net.ResolveTCPAddr("tcp", *host)
	if err != nil {
		log.Println(err)
	}

	//	if flag.NArg() != 2 {
	//		//     topic = "/u/2cKvvd_6yxOUh-r32p97riP0yew="
	//		payload = proto.BytesPayload([]byte("hello"))
	//	} else {
	//		topic = flag.Arg(0)
	//		payload = proto.BytesPayload([]byte(flag.Arg(1)))
	//	}

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c)
		//监听指定信号
		//signal.Notify(c, syscall.SIGHUP, syscall.SIGUSR2)

		//阻塞直至有信号传入
		for s := range c {
			if s.String() == "interrupt" {
				publish = false
				fmt.Printf("等待 %d 秒，看消息能不能发送完毕\n", *close_wait)
				time.Sleep((time.Duration)(*close_wait) * time.Second)
				//       fmt.Printf("send: %d, recv: %d\n", send_count, recv_count)
				os.Exit(0)
				panic("中断")
			}
		}
	}()
	//   i := 0

	if *t {
		go teding_client("master_8859", "/u/teding_topic")
	}
	conf := *config.GetModel()

	sleep_time := time.Duration(*wait) * time.Millisecond
	for client_id, topic := range conf {
		go client(client_id, topic)
		//     i++

		//     *conns--
		//     if *conns == 0 {
		//       break
		//     }
		time.Sleep(sleep_time)
	}

	// sleep forever
	<-make(chan bool)
}

func client(client_id, topic string) {
	//   log.Print("starting client ", i)
	defer func() {
		if err := recover(); err != nil {
			log.Printf("client_id: %s, 发生错误啦，重连！！！！！ %s\n", client_id, err)
			time.Sleep(5 * time.Second)
			client(client_id, topic)
		}
	}()

	send_count := 0

	// First, create the set of root certificates. For this example we only
	// have one. It's also possible to omit this in order to use the
	// default root set of the current operating system.

	var conn net.Conn

	if *ssl {
		now_port++
		var port uint16 = now_port + uint16(*start_port)
		laddr_string := fmt.Sprintf("%s:%d", *laddr, port)
		laddr, err := net.ResolveTCPAddr("tcp", laddr_string)
		if err != nil {
			log.Print("addr err:", err)
			panic(err)
		}
		d := &net.Dialer{LocalAddr: laddr}

		conn, err = tls.DialWithDialer(d, "tcp", *host, &tls.Config{
			//     conn, err = tls.Dial("tcp", *host, &tls.Config{
			InsecureSkipVerify: true,
		})
		if err != nil {
			panic("failed to connect: " + err.Error())
		}
	} else {
		//     conn, err = net.Dial("tcp", *host)
		now_port++
		var port uint16 = now_port + uint16(*start_port)
		laddr_string := fmt.Sprintf("%s:%d", *laddr, port)
		laddr, err := net.ResolveTCPAddr("tcp", laddr_string)
		if err != nil {
			log.Print("addr err:", err)
			panic(err)
		}

		conn, err = net.DialTCP("tcp", laddr, raddr)

		if err != nil {
			log.Print("dial: ", err)
			panic(err)
		}
	}

	cc := mqtt.NewClientConn(conn)
	cc.ClientId = client_id
	cc.KeepAliveTimer = 60
	cc.Dump = *dump

	will_subscribe_count++

	tmp_id := will_subscribe_count
	if publish {
		log.Printf("client %d ：尝试连接到： %s, %s\n", will_subscribe_count, client_id, topic)
		if err := cc.Connect(*user, *pass); err != nil {
			log.Fatalf("connect: %v\n", err)
			os.Exit(1)
		}

		//   half := int32(*pace / 2)

		_ = cc.Subscribe([]proto.TopicQos{
			{topic, proto.QosAtLeastOnce},
		})
		subscribe_count++
		log.Printf("已连接上： %d\n", subscribe_count)
	} else {
		return
	}
	//   fmt.Printf("suback: %#v\n", ack)
	if *dump {
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("ping 出错： %s\n", err)
				time.Sleep(10 * time.Second)
			}
		}()

		for {
			if publish {
				cc.Ping(&proto.PingReq{Header: proto.Header{}})
				time.Sleep(time.Minute)
			} else {
				return
			}
		}
	}()

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("出错： %s\n", err)
				time.Sleep(10 * time.Second)
			}
		}()

		//     for _ = range cc.Incoming {
		for incom := range cc.Incoming {
			time_now := time.Now().UnixNano()
			fmt.Printf("{payload: %s, now: %d },\n", incom.Payload, time_now)
			//       recv_count++
			//       if recv_count%100 == 0 {
			//         log.Printf("%s 接收消息: %d \n", client_id, recv_count)
			//       }
			//       fmt.Printf("%s\n", incom.Payload)
		}
	}()

	time.Sleep(start_wait_time)

	//   log.Printf("%d 哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈\n", will_subscribe_count)
	_ = tmp_id
	if *ping_only {
		<-make(chan bool)
	} else if tmp_id < 250 {
		//   if true {
		//   if false {
		for {
			if publish {
				//         payload := base64.StdEncoding.EncodeToString(([]byte)(fmt.Sprintf("hello, I'm %s, now message count is %d", client_id, send_count)))
				payload := base64.StdEncoding.EncodeToString(([]byte)(fmt.Sprintf("%d", time.Now().UnixNano())))
				message, err := json.Marshal(&BroadCastMessage{Clients: recv_qunliao, Payload: payload})
				//         fmt.Println((string)(message))
				if err != nil {
					log.Printf("marshal failed: %s \n", err)
				}
				marshaled_message := proto.BytesPayload(message)
				//         message_payload := proto.BytesPayload(base64.StdEncoding.EncodeToString(marshaled_message))

				cc.Publish(&proto.Publish{
					Header:    proto.Header{},
					TopicName: broadcast_topic,
					Payload:   marshaled_message,
				})
				log.Printf("%s 发送群聊 \n", client_id)
				send_count++
				time.Sleep(pace_time)
				//     sltime := rand.Int31n(half) - (half / 2) + int32(*pace)
			} else {
				return
			}
		}
	} else if tmp_id < 8000 {
		for {
			if publish {
				//         payload := proto.BytesPayload(fmt.Sprintf("hello, I'm %s, now message count is %d", client_id, send_count))
				payload := proto.BytesPayload(fmt.Sprintf("%d", time.Now().UnixNano()))
				//         payload := base64.StdEncoding.EncodeToString(([]byte)(fmt.Sprintf("%d", time.Now().UnixNano())))
				//         message_payload := proto.BytesPayload(base64.StdEncoding.EncodeToString(marshaled_message))

				cc.Publish(&proto.Publish{
					Header:    proto.Header{},
					TopicName: topic,
					Payload:   payload,
				})
				log.Printf("%s 发送私聊 \n", client_id)
				send_count++
				time.Sleep(pace_time)
				//     sltime := rand.Int31n(half) - (half / 2) + int32(*pace)
			} else {
				return
			}
		}
	}
}

func teding_client(client_id, topic string) {
	//   log.Print("starting client ", i)
	defer func() {
		if err := recover(); err != nil {
			log.Printf("发生错误啦，重连！！！！！ %s\n", err)
			time.Sleep(10 * time.Second)
			client(client_id, topic)
		}
	}()

	will_subscribe_count++

	var conn net.Conn
	var err error

	if *ssl {
		conn, err = tls.Dial("tcp", *host, &tls.Config{
			InsecureSkipVerify: true,
		})
		if err != nil {
			panic("failed to connect: " + err.Error())
		}
	} else {
		conn, err = net.Dial("tcp", *host)
		if err != nil {
			log.Print("dial: ", err)
			panic(err)
		}
	}

	cc := mqtt.NewClientConn(conn)
	cc.ClientId = client_id
	cc.KeepAliveTimer = 60
	cc.Dump = *dump

	if publish {
		log.Printf("teding_client %s ：尝试连接到：%s\n", client_id, topic)
		if err := cc.Connect(*user, *pass); err != nil {
			log.Fatalf("connect: %v\n", err)
			os.Exit(1)
		}

		//   half := int32(*pace / 2)

		_ = cc.Subscribe([]proto.TopicQos{
			{topic, proto.QosAtLeastOnce},
		})
		subscribe_count++
		log.Printf("已连接上： %d\n", subscribe_count)
	} else {
		return
	}
	//   fmt.Printf("suback: %#v\n", ack)
	if *dump {
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("ping 出错： %s\n", err)
				time.Sleep(10 * time.Second)
			}
		}()

		for {
			if publish {
				cc.Ping(&proto.PingReq{Header: proto.Header{}})
				time.Sleep(time.Minute)
			} else {
				return
			}
		}
	}()

	go func() {
		//     for _ = range cc.Incoming {
		for incom := range cc.Incoming {
			//TODO: 解析收到消息时的时间，把消息里的时间和现在的时间一起打印
			time_now := time.Now().UnixNano()
			fmt.Printf("{payload: %s, now: %d },\n", incom.Payload, time_now)
			//       fmt.Println(pub.Payload)
		}
	}()

	for {
		if publish {
			//     if false {
			var send_topic string
			var payload proto.Payload

			//       r := rand.Int31n(5)
			r := 4
			switch r {
			case 0:
				send_topic = topic
				payload = proto.BytesPayload(fmt.Sprintf("%d", time.Now().UnixNano()))
			case 1:
				send_topic = broadcast_topic
				msg := base64.StdEncoding.EncodeToString(([]byte)(fmt.Sprintf("%d", time.Now().UnixNano())))
				message, err := json.Marshal(&BroadCastMessage{Clients: small_qunliao, Payload: msg})
				//         fmt.Println((string)(message))
				if err != nil {
					log.Printf("marshal failed: %s \n", err)
				}
				payload = proto.BytesPayload(message)
			case 2:
				send_topic = broadcast_topic
				msg := base64.StdEncoding.EncodeToString(([]byte)(fmt.Sprintf("%d", time.Now().UnixNano())))
				message, err := json.Marshal(&BroadCastMessage{Clients: middle_qunliao, Payload: msg})
				//         fmt.Println((string)(message))
				if err != nil {
					log.Printf("marshal failed: %s \n", err)
				}
				payload = proto.BytesPayload(message)
			case 3:
				send_topic = broadcast_topic
				msg := base64.StdEncoding.EncodeToString(([]byte)(fmt.Sprintf("%d", time.Now().UnixNano())))
				message, err := json.Marshal(&BroadCastMessage{Clients: big_qunliao, Payload: msg})
				//         fmt.Println((string)(message))
				if err != nil {
					log.Printf("marshal failed: %s \n", err)
				}
				payload = proto.BytesPayload(message)
			case 4:
				send_topic = broadcast_topic
				msg := base64.StdEncoding.EncodeToString(([]byte)(fmt.Sprintf("%d", time.Now().UnixNano())))
				message, err := json.Marshal(&BroadCastMessage{Clients: super_big_qunliao, Payload: msg})
				//         fmt.Println((string)(message))
				if err != nil {
					log.Printf("marshal failed: %s \n", err)
				}
				payload = proto.BytesPayload(message)
			default:
			}

			cc.Publish(&proto.Publish{
				Header:    proto.Header{},
				TopicName: send_topic,
				Payload:   payload,
			})

			//     sltime := rand.Int31n(half) - (half / 2) + int32(*pace)
			time.Sleep((time.Duration)(*tpace) * time.Millisecond)
		} else {
			return
		}
	}
}

type BroadCastMessage struct {
	Clients []string `json:"clients"`
	Payload string   `json:"payload"`
}
