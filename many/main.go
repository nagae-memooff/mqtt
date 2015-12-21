package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"time"

	proto "github.com/huin/mqtt"
	"github.com/nagae-memooff/config"
	"github.com/nagae-memooff/mqtt"
)

var conns = flag.Int("conns", 1, "how many conns (0 means infinite)")
var host = flag.String("host", "192.168.100.236:1884", "hostname of broker")
var user = flag.String("user", "server", "user")
var pass = flag.String("pass", "minxing123", "password")
var dump = flag.Bool("dump", false, "dump messages?")
var wait = flag.Int("wait", 100, "每个客户端建立连接前等待的毫秒数")
var pace = flag.Int("pace", 1, "每个客户端每发布一条消息时，平均等待的秒数")
var up_rd = flag.Int("up_rd", 3, "有十分之几的概率是发回执")
var close_wait = flag.Int("close_wait", 1, "中断测试时，等待mqtt服务器推送完毕的秒数")

var feedback_topic = "/u"
var broadcast_topic = "/b/272d56f8515874d695f1a1a4db899b38"

// var broadcast_message_temp = `{"clients"}`
var publish bool = true

var teding_client_id = "fake_teding_recv"
var teding_topic = "/u/teding_topic"

// var payload proto.Payload
// var topic string

var recv_count uint64 = 0
var will_subscribe_count uint64 = 0
var subscribe_count uint64 = 0

func main() {
	err := config.Parse("channel.list")
	if err != nil {
		log.Println("load config failed." + err.Error())
		os.Exit(1)
	}

	flag.Parse()

	//	if flag.NArg() != 2 {
	//		//     topic = "/u/2cKvvd_6yxOUh-r32p97riP0yew="
	//		payload = proto.BytesPayload([]byte("hello"))
	//	} else {
	//		topic = flag.Arg(0)
	//		payload = proto.BytesPayload([]byte(flag.Arg(1)))
	//	}

	if *conns == 0 {
		*conns = -1
	}

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c)
		//监听指定信号
		//signal.Notify(c, syscall.SIGHUP, syscall.SIGUSR2)

		//阻塞直至有信号传入
		s := <-c
		if s.String() == "interrupt" {
			publish = false
			fmt.Printf("等待 %d 秒，看消息能不能发送完毕", *close_wait)
			time.Sleep((time.Duration)(*close_wait) * time.Second)
			//       fmt.Printf("send: %d, recv: %d\n", send_count, recv_count)
			os.Exit(0)
		}
	}()
	//   i := 0

	go teding_client("fake_teding_recv", "/u/teding_topic")
	conf := *config.GetModel()

	for client_id, topic := range conf {
		go client(client_id, topic)
		//     i++

		//     *conns--
		//     if *conns == 0 {
		//       break
		//     }
		time.Sleep(time.Duration(*wait) * time.Millisecond)
	}

	// sleep forever
	<-make(chan struct{})
}

