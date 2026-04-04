package rime

import (
	"testing"

	"github.com/EasyIME/pime-go/pime"
)

func newTestIME() *IME {
	return New(&pime.Client{ID: "test-client"}).(*IME)
}

func TestNewInitialState(t *testing.T) {
	ime := newTestIME()

	if !ime.style.DisplayTrayIcon {
		t.Fatal("expected tray icon style enabled by default")
	}
	if ime.keyComposing {
		t.Fatal("expected keyComposing to be false initially")
	}
}

func TestFilterKeyDownDelegatesToOnKeyDown(t *testing.T) {
	ime := newTestIME()

	resp := ime.filterKeyDown(&pime.Request{
		SeqNum:   1,
		KeyCode:  0x4E,
		CharCode: 'n',
	}, pime.NewResponse(1, true))

	if resp.ReturnValue != 1 {
		t.Fatalf("expected n to be handled, got %d", resp.ReturnValue)
	}
	if resp.CompositionString != "ni" {
		t.Fatalf("expected composition ni, got %q", resp.CompositionString)
	}
	if len(resp.CandidateList) != 5 {
		t.Fatalf("expected 5 candidates, got %v", resp.CandidateList)
	}
}

func TestOnKeyDownIRequiresExistingComposition(t *testing.T) {
	ime := newTestIME()

	resp := ime.onKeyDown(&pime.Request{
		SeqNum:   2,
		KeyCode:  0x49,
		CharCode: 'i',
	}, pime.NewResponse(2, true))

	if resp.ReturnValue != 0 {
		t.Fatalf("expected bare i to be ignored, got %d", resp.ReturnValue)
	}
}

func TestOnKeyDownIWithNiCompositionShowsCandidates(t *testing.T) {
	ime := newTestIME()

	resp := ime.onKeyDown(&pime.Request{
		SeqNum:            3,
		KeyCode:           0x49,
		CharCode:          'i',
		CompositionString: "ni",
	}, pime.NewResponse(3, true))

	if resp.ReturnValue != 1 {
		t.Fatalf("expected i with ni composition to be handled, got %d", resp.ReturnValue)
	}
	if resp.CompositionString != "ni" {
		t.Fatalf("expected composition ni, got %q", resp.CompositionString)
	}
	if !resp.ShowCandidates {
		t.Fatal("expected candidates to be shown")
	}
	if resp.CandidateList[0] != "你" {
		t.Fatalf("expected first candidate 你, got %q", resp.CandidateList[0])
	}
}

func TestOnKeyDownNumberSelectsCandidate(t *testing.T) {
	ime := newTestIME()

	resp := ime.onKeyDown(&pime.Request{
		SeqNum:         4,
		KeyCode:        0x32,
		CandidateList:  []string{"你", "泥", "尼", "呢", "倪"},
		ShowCandidates: true,
	}, pime.NewResponse(4, true))

	if resp.ReturnValue != 1 {
		t.Fatalf("expected number selection to be handled, got %d", resp.ReturnValue)
	}
	if resp.CommitString != "泥" {
		t.Fatalf("expected second candidate 泥, got %q", resp.CommitString)
	}
}

func TestOnKeyDownUnhandledKeyReturnsZero(t *testing.T) {
	ime := newTestIME()

	resp := ime.onKeyDown(&pime.Request{
		SeqNum:   5,
		KeyCode:  0x41,
		CharCode: 'a',
	}, pime.NewResponse(5, true))

	if resp.ReturnValue != 0 {
		t.Fatalf("expected unrelated key to be ignored, got %d", resp.ReturnValue)
	}
}

func TestOnCommandHandlesKnownAndMissingCommand(t *testing.T) {
	ime := newTestIME()

	validResp := ime.onCommand(&pime.Request{
		SeqNum: 6,
		Data: map[string]interface{}{
			"commandId": float64(ID_ASCII_MODE),
		},
	}, pime.NewResponse(6, true))
	if validResp.ReturnValue != 1 {
		t.Fatalf("expected known command to be handled, got %d", validResp.ReturnValue)
	}

	missingResp := ime.onCommand(&pime.Request{
		SeqNum: 7,
	}, pime.NewResponse(7, true))
	if missingResp.ReturnValue != 0 {
		t.Fatalf("expected missing commandId to be ignored, got %d", missingResp.ReturnValue)
	}
}

func TestHandleRequestOnDeactivateReturnsHandled(t *testing.T) {
	ime := newTestIME()

	resp := ime.HandleRequest(&pime.Request{
		SeqNum: 8,
		Method: "onDeactivate",
	})

	if resp.ReturnValue != 1 {
		t.Fatalf("expected onDeactivate to return 1, got %d", resp.ReturnValue)
	}
}
