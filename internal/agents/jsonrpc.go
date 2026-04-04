package agents

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Constantes JSON-RPC
const JSONRPCVersion = "2.0"

// JSONRPCMessage representa a estrutura base de qualquer mensagem ACP (JSON-RPC).
type JSONRPCMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`     // Pode ser int ou string
	Method  string          `json:"method,omitempty"`  // Requisições/Notificações
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`  // Respostas
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError define um erro do protocolo JSON-RPC
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// JSONRPCHandler define como o Lumaestro deve reagir às mensagens do Gemini CLI.
type JSONRPCHandler interface {
	HandleNotification(method string, params json.RawMessage)
	HandleRequest(id interface{}, method string, params json.RawMessage)
	HandleResponse(id interface{}, result json.RawMessage, err *RPCError)
}

// StartJSONRPCListener lê mensagens no formato ACP oficial (ndJSON - Newline-Delimited JSON).
// Cada mensagem JSON-RPC é uma única linha terminada com '\n'.
func StartJSONRPCListener(r io.Reader, handler JSONRPCHandler) {
	scanner := bufio.NewScanner(r)
	
	// Aumenta o buffer do scanner para lidar com mensagens grandes (ex: listas de arquivos)
	const maxCapacity = 1 * 1024 * 1024 // 1MB
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Bytes()
		
		// 🛠️ LOG DE DEPURAÇÃO
		fmt.Printf("[STDOUT RAW] %s\n", string(line))
		
		// Remove espaços em branco
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}

		// Decodifica a mensagem JSON-RPC
		var msg JSONRPCMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			// Pode ser um log ou ruído no stdout, ignoramos se não for JSON válido
			continue
		}

		// Roteamento Inteligente (Idêntico ao anterior)
		if msg.Method != "" {
			if msg.ID != nil {
				handler.HandleRequest(msg.ID, msg.Method, msg.Params)
			} else {
				handler.HandleNotification(msg.Method, msg.Params)
			}
		} else {
			handler.HandleResponse(msg.ID, msg.Result, msg.Error)
		}
	}
}
