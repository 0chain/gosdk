//go:build integration

// Package s3interop verifies that AWS CLI, mc (MinIO client), and rclone all
// speak the same language when pointed at the same Züs wallet/allocation via
// the zs3server S3 gateway.  Every PUT/GET/DELETE/LIST/overwrite combination
// across all three clients is exercised.
//
// Prerequisites: aws, mc, and rclone must be on PATH (or override via
// AWS_BIN, MC_BIN, RCLONE_BIN env vars).
//
// Required env vars:
//
//	S3_ENDPOINT   e.g. http://144.76.58.147:8711
//	S3_ACCESS_KEY wallet client_id / configured S3 credential
//	S3_SECRET_KEY wallet private_key / configured S3 credential
//	S3_BUCKET     pre-existing bucket name in the allocation
//
// Run:
//
//	go test -v -tags integration -count=1 -timeout 5m \
//	    ./system_tests/s3interop/
package s3interop_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ------------------------------------------------------------------ env / setup

type testEnv struct {
	endpoint  string
	accessKey string
	secretKey string
	bucket    string
	awsBin    string
	mcBin     string
	rcloneBin string
	mcAlias   string
	mcCfgDir  string
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()
	endpoint := os.Getenv("S3_ENDPOINT")
	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")
	bucket := os.Getenv("S3_BUCKET")
	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" {
		t.Skip("set S3_ENDPOINT, S3_ACCESS_KEY, S3_SECRET_KEY, S3_BUCKET to run interop tests")
	}

	bin := func(envKey, dflt string) string {
		if v := os.Getenv(envKey); v != "" {
			return v
		}
		return dflt
	}

	e := &testEnv{
		endpoint:  endpoint,
		accessKey: accessKey,
		secretKey: secretKey,
		bucket:    bucket,
		awsBin:    bin("AWS_BIN", "aws"),
		mcBin:     bin("MC_BIN", "mc"),
		rcloneBin: bin("RCLONE_BIN", "rclone"),
	}

	// Configure a temporary mc alias so tests don't touch the user's mc config.
	e.mcCfgDir = t.TempDir()
	e.mcAlias = fmt.Sprintf("zus%d", time.Now().UnixNano()%100000)
	out, err := exec.Command(e.mcBin,
		"--config-dir", e.mcCfgDir,
		"alias", "set", e.mcAlias,
		endpoint, accessKey, secretKey,
	).CombinedOutput()
	require.NoError(t, err, "mc alias set: %s", out)

	return e
}

func uniqueKey(prefix string) string {
	return fmt.Sprintf("interop/%s/%d", prefix, time.Now().UnixNano())
}

func writeLocalFile(t *testing.T, data []byte) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), "payload")
	require.NoError(t, os.WriteFile(f, data, 0o600))
	return f
}

// ------------------------------------------------------------------ aws CLI helpers

func (e *testEnv) awsEnv() []string {
	return append(os.Environ(),
		"AWS_ACCESS_KEY_ID="+e.accessKey,
		"AWS_SECRET_ACCESS_KEY="+e.secretKey,
		"AWS_DEFAULT_REGION=us-east-1",
	)
}

func (e *testEnv) awsRun(t *testing.T, args ...string) []byte {
	t.Helper()
	cmd := exec.Command(e.awsBin, args...)
	cmd.Env = e.awsEnv()
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "aws %v: %s", args, out)
	return out
}

func (e *testEnv) awsPut(t *testing.T, key string, data []byte) {
	t.Helper()
	e.awsRun(t, "s3", "cp", writeLocalFile(t, data),
		fmt.Sprintf("s3://%s/%s", e.bucket, key),
		"--endpoint-url", e.endpoint, "--no-progress")
}

func (e *testEnv) awsGet(t *testing.T, key string) []byte {
	t.Helper()
	dst := filepath.Join(t.TempDir(), "aws-get")
	e.awsRun(t, "s3", "cp",
		fmt.Sprintf("s3://%s/%s", e.bucket, key), dst,
		"--endpoint-url", e.endpoint, "--no-progress")
	data, err := os.ReadFile(dst)
	require.NoError(t, err)
	return data
}

func (e *testEnv) awsDelete(t *testing.T, key string) {
	t.Helper()
	cmd := exec.Command(e.awsBin, "s3", "rm",
		fmt.Sprintf("s3://%s/%s", e.bucket, key),
		"--endpoint-url", e.endpoint)
	cmd.Env = e.awsEnv()
	// ignore errors — cleanup best-effort
	cmd.Run() //nolint:errcheck
}

