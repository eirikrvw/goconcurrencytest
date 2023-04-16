package main

import (
	"io"
	"net"

	"github.com/anthdm/hollywood/actor"
	"github.com/anthdm/hollywood/log"
)

type session struct {
	conn net.Conn
}

func newSession(conn net.Conn) actor.Producer {
	return func() actor.Receiver {
		return &session{
			conn: conn,
		}
	}
}

func (s *session) readLoop(c *actor.Context) {
	buf := make([]byte, 1024)
	for {
		n, err := s.conn.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorw("error reading data from client", log.M{"err": err})
			break
		}
		msg := buf[:n]
		c.Send(c.PID(), msg)

	}
	// For loop is broken
	c.Send(c.Parent(), &connRemove{pid: c.PID()})

}

func (s *session) Receive(c *actor.Context) {
	switch msg := c.Message().(type) {
	case actor.Initialized:
	case actor.Started:
		log.Infow("New connection", log.M{"addr": s.conn.RemoteAddr()})
		go s.readLoop(c)
	case actor.Stopped:
	case []byte:
		s.conn.Write(msg)
	}
}

type server struct {
	listenAddr string
	ln         net.Listener
	sessions   map[*actor.PID]net.Conn
}

type connAdd struct {
	pid  *actor.PID
	conn net.Conn
}

type connRemove struct {
	pid *actor.PID
}

func newServer(listenAddr string) actor.Producer {
	return func() actor.Receiver {
		return &server{
			listenAddr: listenAddr,
			sessions:   make(map[*actor.PID]net.Conn),
		}
	}
}

func (s *server) Receive(c *actor.Context) {
	switch msg := c.Message().(type) {
	case actor.Started:
		log.Infow("server started", log.M{"addr": s.listenAddr})
		go s.acceptLoop(c)
	case actor.Stopped:
	case actor.Initialized:
		ln, err := net.Listen("tcp", s.listenAddr)
		if err != nil {
			panic(err)

		}
		s.ln = ln
	case *connAdd:
		log.Tracew("added new connection to session map", log.M{"addr": msg.conn.RemoteAddr(), "pid": msg.pid})
		s.sessions[msg.pid] = msg.conn
	case *connRemove:
		log.Tracew("remove connection from session map", log.M{"pid": msg.pid})
		delete(s.sessions, msg.pid)
	}
}

func (s *server) acceptLoop(c *actor.Context) {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			log.Errorw("accept failed", log.M{"err": err})
			break
		}
		pid := c.SpawnChild(newSession(conn), "session", actor.WithTags(conn.RemoteAddr().String()))
		c.Send(c.PID(), &connAdd{
			pid:  pid,
			conn: conn,
		})

	}
}

func main() {
	e := actor.NewEngine()
	e.Spawn(newServer(":6000"), "server")

	<-make(chan struct{})
}
