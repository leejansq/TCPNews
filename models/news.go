package models

/*
注册消息结构体:
	typedef struct
	{
		int uidLen;
		char uid[28];
	}RegMsg;

通知消息结构体：
	typedef struct
	{
		char  msgType;        // 1:通知消息， 2: 心跳消息
		char  cmdType;        //针对通知消息，1：start, 2: stop
		short heartbeatTime; //心跳间隔时间,每次必须带上
		char reserved[28];
	}NotifyMsg;


流程:
	1.  设备起来之后首先发送RegMsg，
	2.  服务器回复NotifyMsg（心跳类型消息）
	3.  设备默认超时时间设置为1分钟
	4.  心跳超时，未收到心跳，执行步骤1*/

import (
	"github.com/astaxie/beego"
	//"log"
	"net"
	//"runtime"
	"strings"
	"time"
)

var (
	udpConn  *net.UDPConn
	Map_con  map[string]*XX = map[string]*XX{}
	chtoken                 = make(chan bool, 1000)
	Apptime  int16          = 255
	Duration time.Duration  = 10 * time.Second
)

type XX struct {
	con *net.UDPAddr
	uid string
	Ch  chan int8
}

func NotifyMsg(msgType int8, cmdType int8, heartbeatTime int16) []byte {
	bs := make([]byte, 32)

	if msgType == 1 {
		bs[0] = 1
	} else {
		bs[0] = 2
	}

	if cmdType == 1 {
		bs[1] = 1
	} else {
		bs[1] = 2
	}
	bs[3] = byte(heartbeatTime)
	bs[2] = byte(heartbeatTime >> 8)

	return bs

}
func tokenIssue() {

	for {
		if len(chtoken) < 1000 {
			for len(chtoken) < 1000 {
				chtoken <- true
			}
			time.Sleep(time.Second)
		} else {
			time.Sleep(time.Second)
			continue
		}
	}

}
func (c *XX) Write(b []byte) error {
	//beego.Info(b, len(chtoken))
	//beego.fmt.Println(b, len(chtoken))
	//<-chtoken
	_, err := udpConn.WriteToUDP(b, c.con)
	//_, err := c.con.Write(b)
	beego.Info(c.con, b, len(Map_con))
	if err != nil {
		//fmt.Println(err)
		return err
	}
	return nil
}

func (c *XX) Brocast() {

	defer func() {
		if r := recover(); r != nil {
			beego.Info(r)
		}
	}()
	defer func() {
		delete(Map_con, c.uid)
		//fmt.Println(c.uid)
	}()
	c.Write(NotifyMsg(2, 1, 2*Apptime))
	for {
		select {
		case msg := <-c.Ch:
			//if msg == 0 {
			//	c.con.Close()
			//	break
			//}
			err := c.Write(NotifyMsg(1, msg, 2*Apptime))
			if err != nil {
				break
			}
		case <-time.After(Duration):
			err := c.Write(NotifyMsg(2, 1, 2*Apptime))
			if err != nil {

				return
			}
		}

		//fmt.Println(c.msg)
		//time.Sleep(time.Duration(duration))
	}
}
func (c *XX) distroy() {
	c.Ch <- 0
	//delete(Map_con, c.uid)
}
func ResetTime(i int64) {
	//tm, _ := beego.AppConfig.Int("Time")
	Apptime = int16(i)
	Duration = time.Duration(Apptime) * time.Second
	//Apptime= i
	//Duration = time.Duration(Apptime) * time.Second
	beego.Info("heartbeatTime Reset:", Duration)
}
func init() {
	//go func() {
	//	for {
	//		fmt.Println(len(Map_con))
	//		time.Sleep(time.Duration(time.Second))
	//	}

	//}()
	go func() {
		//runtime.GOMAXPROCS(runtime.NumCPU())
		//fmt.Println("start...")
		//beego.AppConfig.Int()
		tm, _ := beego.AppConfig.Int("Time")
		Apptime = int16(tm)
		Duration = time.Duration(Apptime) * time.Second
		//ResetTime()

		//go tokenIssue()
		udpaddr, _ := net.ResolveUDPAddr("udp", ":8603")
		//fmt.Printf("%#v\n", udpaddr)
		beego.Info("heartbeatTime:", Duration, "  Listen on:", udpaddr)
		udpConn, _ = net.ListenUDP("udp", udpaddr)
		for {
			b := make([]byte, 32)
			n, readdr, _ := udpConn.ReadFromUDP(b)
			if n < 20 {
				continue
			}
			uid := string(b[4:])
			uid = strings.TrimRight(uid, string(byte(0)))
			beego.Info(uid, " >>>login")
			if Map_con[uid] == nil {
				ch := make(chan int8)
				t_xx := &XX{readdr, uid, ch}
				Map_con[uid] = t_xx
				go t_xx.Brocast()
			} else {
				Map_con[uid].con = readdr
			}

		}

		//fmt.Println("Hello World!")
	}()

}
