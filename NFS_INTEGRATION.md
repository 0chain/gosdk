# GoSDK — NFS Gateway Integration

The zs3server NFS gateway (`zs3server/cmd/gateway/zcn/nfs_*.go`) provides NFSv3 filesystem access to blobber data. It calls GoSDK allocation methods for all filesystem operations.

## SDK Methods Used by NFS

| NFS Operation | GoSDK Method | Notes |
|--------------|-------------|-------|
| `readdir` | `getRegularRefs(alloc, path, ...)` | Returns `ObjectTreeResult` with `[]ORef` |
| `stat` | `getSingleRegularRef(alloc, path)` | Returns single `ORef` |
| `read` | `getFileReader(ctx, alloc, ...)` | Downloads via `DownloadByBlocksToFileHandler` |
| `write` | `putFile(ctx, alloc, path, reader, ...)` | Uploads via `DoMultiOperation` batch |
| `mkdir` | `alloc.DoMultiOperation(createdir)` | Creates directory on blobbers |
| `unlink` | `alloc.DeleteFile(path)` | Deletes file from blobbers |
| `rename` | `alloc.DoMultiOperation(move, rename)` | Move + rename via multi-op |

## ORef → os.FileInfo Mapping

The NFS layer converts `sdk.ORef` to `os.FileInfo`:
- `ORef.Name` → `FileInfo.Name()`
- `ORef.Size` → `FileInfo.Size()`
- `ORef.Type` — `"d"` = directory, `"f"` = file → `FileInfo.IsDir()`, `FileInfo.Mode()`
- `ORef.UpdatedAt` (common.Timestamp) → `FileInfo.ModTime()`

## Performance-Critical SDK Settings

From enterprise blobber benchmarking:
- `LockedBlobbersCap` in `writemarker_mutex.go` — controls concurrent blobber locks (default 1, set to 5 for throughput)
- `fileref/fileref.go` LRU cache size — default 100, needs 50000 for high-throughput NFS
- `sdk.BatchSize` — controls multi-operation batch size
- `sdk.SetHighModeWorkers(n)` — upload concurrency

## Related Branches

- `perf/small-file-throughput` — LockedBlobbersCap + LRU fix
- `feat/enterprise-blobber` — enterprise timing optimizations

## Architecture

See [zs3server/NFS_GATEWAY_ARCHITECTURE.md](../zs3server/NFS_GATEWAY_ARCHITECTURE.md) for the full NFS gateway design including WAL integration, caching, and S3 Files feature parity.
