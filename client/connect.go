package main

import (
	"net"
	"os"
	"sync"
	"time"
	"fmt"
	"runtime/debug"
	"flag"
)

type single struct {
    mu     sync.Mutex
    values map[string]int64
}

var counters = single{
    values: make(map[string]int64),
}

func (s *single) Get(key string) int64 {
    s.mu.Lock()
    defer s.mu.Unlock()
    return s.values[key]
}

func (s *single) Incr(key string) int64 {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.values[key]++
    return s.values[key]
}

func (s *single) Dec(key string) int64 {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.values[key]--
    return s.values[key]
}

func printCounter() {
    for {
        fmt.Println("counted: ",counters.Get("client"))
        time.Sleep(100* time.Millisecond)
    }
}

func client(strEcho string, servAddr string, connId int) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	if err != nil {
		println("[", connId, "]", "ResolveTCPAddr failed:", err.Error())
		os.Exit(1)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		println("[", connId, "]", "Dial failed:", err.Error())
		os.Exit(1)
	}

        err = conn.SetKeepAlive(true)
	if err != nil {
		println("[", connId, "]", "Setting tcp keepalive failed:", err.Error())
	}
	println("[", connId, "]", "connected to server")
        counters.Incr("client")

	for i := 1; i <= 5; i++ {
		_, err = conn.Write([]byte(strEcho))
		if err != nil {
			println("[", connId, "]", "Write to server failed:", err.Error())
			//os.Exit(1)
		}

		//println("[",connId,"]","write to server = ", strEcho)

		reply := make([]byte, 1024)

		_, err = conn.Read(reply)
		if err != nil {
			println("[", connId, "]", "Write to server failed:", err.Error())
			//os.Exit(1)
		}

		//println("[",connId,"]","reply from server=", string(reply))
		time.Sleep(time.Second)
	}
	println("[", connId, "]", "Disconnected from server")
        counters.Dec("client")
	conn.Close()
}

func main() {
	debug.SetMaxThreads(1000000)
	strEcho := "Hallo"
	var servAddr = flag.String("addr", "10.0.80.4:6000", "Server to connect to")
	var numconn = flag.Int("n", 25000, "Number of connections")
	flag.Parse()
	//servAddr := "10.0.80.4:6000"
	//numconn := 25000
	var wg sync.WaitGroup

        go printCounter() 


	for connId := 1; connId <= *numconn; connId++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			client(strEcho, *servAddr, id)
		}(connId)
	}
	wg.Wait()
}
