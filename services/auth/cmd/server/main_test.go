package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestServerRefusesToBootWithoutProdEnv is the regression guard for the
// missing EnsureProdReady call: the server binary must fail fast without a
// real JWT_SECRET / DATABASE_URL instead of booting on the public dev
// secret fallback. Runs the built binary (main cannot be tested in-process:
// log.Fatalf exits), so this needs the Go toolchain — always true in CI.
func TestServerRefusesToBootWithoutProdEnv(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "auth-test")

	buildCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	build := exec.CommandContext(buildCtx, "go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build server: %v\n%s", err, out)
	}

	runCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	run := exec.CommandContext(runCtx, bin)
	run.Env = scrubProdEnv(os.Environ())
	out, err := run.CombinedOutput()
	if err == nil {
		t.Fatalf("server booted without prod env, want refusal\n%s", out)
	}
	if !strings.Contains(string(out), "not production-ready") {
		t.Errorf("stderr missing refusal reason, got:\n%s", out)
	}
}

// scrubProdEnv removes the production settings so the binary must take the
// refusal path. Everything else (PATH, HOME, ...) is inherited.
func scrubProdEnv(env []string) []string {
	kept := env[:0]
	for _, kv := range env {
		if strings.HasPrefix(kv, "JWT_SECRET=") || strings.HasPrefix(kv, "DATABASE_URL=") {
			continue
		}
		kept = append(kept, kv)
	}
	return kept
}
