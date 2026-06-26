package sdk

import (
	"crypto/md5"
	"encoding/hex"
	"testing"
)

// TestSkipActualFileHash verifies the GOSDK_SKIP_ACTUAL_FILE_HASH gate:
//   - OFF: WriteToFile feeds the whole-file MD5, GetFileHash returns it.
//   - ON:  WriteToFile is a no-op and GetFileHash returns the fixed placeholder,
//     while the per-block DataHash (the only hash the blobber recomputes +
//     enforces at commit) is UNCHANGED — so skipping ActualHash can never make
//     a commit fail with hash_mismatch.
func TestSkipActualFileHash(t *testing.T) {
	data := []byte("the quick brown fox jumps over the lazy dog")
	sum := md5.Sum(data)
	wantHex := hex.EncodeToString(sum[:])
	const placeholder = "00000000000000000000000000000000"

	saved := skipActualFileHash
	defer func() { skipActualFileHash = saved }()

	// --- skip OFF: real content MD5 ---
	skipActualFileHash = false
	h := CreateHasher(int64(len(data))).(*hasher)
	if err := h.WriteToFile(data); err != nil {
		t.Fatalf("WriteToFile (skip off): %v", err)
	}
	got, err := h.GetFileHash()
	if err != nil {
		t.Fatalf("GetFileHash (skip off): %v", err)
	}
	if got != wantHex {
		t.Fatalf("skip off: file hash = %s, want %s", got, wantHex)
	}

	// --- skip ON: no-op write + placeholder hash ---
	skipActualFileHash = true
	h2 := CreateHasher(int64(len(data))).(*hasher)
	if err := h2.WriteToFile(data); err != nil {
		t.Fatalf("WriteToFile (skip on): %v", err)
	}
	got2, err := h2.GetFileHash()
	if err != nil {
		t.Fatalf("GetFileHash (skip on): %v", err)
	}
	if got2 != placeholder {
		t.Fatalf("skip on: file hash = %s, want placeholder %s", got2, placeholder)
	}
	if got2 == wantHex {
		t.Fatal("skip on: file hash must NOT equal the real content MD5")
	}

	// CRITICAL SAFETY: the per-block DataHash the blobber enforces is untouched.
	if err := h2.WriteToBlockHasher(data); err != nil {
		t.Fatalf("WriteToBlockHasher (skip on): %v", err)
	}
	blk, err := h2.GetBlockHash()
	if err != nil {
		t.Fatalf("GetBlockHash (skip on): %v", err)
	}
	if blk != wantHex {
		t.Fatalf("skip on: DataHash = %s, want real MD5 %s — blobber would reject", blk, wantHex)
	}
}