func (e *testEnv) awsDeleteForce(t *testing.T, key string) {
	t.Helper()
	e.awsRun(t, "s3", "rm",
		fmt.Sprintf("s3://%s/%s", e.bucket, key),
		"--endpoint-url", e.endpoint)
}

func (e *testEnv) awsList(t *testing.T, prefix string) []string {
	t.Helper()
	out := e.awsRun(t, "s3", "ls",
		fmt.Sprintf("s3://%s/%s", e.bucket, prefix),
		"--endpoint-url", e.endpoint, "--recursive")
	var keys []string
	for _, line := range strings.Split(string(out), "\n") {
		// aws s3 ls --recursive lines: "2024-01-01 00:00:00      123 key/name"
		parts := strings.Fields(line)
		if len(parts) >= 4 {
			keys = append(keys, parts[3])
		}
	}
	return keys
}

func (e *testEnv) awsExists(t *testing.T, key string) bool {
	t.Helper()
	cmd := exec.Command(e.awsBin, "s3", "ls",
		fmt.Sprintf("s3://%s/%s", e.bucket, key),
		"--endpoint-url", e.endpoint)
	cmd.Env = e.awsEnv()
	return cmd.Run() == nil
}

// ------------------------------------------------------------------ mc CLI helpers

func (e *testEnv) mcRun(t *testing.T, args ...string) []byte {
	t.Helper()
	base := []string{"--config-dir", e.mcCfgDir}
	cmd := exec.Command(e.mcBin, append(base, args...)...)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "mc %v: %s", args, out)
	return out
}

func (e *testEnv) mcRemote(key string) string {
	return fmt.Sprintf("%s/%s/%s", e.mcAlias, e.bucket, key)
}

func (e *testEnv) mcPut(t *testing.T, key string, data []byte) {
	t.Helper()
	e.mcRun(t, "cp", writeLocalFile(t, data), e.mcRemote(key))
}

func (e *testEnv) mcGet(t *testing.T, key string) []byte {
	t.Helper()
	dst := filepath.Join(t.TempDir(), "mc-get")
	e.mcRun(t, "cp", e.mcRemote(key), dst)
	data, err := os.ReadFile(dst)
	require.NoError(t, err)
	return data
}

func (e *testEnv) mcDelete(t *testing.T, key string) {
	t.Helper()
	e.mcRun(t, "rm", e.mcRemote(key))
}

// mcList returns keys (full paths from bucket root) under the given prefix.
func (e *testEnv) mcList(t *testing.T, prefix string) []string {
	t.Helper()
	out := e.mcRun(t, "ls", "--json", "--recursive", e.mcRemote(prefix))
	var keys []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var obj struct {
			Key string `json:"key"`
		}
		if err := json.Unmarshal([]byte(line), &obj); err == nil && obj.Key != "" {
			keys = append(keys, obj.Key)
		}
	}
	return keys
}

func (e *testEnv) mcExists(t *testing.T, key string) bool {
	t.Helper()
	cmd := exec.Command(e.mcBin, "--config-dir", e.mcCfgDir,
		"stat", e.mcRemote(key))
	return cmd.Run() == nil
}

// ------------------------------------------------------------------ rclone CLI helpers

func (e *testEnv) rcloneEnv() []string {
	return append(os.Environ(),
		"RCLONE_S3_PROVIDER=Other",
		"RCLONE_S3_ENDPOINT="+e.endpoint,
		"RCLONE_S3_ACCESS_KEY_ID="+e.accessKey,
		"RCLONE_S3_SECRET_ACCESS_KEY="+e.secretKey,
		"RCLONE_S3_PATH_STYLE=true",
	)
}

func (e *testEnv) rcloneRemote(key string) string {
	return fmt.Sprintf(":s3:%s/%s", e.bucket, key)
}

func (e *testEnv) rcloneRun(t *testing.T, args ...string) []byte {
	t.Helper()
	cmd := exec.Command(e.rcloneBin, args...)
	cmd.Env = e.rcloneEnv()
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "rclone %v: %s", args, out)
	return out
}

func (e *testEnv) rclonePut(t *testing.T, key string, data []byte) {
	t.Helper()
	e.rcloneRun(t, "copyto", writeLocalFile(t, data), e.rcloneRemote(key))
}

