package rtmp

import (
	//"log"
	"fmt"
	"net"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/gwuhaolin/livego/protocol/amf"
)

// Server returns a new RTMP server side conncetion
func Server(conn net.Conn, config *Config) *Conn {
	c := &Conn{
		conn:   conn,
		config: config,
	}
	c.handshakeFn = c.serverHandshake

	//TODO: config
	c.localChunksize = 128
	c.remoteChunkSize = 128
	c.localWindowAckSize = 2500000
	c.remoteWindowAckSize = 250000

	c.chunks = make(map[uint32]*ChunkStream)
	c.amfDecoder = &amf.Decoder{}
	c.amfEncoder = &amf.Encoder{}

	c.logger = config.logger

	return c
}

// Client returns a new RTMP client side connection
func Client(conn net.Conn, config *Config) *Conn {
	c := &Conn{
		conn:     conn,
		config:   config,
		isClient: true,
	}
	c.handshakeFn = c.clientHandshake
	return c
}

type listener struct {
	net.Listener
	config *Config
}

func (l *listener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return Server(c, l.config), nil
}

func NewListener(inner net.Listener, config *Config) net.Listener {
	l := new(listener)
	l.Listener = inner
	l.config = config
	return l
}

func Listen(network, laddr string, config *Config) (net.Listener, error) {
	l, err := net.Listen(network, laddr)
	if err != nil {
		return nil, err
	}
	return NewListener(l, config), nil
}

func ListenAndServe(network, laddr string, config *Config) error {
	if config.logger == nil {
		logger := log.NewLogfmtLogger(os.Stdout)
		logger = log.With(logger, "time", log.DefaultTimestamp)
		logger = log.With(logger, "caller", log.DefaultCaller)
		config.logger = logger
	}

	l, err := Listen(network, laddr, config)
	if err != nil {
		return err
	}

	_ = config.logger.Log("event", "rtmp listen", "addr", fmt.Sprintf("%s[%s]", l.Addr().String(), l.Addr().Network()))

	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}

		go conn.(*Conn).Serve()
	}
}
