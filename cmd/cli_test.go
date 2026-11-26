package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runCLI(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()
	root := newRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader(stdin))
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}

func newTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	t.Setenv(envDeveloperCode, "test-code")
	t.Setenv(envBaseURL, server.URL)
	return server
}

func TestHelpListsGroups(t *testing.T) {
	out, err := runCLI(t, "", "--help")
	if err != nil {
		t.Fatalf("help: %v", err)
	}
	for _, want := range []string{"who", "templates", "locks"} {
		if !strings.Contains(out, want) {
			t.Errorf("help does not mention %q:\n%s", want, out)
		}
	}
}

func TestVersion(t *testing.T) {
	out, err := runCLI(t, "", "--version")
	if err != nil {
		t.Fatalf("version: %v", err)
	}
	if !strings.Contains(out, "chastisafe version") {
		t.Fatalf("version output = %q", out)
	}
}

func TestYESAndDryRunConflict(t *testing.T) {
	_, err := runCLI(t, "", "locks", "freeze", "lock-id", "--yes", "--dry-run")
	if err == nil || !strings.Contains(err.Error(), "cannot be used together") {
		t.Fatalf("error = %v", err)
	}
}

func TestDryRunAbortMakesNoRequest(t *testing.T) {
	hits := 0
	newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		hits++
		_, _ = w.Write([]byte(`{"id":"lock-id"}`))
	})

	out, err := runCLI(t, "n\n", "locks", "freeze", "lock-id")
	if err != nil {
		t.Fatalf("freeze: %v", err)
	}
	if hits != 0 {
		t.Fatalf("server was called %d times", hits)
	}
	if !strings.Contains(out, "dry run: locks freeze lock-id") {
		t.Fatalf("output = %q", out)
	}
	if !strings.Contains(out, "aborted") {
		t.Fatalf("output = %q", out)
	}
}

func TestDryRunFlagSkipsPrompt(t *testing.T) {
	hits := 0
	newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		hits++
		_, _ = w.Write([]byte(`{"id":"lock-id"}`))
	})

	out, err := runCLI(t, "", "locks", "freeze", "lock-id", "--dry-run")
	if err != nil {
		t.Fatalf("freeze: %v", err)
	}
	if hits != 0 {
		t.Fatalf("server was called %d times", hits)
	}
	if strings.Contains(out, "Apply?") {
		t.Fatalf("dry run prompted:\n%s", out)
	}
}

func TestYesAppliesFreeze(t *testing.T) {
	var body map[string]any
	newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/update" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		_, _ = w.Write([]byte(`{"id":"lock-id","lockname":"demo","type":"card"}`))
	})

	out, err := runCLI(t, "", "locks", "freeze", "lock-id", "--yes")
	if err != nil {
		t.Fatalf("freeze: %v", err)
	}
	if body["id"] != "lock-id" || body["frozenbykh"] != true {
		t.Fatalf("body = %+v", body)
	}
	if !strings.Contains(out, `"lockname": "demo"`) {
		t.Fatalf("output = %q", out)
	}
}

func TestPromptYesApplies(t *testing.T) {
	var body map[string]any
	newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		_, _ = w.Write([]byte(`{"id":"lock-id"}`))
	})

	_, err := runCLI(t, "yes\n", "locks", "freeze", "lock-id")
	if err != nil {
		t.Fatalf("freeze: %v", err)
	}
	if body["frozenbykh"] != true {
		t.Fatalf("body = %+v", body)
	}
}

func TestTemplatesUpdateInjectsIDFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "template.json")
	if err := os.WriteFile(path, []byte(`{"description":"from file","id":"wrong"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	var body map[string]any
	newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/updatetemplate" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		_, _ = w.Write([]byte(`{"status":0}`))
	})

	_, err := runCLI(t, "", "templates", "update", "template-id", "@"+path, "--yes")
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if body["templateId"] != "template-id" {
		t.Fatalf("templateId = %v", body["templateId"])
	}
	if _, ok := body["id"]; ok {
		t.Fatalf("body still has id: %+v", body)
	}
	if body["description"] != "from file" {
		t.Fatalf("description = %v", body["description"])
	}
}

func TestCardsBuildsChanges(t *testing.T) {
	var body map[string]any
	newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/update" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		_, _ = w.Write([]byte(`{"id":"lock-id"}`))
	})

	_, err := runCLI(t, "",
		"locks", "cards", "lock-id",
		"--add", "green:2,red:1",
		"--remove", "yellow:1",
		"--message", "rebalance",
		"--show",
		"--yes",
	)
	if err != nil {
		t.Fatalf("cards: %v", err)
	}

	rawCards, ok := body["cards"].([]any)
	if !ok {
		t.Fatalf("cards = %T %v", body["cards"], body["cards"])
	}
	want := []struct {
		card  string
		delta float64
	}{
		{"GREEN", 2},
		{"RED", 1},
		{"YELLOW", -1},
	}
	if len(rawCards) != len(want) {
		t.Fatalf("cards = %+v", rawCards)
	}
	for i, expected := range want {
		card, ok := rawCards[i].(map[string]any)
		if !ok {
			t.Fatalf("card %d = %T", i, rawCards[i])
		}
		if card["card"] != expected.card || card["delta"] != expected.delta {
			t.Errorf("card %d = %+v, want %s %v", i, card, expected.card, expected.delta)
		}
	}
	if body["message"] != "rebalance" || body["showupdate"] != true {
		t.Errorf("body = %+v", body)
	}
}

func TestRequiredDurationFlag(t *testing.T) {
	_, err := runCLI(t, "", "locks", "add-time", "lock-id")
	if err == nil || !strings.Contains(err.Error(), `required flag(s) "duration" not set`) {
		t.Fatalf("error = %v", err)
	}
}
