package main

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	meowime "github.com/EasyIME/pime-go/input_methods/meow"
	"github.com/EasyIME/pime-go/pime"
)

const testMeowGUID = "{7A1C2E93-5B64-4F88-AE21-3D9C6B70F145}"

func newTestServerWithMeow() *Server {
	server := NewServer()
	server.RegisterService(testMeowGUID, func(clientID, guid string) pime.TextService {
		return meowime.New(&pime.Client{ID: clientID})
	})
	return server
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}

	os.Stdout = writer
	defer func() {
		os.Stdout = oldStdout
	}()

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read captured stdout: %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("close stdout reader: %v", err)
	}

	return strings.TrimSpace(string(output))
}

func sendProtocolMessage(t *testing.T, server *Server, clientID string, payload map[string]interface{}) (string, map[string]interface{}) {
	t.Helper()

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	line := clientID + "|" + string(data)
	output := captureStdout(t, func() {
		if err := server.handleMessage(line); err != nil {
			t.Fatalf("handleMessage failed: %v", err)
		}
	})

	prefix := pime.MsgPIME + "|" + clientID + "|"
	if !strings.HasPrefix(output, prefix) {
		t.Fatalf("expected %q prefix, got %q", prefix, output)
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimPrefix(output, prefix)), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	return output, response
}

func TestServerHandleMessageInitUsesTopLevelID(t *testing.T) {
	server := newTestServerWithMeow()

	_, response := sendProtocolMessage(t, server, "client-1", map[string]interface{}{
		"method":          "init",
		"seqNum":          1,
		"id":              testMeowGUID,
		"isWindows8Above": true,
		"isMetroApp":      false,
		"isUiLess":        false,
		"isConsole":       false,
	})

	if response["success"] != true {
		t.Fatalf("expected init success, got %#v", response)
	}
	if response["seqNum"] != float64(1) {
		t.Fatalf("expected seqNum 1, got %#v", response["seqNum"])
	}

	client := server.clients["client-1"]
	if client == nil {
		t.Fatal("expected client to be registered after init")
	}
	if client.GUID != testMeowGUID {
		t.Fatalf("expected guid %q, got %q", testMeowGUID, client.GUID)
	}
}

func TestServerHandleMessageMeowRequestResponseFlow(t *testing.T) {
	server := newTestServerWithMeow()

	sendProtocolMessage(t, server, "client-2", map[string]interface{}{
		"method":          "init",
		"seqNum":          1,
		"id":              testMeowGUID,
		"isWindows8Above": true,
		"isMetroApp":      false,
		"isUiLess":        false,
		"isConsole":       false,
	})

	_, filterResp := sendProtocolMessage(t, server, "client-2", map[string]interface{}{
		"method":   "filterKeyDown",
		"seqNum":   2,
		"keyCode":  0x4D,
		"charCode": 'm',
	})
	if filterResp["return"] != float64(1) {
		t.Fatalf("expected filterKeyDown to handle m, got %#v", filterResp)
	}

	_, firstKeyResp := sendProtocolMessage(t, server, "client-2", map[string]interface{}{
		"method":   "onKeyDown",
		"seqNum":   3,
		"keyCode":  0x4D,
		"charCode": 'm',
	})
	if firstKeyResp["compositionString"] != "喵" {
		t.Fatalf("expected first m to build composition 喵, got %#v", firstKeyResp)
	}
	if firstKeyResp["return"] != float64(1) {
		t.Fatalf("expected first m return 1, got %#v", firstKeyResp)
	}

	_, secondKeyResp := sendProtocolMessage(t, server, "client-2", map[string]interface{}{
		"method":   "onKeyDown",
		"seqNum":   4,
		"keyCode":  0x4D,
		"charCode": 'm',
	})
	if secondKeyResp["showCandidates"] != true {
		t.Fatalf("expected second m to show candidates, got %#v", secondKeyResp)
	}
	candidateList, ok := secondKeyResp["candidateList"].([]interface{})
	if !ok {
		t.Fatalf("expected candidate list array, got %#v", secondKeyResp["candidateList"])
	}
	if len(candidateList) != 4 {
		t.Fatalf("expected 4 candidates, got %d", len(candidateList))
	}
	if candidateList[1] != "描" {
		t.Fatalf("expected second candidate 描, got %#v", candidateList[1])
	}

	_, selectResp := sendProtocolMessage(t, server, "client-2", map[string]interface{}{
		"method":  "onKeyDown",
		"seqNum":  5,
		"keyCode": 0x32,
	})
	if selectResp["commitString"] != "描" {
		t.Fatalf("expected number key to commit 描, got %#v", selectResp)
	}
	if selectResp["showCandidates"] != false {
		t.Fatalf("expected candidate window to close, got %#v", selectResp)
	}
	if selectResp["return"] != float64(1) {
		t.Fatalf("expected candidate selection return 1, got %#v", selectResp)
	}
}

func TestServerHandleMessageUninitializedClientReturnsProtocolError(t *testing.T) {
	server := newTestServerWithMeow()

	_, response := sendProtocolMessage(t, server, "client-3", map[string]interface{}{
		"method":   "onKeyDown",
		"seqNum":   9,
		"keyCode":  0x4D,
		"charCode": 'm',
	})

	if response["success"] != false {
		t.Fatalf("expected uninitialized client to fail, got %#v", response)
	}
	if response["seqNum"] != float64(9) {
		t.Fatalf("expected seqNum 9, got %#v", response["seqNum"])
	}
	if response["error"] != "客户端未初始化" {
		t.Fatalf("expected protocol error for uninitialized client, got %#v", response["error"])
	}
}
