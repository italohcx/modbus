package modbus

import (
	"fmt"
	"io"
	"net"
	"time"
)

// RTUOverTCPClientHandler envia frames RTU (com CRC) via conexão TCP.
// Ele implementa a interface ClientHandler.
type RTUOverTCPClientHandler struct {
	rtuPackager // embutido, adiciona Encode, Decode, Verify
	Address     string
	Timeout     time.Duration
	SlaveId     byte
	Conn        net.Conn
	closed      bool
}

// Construtor
func NewRTUOverTCPClientHandler(address string) *RTUOverTCPClientHandler {
	return &RTUOverTCPClientHandler{
		Address: address,
		Timeout: 5 * time.Second, // default
	}
}

// Connect conecta ao servidor TCP
func (h *RTUOverTCPClientHandler) Connect() error {
	if h.Conn != nil {
		return nil // já conectado
	}
	conn, err := net.DialTimeout("tcp", h.Address, h.Timeout)
	if err != nil {
		return err
	}
	h.Conn = conn
	h.closed = false
	return nil
}

// Close fecha a conexão
func (h *RTUOverTCPClientHandler) Close() error {
	h.closed = true
	if h.Conn != nil {
		return h.Conn.Close()
	}
	return nil
}

// Encode envia o frame com CRC já incluso
func (h *RTUOverTCPClientHandler) Send(request []byte) ([]byte, error) {
	if h.closed || h.Conn == nil {
		return nil, fmt.Errorf("connection is closed")
	}

	if h.Timeout > 0 {
		h.Conn.SetDeadline(time.Now().Add(h.Timeout))
	}

	// Escreve a requisição
	_, err := h.Conn.Write(request)
	if err != nil {
		return nil, err
	}

	// Lê a resposta (tamanho máximo de um frame RTU típico)
	buf := make([]byte, 256)
	n, err := h.Conn.Read(buf)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return buf[:n], nil
}

// SlaveId retorna o ID do escravo
func (h *RTUOverTCPClientHandler) GetSlave() byte {
	return h.SlaveId
}

// SetSlave define o ID do escravo
func (h *RTUOverTCPClientHandler) SetSlave(id byte) {
	h.SlaveId = id
}
