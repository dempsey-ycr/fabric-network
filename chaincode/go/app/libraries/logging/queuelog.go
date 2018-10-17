package logging

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"path/filepath"
	"runtime"
	"time"
)

type QueueLogBackend struct {
}

type LogInfor struct {
	IP      string
	Method  string
	URI     string
	Version string
}

type LogServer struct {
	Addr  string
	AppID string
	On    bool
	Port  string
}

const (
	statusOnline int = iota
	statusDisconnect
	statusTrying
	statusConnecting
)

var gConnStatus int
var gLogChan = make(chan string, 10000)
var gLogInfor *LogInfor
var gLogServer *LogServer
var gPrivateConn net.Conn

func NewQueueLogBackend() *QueueLogBackend {
	return &QueueLogBackend{}
}

func (this *QueueLogBackend) IsEnabledFor(level Level, module string) bool {
	return true
}

func (this *QueueLogBackend) Log(level Level, calldepth int, rec *Record) error {
	_, file, line, _ := runtime.Caller(calldepth + 1)
	switch level {
	case REDIS:
		fallthrough
	case SQL:
		_, file, line, _ = runtime.Caller(calldepth + 3)
	}

	content := ""
	formatted := rec.Formatted(calldepth + 1)
	time := rec.Time.Format("2006-01-02 15:04:05")

	if gLogInfor == nil {
		content = fmt.Sprintf("%s[%s] %s  %s:%d %s \033[0m", colors[level], level, time, filepath.Base(file), line, formatted)
	} else {
		content = fmt.Sprintf("%s[%s] %s | %s | %s %s | %s:%d | %s \033[0m", colors[level], level, time, gLogInfor.IP, gLogInfor.Method, gLogInfor.URI, filepath.Base(file), line, formatted)
	}

	if gLogServer != nil && gLogServer.On && gConnStatus == statusOnline {
		gLogChan <- content
	}

	//f, err := os.OpenFile("./log/log.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0664)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//defer f.Close()
	//
	//_, err = f.Write([]byte(content + "\n"))
	//if err != nil {
	//	fmt.Println(err)
	//}
	fmt.Println(content)
	return nil
}

func (this *QueueLogBackend) SetLog(aid string, addr string, port string, on bool) {
	gLogServer = &LogServer{AppID: aid, Addr: addr, Port: port, On: on}
}

func (this *QueueLogBackend) SetLogInfo(IP string, URI string, method string) {

	gLogInfor = &LogInfor{IP: IP, URI: URI, Method: method}
}

func (this *QueueLogBackend) ListenLog(aid string, addr string, port string, on bool) {
	if !on {
		fmt.Println("Log server not enabled, give up!")
		return
	}

	this.SetLog(aid, addr, port, on)

	go func() {
		for {
			data, ok := <-gLogChan
			if ok {
				jsn := map[string]interface{}{
					"app_id": gLogServer.AppID,
					"data":   data,
				}
				rec, _ := json.Marshal(jsn)

				_con := getConnect()
				_con.Write([]byte(string(rec) + "\n"))
			}
		}
	}()
}

// --------------------------------------------------------------------------------

func getConnect() net.Conn {
	if gPrivateConn == nil {
		gConnStatus = statusConnecting
		con, err := net.DialTimeout("tcp", gLogServer.Addr+":"+gLogServer.Port, 30*time.Second)
		if err != nil {
			i := 0
			for {
				gConnStatus = statusTrying
				fmt.Println("Try to Log server reconnect", i, "...")
				time.Sleep(2 * time.Second)
				i++

				con, err := net.DialTimeout("tcp", gLogServer.Addr+":"+gLogServer.Port, 30*time.Second)
				if err == nil {
					gPrivateConn = con
					gConnStatus = statusOnline
					go heart(gPrivateConn)

					Notice("Great! I am connection server!")
					break
				}
			}
		} else {
			gPrivateConn = con
			gConnStatus = statusOnline
			go heart(gPrivateConn)
		}
	}

	return gPrivateConn
}

func heart(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		_, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Disconnected Log server | ", err)
			gPrivateConn = nil
			gConnStatus = statusDisconnect

			go getConnect()
			break
		}
	}
}
