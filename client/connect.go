package main

import (
	"net"
	"os"
	"sync"
	"time"
)

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
	println("[", connId, "]", "connected to server")

	for i := 1; i <= 5; i++ {
		_, err = conn.Write([]byte(strEcho))
		if err != nil {
			println("[", connId, "]", "Write to server failed:", err.Error())
			os.Exit(1)
		}

		//println("[",connId,"]","write to server = ", strEcho)

		reply := make([]byte, 1024)

		_, err = conn.Read(reply)
		if err != nil {
			println("[", connId, "]", "Write to server failed:", err.Error())
			os.Exit(1)
		}

		//println("[",connId,"]","reply from server=", string(reply))
		time.Sleep(time.Second)
	}

	conn.Close()
}

func main() {
	strEcho := "Halo"
	servAddr := "localhost:6000"
	numconn := 1000
	var wg sync.WaitGroup

	for connId := 1; connId <= numconn; connId++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			client(strEcho, servAddr, id)
		}(connId)
	}
	wg.Wait()
}