func client(client_id, topic string) {
	//   log.Print("starting client ", i)
	defer func() {
		if err := recover(); err != nil {
			log.Printf("发生错误啦，重连！！！！！ %s", err)
			time.Sleep(10 * time.Second)
			client(client_id, topic)
		}
	}()

	will_subscribe_count++
	send_count := 0
	conn, err := net.Dial("tcp", *host)
	if err != nil {
		log.Print("dial: ", err)
		panic(err)
	}
	cc := mqtt.NewClientConn(conn)
	cc.ClientId = client_id
	cc.KeepAliveTimer = 60
	cc.Dump = *dump

	if publish {
		log.Printf("client %d ：尝试连接到： %s, %s\n", will_subscribe_count, client_id, topic)
		if err := cc.Connect(*user, *pass); err != nil {
			log.Fatalf("connect: %v\n", err)
			os.Exit(1)
		}

		//   half := int32(*pace / 2)

		_ = cc.Subscribe([]proto.TopicQos{
			{topic, proto.QosAtMostOnce},
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
		for {
			if publish {
				cc.Ping(&proto.PingReq{Header: proto.Header{}})
				time.Sleep(1 * time.Minute)
			} else {
				return
			}
		}
	}()

	go func() {
		for _ = range cc.Incoming {
			//     for incom := range cc.Incoming {
			recv_count++
			log.Printf("%s 接收消息: %d \n", client_id, recv_count)
			//       fmt.Println(pub.Payload)
		}
	}()

	var feedback_payload = proto.BytesPayload(`53:ewogICJkYXRhIiA6IFsKICAgIHsKICAgICAgImNvbnZlcnNhdGlvbl9pZCIgOiAyMDAyMDE5MCwKICAgICAgIm1lc3NhZ2VfaWQiIDogNzE2LAogICAgICAidXNlcl9pZCIgOiA1MwogICAgfQogIF0sCiAgInR5cGUiIDogInJlY2VpcHQiCn0=`)
	time.Sleep(50 * time.Second)

	if will_subscribe_count > 1000 {
		for {
			if publish {
				//       payload := proto.BytesPayload([]byte(fmt.Sprintf("hello: %s, %d", client_id, send_count)))
				rd := rand.Int31n(10)
				if (int)(rd) < *up_rd {
					cc.Publish(&proto.Publish{
						Header:    proto.Header{},
						TopicName: feedback_topic,
						Payload:   feedback_payload,
					})
					log.Printf("%s 发送回执 \n", client_id)
				} else {
					//         p := proto.BytesPayload([]byte(fmt.Sprintf("hello: %s, %d", client_id, send_count))))
					//         message := proto.BytesPayload(base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("hello: %s, %d", client_id, send_count))))

					payload := base64.StdEncoding.EncodeToString(([]byte)(fmt.Sprintf("hello, I'm %s, now message count is %d", client_id, send_count)))
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
				}
				send_count++
				time.Sleep((time.Duration)(rand.Int31n((int32)(*pace))) * time.Second)
				//     sltime := rand.Int31n(half) - (half / 2) + int32(*pace)
			} else {
				return
			}
		}
	} else {
		for {
			if publish {
				//       payload := proto.BytesPayload([]byte(fmt.Sprintf("hello: %s, %d", client_id, send_count)))
				rd := rand.Int31n(10)
				if (int)(rd) < *up_rd {
					cc.Publish(&proto.Publish{
						Header:    proto.Header{},
						TopicName: feedback_topic,
						Payload:   feedback_payload,
					})
					log.Printf("%s 发送回执 \n", client_id)
				} else {
					message := base64.StdEncoding.EncodeToString(([]byte)(fmt.Sprintf("hello, I'm %s, now message count is %d", client_id, send_count)))
					payload := proto.BytesPayload(message)
					//         message_payload := proto.BytesPayload(base64.StdEncoding.EncodeToString(marshaled_message))

					cc.Publish(&proto.Publish{
						Header:    proto.Header{},
						TopicName: topic,
						Payload:   payload,
					})
					log.Printf("%s 发送私聊 \n", client_id)
				}
				send_count++
				time.Sleep((time.Duration)(rand.Int31n((int32)(*pace))) * time.Second)
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
			log.Printf("发生错误啦，重连！！！！！ %s", err)
			time.Sleep(10 * time.Second)
			client(client_id, topic)
		}
	}()

	will_subscribe_count++
	conn, err := net.Dial("tcp", *host)
	if err != nil {
		log.Print("dial: ", err)
		panic(err)
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
			{topic, proto.QosAtMostOnce},
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
		for {
			if publish {
				cc.Ping(&proto.PingReq{Header: proto.Header{}})
				time.Sleep(1 * time.Minute)
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
			var send_topic string
			var payload proto.Payload

			r := rand.Int31n(5)
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

			time.Sleep(1 * time.Second)
			//     sltime := rand.Int31n(half) - (half / 2) + int32(*pace)
		} else {
			return
		}
	}
}

type BroadCastMessage struct {
	Clients []string `json:"clients"`
	Payload string   `json:"payload"`
}
