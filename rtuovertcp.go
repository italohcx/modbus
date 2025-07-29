package modbus

import (
	"io"
	"net"
	"sync"
	"time"
)

// RTUOverTCPClientHandler envia frames RTU (com CRC) por uma conexão TCP.
type RTUOverTCPClientHandler struct {
	rtuPackager
	rtuTCPTransporter
}

// Novo handler
func NewRTUOverTCPClientHandler(address string) *RTUOverTCPClientHandler {
	return &RTUOverTCPClientHandler{
		rtuPackager: rtuPackager{},
		rtuTCPTransporter: rtuTCPTransporter{
			Address: address,
			Timeout: 2 * time.Second,
		},
	}
}

// TCPClient creates TCP client with default handler and given connect string.
func RtuOverTcpClient(address string) Client {
	handler := NewRTUOverTCPClientHandler(address)
	return NewClient(handler)
}

// Implementa a camada de transporte TCP usando frames RTU
type rtuTCPTransporter struct {
	conn    net.Conn
	Address string
	Timeout time.Duration
	mu      sync.Mutex
}

func (t *rtuTCPTransporter) connect() error {
	if t.conn != nil {
		return nil
	}
	conn, err := net.DialTimeout("tcp", t.Address, t.Timeout)
	if err != nil {
		return err
	}
	t.conn = conn
	return nil
}

func (t *rtuTCPTransporter) Send(aduRequest []byte) ([]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if err := t.connect(); err != nil {
		return nil, err
	}

	_, err := t.conn.Write(aduRequest)
	if err != nil {
		return nil, err
	}

	// Recebe a resposta
	var buf [rtuMaxSize]byte
	t.conn.SetReadDeadline(time.Now().Add(t.Timeout))

	n, err := io.ReadAtLeast(t.conn, buf[:], rtuMinSize)
	if err != nil {
		return nil, err
	}

	function := aduRequest[1]
	functionFail := aduRequest[1] | 0x80
	bytesToRead := calculateResponseLength(aduRequest)

	if buf[1] == function && n < bytesToRead {
		n1, err := io.ReadFull(t.conn, buf[n:bytesToRead])
		n += n1
		if err != nil {
			return nil, err
		}
	} else if buf[1] == functionFail && n < rtuExceptionSize {
		n1, err := io.ReadFull(t.conn, buf[n:rtuExceptionSize])
		n += n1
		if err != nil {
			return nil, err
		}
	}

	return buf[:n], nil
}

// closeLocked closes current connection. Caller must hold the mutex before calling this method.
func (mb *rtuTCPTransporter) close() (err error) {
	if mb.conn != nil {
		err = mb.conn.Close()
		mb.conn = nil
	}
	return
}
