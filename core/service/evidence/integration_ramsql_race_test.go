//go:build race

package evidence

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
)

// TestIntegration_Repro_RamsqlRace_StoreEvidence reproduces the in-memory DB
// race crash by running the already-failing evidence race test command and
// checking for the known runtime signature.
//
// This test is opt-in by design to keep CI stable while still providing a
// check-in reproducer for database debugging.
//
// Run with:
// CONFIRMATE_REPRO_RACE_DB=1 go test -race ./service/evidence -run TestIntegration_Repro_RamsqlRace_StoreEvidence -count=1 -v
func TestIntegration_Repro_RamsqlRace_StoreEvidence(t *testing.T) {
	var (
		cmd *exec.Cmd
		out []byte
		err error
	)

	if os.Getenv("CONFIRMATE_REPRO_RACE_DB") != "1" {
		t.Skip("set CONFIRMATE_REPRO_RACE_DB=1 to run this reproducer")
	}

	cmd = exec.Command("go", "test", "-race", ".", "-run", "TestService_StoreEvidence", "-count=5", "-timeout=300s")
	cmd.Env = os.Environ()
	out, err = cmd.CombinedOutput()

	if err == nil {
		t.Fatalf("expected nested race run to fail, but it succeeded\noutput:\n%s", string(out))
	}

	if !bytes.Contains(out, []byte("checkptr: pointer arithmetic result points to invalid allocation")) {
		t.Fatalf("expected checkptr crash signature in nested race run\noutput:\n%s", string(out))
	}
}