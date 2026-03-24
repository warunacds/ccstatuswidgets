package cli_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/cli"
)

func TestRunHN_PrintsTop5Stories(t *testing.T) {
	stories := []map[string]interface{}{
		{"id": 1, "title": "Story One", "score": 100, "url": "https://example.com/1"},
		{"id": 2, "title": "Story Two", "score": 200, "url": "https://example.com/2"},
		{"id": 3, "title": "Story Three", "score": 300, "url": "https://example.com/3"},
		{"id": 4, "title": "Story Four", "score": 400, "url": "https://example.com/4"},
		{"id": 5, "title": "Story Five", "score": 500, "url": "https://example.com/5"},
		{"id": 6, "title": "Story Six", "score": 600, "url": "https://example.com/6"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v0/topstories.json" {
			ids := []int{1, 2, 3, 4, 5, 6}
			json.NewEncoder(w).Encode(ids)
			return
		}
		for _, s := range stories {
			id := int(s["id"].(int))
			expected := fmt.Sprintf("/v0/item/%d.json", id)
			if r.URL.Path == expected {
				json.NewEncoder(w).Encode(s)
				return
			}
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	// Capture stdout.
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cli.RunHN(srv.URL)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("RunHN returned error: %v", err)
	}

	out, _ := io.ReadAll(r)
	output := string(out)

	// Verify all 5 stories are present.
	for i := 0; i < 5; i++ {
		title := stories[i]["title"].(string)
		if !strings.Contains(output, title) {
			t.Errorf("expected output to contain %q", title)
		}
		url := stories[i]["url"].(string)
		if !strings.Contains(output, url) {
			t.Errorf("expected output to contain URL %q", url)
		}
	}

	// Story Six should NOT be in output (only top 5).
	if strings.Contains(output, "Story Six") {
		t.Error("expected output to NOT contain Story Six (only top 5)")
	}

	// Verify numbered format.
	if !strings.Contains(output, "1.") {
		t.Error("expected output to contain numbered list starting with '1.'")
	}
	if !strings.Contains(output, "5.") {
		t.Error("expected output to contain '5.' for the fifth story")
	}

	// Verify score is shown.
	if !strings.Contains(output, "100 pts") {
		t.Error("expected output to contain score '100 pts'")
	}
}
