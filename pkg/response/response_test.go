package response

import "testing"

func TestSuccess(t *testing.T) {
	t.Parallel()

	resp := Success(map[string]string{"status": "ok"})
	if !resp.Success {
		t.Fatal("expected success response")
	}
	if resp.Data == nil {
		t.Fatal("expected data to be populated")
	}
}

func TestError(t *testing.T) {
	t.Parallel()

	resp := Error("boom")
	if resp.Success {
		t.Fatal("expected error response")
	}
	if resp.Message != "boom" {
		t.Fatalf("message mismatch: got %q", resp.Message)
	}
}