func (e *testEnv) rcloneGet(t *testing.T, key string) []byte {
	t.Helper()
	dst := filepath.Join(t.TempDir(), "rclone-get")
	e.rcloneRun(t, "copyto", e.rcloneRemote(key), dst)
	data, err := os.ReadFile(dst)
	require.NoError(t, err)
	return data
}

func (e *testEnv) rcloneDelete(t *testing.T, key string) {
	t.Helper()
	e.rcloneRun(t, "deletefile", e.rcloneRemote(key))
}

// rcloneList returns keys under prefix using rclone lsf.
func (e *testEnv) rcloneList(t *testing.T, prefix string) []string {
	t.Helper()
	cmd := exec.Command(e.rcloneBin, "lsf", "--recursive", e.rcloneRemote(prefix))
	cmd.Env = e.rcloneEnv()
	out, err := cmd.CombinedOutput()
	// rclone exits non-zero on empty directory; that's fine.
	if err != nil && len(out) == 0 {
		return nil
	}
	var keys []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			// lsf returns paths relative to the prefix; prepend it.
			keys = append(keys, prefix+line)
		}
	}
	return keys
}

func (e *testEnv) rcloneExists(t *testing.T, key string) bool {
	t.Helper()
	cmd := exec.Command(e.rcloneBin, "lsf", e.rcloneRemote(key))
	cmd.Env = e.rcloneEnv()
	out, _ := cmd.CombinedOutput()
	return len(strings.TrimSpace(string(out))) > 0
}

// ================================================================== Tests

// TestInteropCrossClientVisibility — write via one client, read via the
// other two. Covers all six writer→reader combinations across aws/mc/rclone.
func TestInteropCrossClientVisibility(t *testing.T) {
	e := setupTestEnv(t)

	payload := []byte("cross-client interop payload — Züs universal S3 interop")

	type clientOps struct {
		name string
		put  func(key string, data []byte)
		get  func(key string) []byte
	}

	clients := []clientOps{
		{
			name: "aws",
			put:  func(k string, d []byte) { e.awsPut(t, k, d) },
			get:  func(k string) []byte { return e.awsGet(t, k) },
		},
		{
			name: "mc",
			put:  func(k string, d []byte) { e.mcPut(t, k, d) },
			get:  func(k string) []byte { return e.mcGet(t, k) },
		},
		{
			name: "rclone",
			put:  func(k string, d []byte) { e.rclonePut(t, k, d) },
			get:  func(k string) []byte { return e.rcloneGet(t, k) },
		},
	}

	for _, writer := range clients {
		writer := writer
		for _, reader1 := range clients {
			for _, reader2 := range clients {
				if reader1.name == writer.name || reader2.name == writer.name || reader1.name == reader2.name {
					continue
				}
				reader1, reader2 := reader1, reader2
				name := fmt.Sprintf("%s_write__%s_%s_read", writer.name, reader1.name, reader2.name)
				t.Run(name, func(t *testing.T) {
					key := uniqueKey(name)
					t.Cleanup(func() { e.awsDelete(t, key) })

					writer.put(key, payload)
					require.Equal(t, payload, reader1.get(key),
						"%s wrote, %s should read identical bytes", writer.name, reader1.name)
					require.Equal(t, payload, reader2.get(key),
						"%s wrote, %s should read identical bytes", writer.name, reader2.name)
				})
			}
		}
	}
}

// TestInteropListAgreement — all three clients put one object each under a
// shared prefix, then list; all three listings must agree on the full set.
func TestInteropListAgreement(t *testing.T) {
	e := setupTestEnv(t)

	prefix := fmt.Sprintf("interop/list/%d/", time.Now().UnixNano())
	objects := map[string][]byte{
		prefix + "aws-object":    []byte("written-by-aws"),
		prefix + "mc-object":     []byte("written-by-mc"),
		prefix + "rclone-object": []byte("written-by-rclone"),
	}

	t.Cleanup(func() {
		for key := range objects {
			e.awsDelete(t, key)
		}
	})

	e.awsPut(t, prefix+"aws-object", objects[prefix+"aws-object"])
	e.mcPut(t, prefix+"mc-object", objects[prefix+"mc-object"])
	e.rclonePut(t, prefix+"rclone-object", objects[prefix+"rclone-object"])

	want := []string{
		prefix + "aws-object",
		prefix + "mc-object",
		prefix + "rclone-object",
	}
	sort.Strings(want)

	awsKeys := e.awsList(t, prefix)
	sort.Strings(awsKeys)
	require.Equal(t, want, awsKeys, "AWS CLI list disagrees with expected set")

	mcKeys := e.mcList(t, prefix)
	sort.Strings(mcKeys)
	require.Equal(t, want, mcKeys, "mc list disagrees with expected set")

	rcloneKeys := e.rcloneList(t, prefix)
	sort.Strings(rcloneKeys)
	require.Equal(t, want, rcloneKeys, "rclone list disagrees with expected set")
}

