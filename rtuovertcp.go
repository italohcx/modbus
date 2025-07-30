package modbus

import (
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
	h := &RTUOverTCPClientHandler{
		Address: address,
		Timeout: 5 * time.Second,
	}
	h.rtuPackager.SlaveId = h.SlaveId
	return h
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
func (h *RTUOverTCPClientHandler) Send(pduRequest []byte) ([]byte, error) {
	// monta o frame com CRC
	aduRequest, err := h.rtuPackager.Encode(&ProtocolDataUnit{
		FunctionCode: pduRequest[0],
		Data:         pduRequest[1:],
	})
	if err != nil {
		return nil, err
	}

	// envia pela conexão TCP
	if h.Timeout != 0 {
		h.Conn.SetDeadline(time.Now().Add(h.Timeout))
	}
	_, err = h.Conn.Write(aduRequest)
	if err != nil {
		return nil, err
	}

	// lê resposta
	buf := make([]byte, 256)
	n, err := h.Conn.Read(buf)
	if err != nil {
		return nil, err
	}

	aduResponse := buf[:n]

	// decodifica e valida CRC
	pduResponse, err := h.rtuPackager.Decode(aduResponse)
	if err != nil {
		return nil, err
	}

	return pduResponse.Data, nil
}
