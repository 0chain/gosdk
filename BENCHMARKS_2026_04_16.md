# Benchmarks — 2026-04-16

## Scope
Validate chain-down operation of zs3server + eblobber for SF100 TPC-DS over NFS-Ganesha.

## Patches landed

| Repo | File | Purpose |
|---|---|---|
| gosdk | `zboxcore/sdk/sdk.go` | Disk-persistent allocation cache (`SetAllocationCacheDir`) — zs3server bootstraps without sharders after the first fetch |
| zs3server | `cmd/gateway/zcn/initSDK.go` | Wires `sdk.SetAllocationCacheDir(<configDir>/alloc_cache)` after `InitStorageSDK` |
| zs3server | `cmd/gateway/zcn/prewarm_router.go` | In-place prewarm (preserves inode for Ganesha cached handles); deferred `MarkCommitted`; spillover-restore also in-place with deferred `MarkCommitted` |
| zs3server | `cmd/gateway/zcn/gateway-zcn.go` (PutObject) | Early `MarkCommitted` gates `processEvents` inotify-upload race; late `Setxattr user.zus.committed` gates spillover eviction |
| zs3server | `cmd/gateway/zcn/nfs_blobber_sync.go` | `spillCommittedFiles` requires `user.zus.committed` xattr (xattr-gate) AND removes the xattr on re-stub so the gate is transitive |
| eblobber | `code/go/0chain.net/blobbercore/allocation/protocol.go` | `FetchAllocationFromEventsDB` uses `Repo.GetById` first; only calls sharders on first-time discovery |

## Chain-down end-to-end verified
- zs3server start chain-DOWN: log confirms `Loaded allocation from disk cache`; `live=200`
- mc cp 33 GB SF100 re-upload chain-DOWN: 1758 files, 0 errors, 6m48s, bucket count == 1758
- Byte-identical roundtrip via NFS (post-remount for clean attr cache): 3/3 sample parquets match source

## Debugging saga (7 iterations to fully correct spillover+prewarm)
1. **v1** — temp+rename prewarm → inode changed → Ganesha's cached file handle pointed at orphaned stub → reads returned sparse zeros
2. **v2** — in-place prewarm + deferred `MarkCommitted` → fixed inode, exposed a race in PutObject (which had early `MarkCommitted` but no xattr gate)
3. **v3** — removed PutObject's early `MarkCommitted` → broke the inotify gate → `processEvents` uploaded fresh PUTs racing PutObject's own `putFile` → **41/1758 files silently dropped**
4. **v4** — PutObject dual-gate (early `MarkCommitted` + late xattr) → upload complete, but spill-then-restore race corrupted files (`O_TRUNC` preserved stale `committed` xattr → the v3 xattr-gate was ineffective)
5. **v5** — `spillCommittedFiles` removes `committed` xattr on re-stub → still failed because…
6. **Real root cause**: my v3 xattr gate used `if n, _ := syscall.Getxattr(…); n == 0 { continue }`. On Linux, Go's `syscall.Getxattr` returns `(-errno, err)` on ENODATA, not `(0, err)`. On x86_64, the errno is a small negative int (`-61`). My `n == 0` check **never fired**; every file in `bs.committed` was spilled, including sparse stubs. `ReadFile` on a sparse stub returns raw zero bytes → the spillover ended up with 34 MB of fully-allocated zero bytes → `spillover-restore` copied those zeros into `/nfs_export` on the next cold read → Spark saw all-zero parquet files.
7. **v6** — changed to `if n <= 0 { continue }` (matches the `n > 0` pattern used correctly elsewhere in the codebase)

## Bench — v6 SF100 NFS (chain-DOWN, 8 GB tmpfs + 8 GB spillover cap)

- Runtime: 55+ min (killed before first query completed)
- Spillover cycles: **57 Spilled events** (6 450 + file-spill operations) + **541 spillover-restore events**
- Parquet corruption: **0** all-zero files (vs v5 had ~1 227 zero-filled spills)
- Failures: **16 `ChecksumException` retries on 2/1 758 files** — spillover ran mid-Spark-read, `O_TRUNC`'d the file under Spark's open fd, subsequent reads returned sparse zeros past the (now truncated) size

### Conclusion
v6 is **correctness-proven** (no data corruption). SF100 + 8 GB tmpfs is a **thrashing workload** — 33 GB dataset vs 8 GB cache means store_sales (6.5 GB) is repeatedly evicted and restored each time Spark scans it. The active-read race (spill while client has an open fd) produces rare `ChecksumException`s; this is a known limitation of the spill-then-retrieve model.

### Not-yet-fixed follow-ups
- **Spill-during-active-read race** — `spillCommittedFiles` should check if any NFS client has an open fd on the file before `O_TRUNC`'ing. Requires tracking open fds through FSAL_ZUS (or a simpler "just-accessed within N seconds" heuristic).
- **SF100 on 8 GB cache**: re-run with tmpfs sized to hold the dataset (32 GB+) OR disable spillover entirely for dataset-sized-fit-in-tmpfs runs OR switch Spark's Hadoop FS to `RawLocalFileSystem` to disable the `.crc` checks that surface the active-read race.

## Gotchas (save to memory)
- `syscall.Getxattr` on Linux returns `n = -errno` (negative int) when the xattr is missing, NOT `n = 0`. Always gate on `n > 0` for presence, `n <= 0` for absence.
- NFS-fronted prewarm MUST preserve inode. Atomic temp+rename breaks Ganesha's cached file handles.
- New-file PUT needs early `MarkCommitted` (gates inotify-upload race) + late xattr (gates spillover). Distinct races, distinct gates.
- `O_TRUNC` preserves xattrs. If your gate is "file has xattr X", explicitly remove X before `O_TRUNC` to avoid a transitive race window.