// TestInteropDeletePropagation — object deleted by one client must be
// immediately absent when queried by the other two clients.
func TestInteropDeletePropagation(t *testing.T) {
	e := setupTestEnv(t)

	cases := []struct {
		deleterName string
		deleteFn    func(key string)
	}{
		{"aws", func(k string) { e.awsDeleteForce(t, k) }},
		{"mc", func(k string) { e.mcDelete(t, k) }},
		{"rclone", func(k string) { e.rcloneDelete(t, k) }},
	}

	for _, tc := range cases {
		tc := tc
		t.Run("deleted_by_"+tc.deleterName, func(t *testing.T) {
			key := uniqueKey("delete-" + tc.deleterName)
			e.awsPut(t, key, []byte("to be deleted"))

			require.True(t, e.awsExists(t, key), "object must exist before delete")

			tc.deleteFn(key)

			require.False(t, e.awsExists(t, key),
				"aws: object still visible after %s deleted it", tc.deleterName)
			require.False(t, e.mcExists(t, key),
				"mc: object still visible after %s deleted it", tc.deleterName)
			require.False(t, e.rcloneExists(t, key),
				"rclone: object still visible after %s deleted it", tc.deleterName)
		})
	}
}

// TestInteropOverwrite — successive overwrites by different clients must be
// reflected with the correct latest content by all other clients.
func TestInteropOverwrite(t *testing.T) {
	e := setupTestEnv(t)

	key := uniqueKey("overwrite")
	t.Cleanup(func() { e.awsDelete(t, key) })

	type version struct {
		writer  string
		content []byte
		write   func()
	}
	versions := []version{
		{"aws", []byte("v1-written-by-aws"), func() { e.awsPut(t, key, []byte("v1-written-by-aws")) }},
		{"mc", []byte("v2-overwritten-by-mc"), func() { e.mcPut(t, key, []byte("v2-overwritten-by-mc")) }},
		{"rclone", []byte("v3-overwritten-by-rclone"), func() { e.rclonePut(t, key, []byte("v3-overwritten-by-rclone")) }},
	}

	for _, v := range versions {
		v := v
		t.Logf("overwriting with %s", v.writer)
		v.write()

		require.Equal(t, v.content, e.awsGet(t, key),
			"aws should see content written by %s", v.writer)
		require.Equal(t, v.content, e.mcGet(t, key),
			"mc should see content written by %s", v.writer)
		require.Equal(t, v.content, e.rcloneGet(t, key),
			"rclone should see content written by %s", v.writer)
	}
}

// TestInteropLargeObjectRoundtrip — 8 MB object round-trip across all three
// clients to exercise multi-chunk uploads and downloads.
func TestInteropLargeObjectRoundtrip(t *testing.T) {
	e := setupTestEnv(t)

	const size = 8 << 20 // 8 MB
	payload := make([]byte, size)
	for i := range payload {
		payload[i] = byte(i % 251)
	}

	writers := []struct {
		name  string
		write func(key string)
	}{
		{"aws", func(k string) { e.awsPut(t, k, payload) }},
		{"mc", func(k string) { e.mcPut(t, k, payload) }},
		{"rclone", func(k string) { e.rclonePut(t, k, payload) }},
	}

	for _, w := range writers {
		w := w
		t.Run("written_by_"+w.name, func(t *testing.T) {
			key := uniqueKey("large-" + w.name)
			t.Cleanup(func() { e.awsDelete(t, key) })

			w.write(key)

			require.Equal(t, payload, e.awsGet(t, key),
				"aws: large object content mismatch after %s write", w.name)
			require.Equal(t, payload, e.mcGet(t, key),
				"mc: large object content mismatch after %s write", w.name)
			require.Equal(t, payload, e.rcloneGet(t, key),
				"rclone: large object content mismatch after %s write", w.name)
		})
	}
}
