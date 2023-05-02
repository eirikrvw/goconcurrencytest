package main

import (
    "runtime/debug"

    "github.com/firstrow/tcp_server"
)

func main() {
	debug.SetMaxThreads(1000000)
	server := tcp_server.New("localhost:6000")

	server.OnNewClient(func(c *tcp_server.Client) {
		// new client connected
		// lets send some message
		c.Send("Hello")
	})
	server.OnNewMessage(func(c *tcp_server.Client, message string) {
		// new message received
	})
	server.OnClientConnectionClosed(func(c *tcp_server.Client, err error) {
		// connection with client lost
	})

	server.Listen()
}
