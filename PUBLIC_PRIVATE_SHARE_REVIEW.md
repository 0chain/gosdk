# Full-Stack Code Review: Public/Private Share Feature

**Date:** 2026-02-27
**Repos:** blobber PR #1551 · gosdk PR #1792 · 0box feat/public-share-with-revoke · web-apps feat/group-pub-priv-share
**Dimensions:** Correctness · Security · Scalability · Speed

---

## Table of Contents

1. [Release Blockers (Must-Fix Summary)](#release-blockers)
2. [Blobber PR #1551](#blobber-pr-1551)
3. [gosdk PR #1792](#gosdk-pr-1792)
4. [0box feat/public-share-with-revoke](#0box-featpublic-share-with-revoke)
5. [web-apps feat/group-pub-priv-share](#web-apps-featgroup-pub-priv-share)
6. [Cross-Cutting Issues](#cross-cutting-issues)

---

## Release Blockers

These issues **must be fixed before any of these branches can merge**. Each is either a data-loss risk, a security vulnerability, a crash, or completely broken functionality.

| # | Repo | Severity | Issue |
|---|------|----------|-------|
| 1 | blobber | **Critical** | `DeletePublicShareInfo` revokes private shares — missing `client_id = ''` filter wipes all shares on a path |
| 2 | blobber | **Critical** | GORM zero-value drops `revoked = false` from WHERE clause — always matches already-revoked records |
| 3 | blobber | **High** | Ownership check ordered after file-path DB lookup — path existence leaks to non-owners |
| 4 | blobber | **High** | `CheckPublicShareExists` returns true for private shares — wrong semantics |
| 5 | blobber | **High** | `GetPublicShareRecipients` returns private recipients and their re-encryption keys |
| 6 | gosdk | **Critical** | Goroutine leak on early return when request construction fails mid-loop |
| 7 | gosdk | **Critical** | Double-send to buffered error channel causes goroutine deadlock in `CheckPublicShareExists` / `GetPublicShareRecipients` |
| 8 | gosdk | **Critical** | Unguarded `data["status"].(float64)` type assertion panics and crashes the process |
| 9 | gosdk | **High** | 100% blobber consensus hard-coded — one offline blobber makes public share permanently irrevocable |
| 10 | 0box | **Critical** | `fmt.Println` before error check in `RemovePublicShareRecipient` — debug leak + nil-dereference risk |
| 11 | 0box | **Critical** | Public share creation is completely broken — `receiverID = authTicketClientID` is `""` for public tickets, hits DB NOT NULL constraint |
| 12 | 0box | **Critical** | Caller-controlled `share_info_type` allows type escalation; omitting the field inserts `UndefinedShare(0)` rows |
| 13 | 0box | **High** | `ReplaceGroupMembers` error swallowed — DELETE committed without INSERT, all group members silently deleted |
| 14 | 0box | **High** | `MaxMembers` limit not enforced during bulk create or replace operations |
| 15 | 0box | **High** | Group owner can self-remove, permanently orphaning the group with no owner |
| 16 | web-apps | **Critical** | Share executes on group suggestion click with no confirmation and no undo |
| 17 | web-apps | **Critical** | `selectedGroups` ephemeral local state — group shares are permanently unrevokable after dialog close |
| 18 | web-apps | **High** | `handleRemoveUser` can never remove pre-loaded members (filters by absent `user_name` field) |
| 19 | web-apps | **High** | `DELETE_GROUP` reducer compares `group.id !== responseBody(object)` — never removes group from Redux state |
| 20 | **cross-repo** | **Architectural** | No coordination between blobber and 0box revoke — revoking in one system leaves the other intact; a "revoked" file is still downloadable |

### Fix Guidance Per Repo

**blobber**
```go
// B-C1: Add client_id = '' to scope DeletePublicShareInfo to public-only
Where("file_path_hash = ? AND client_id = '' AND revoked = ?", hash, false)

// B-C2: Replace struct-based WHERE for booleans with explicit string WHERE
Where("client_id = ? AND file_path_hash = ? AND revoked = ?", clientID, hash, false).
    Updates(map[string]interface{}{"revoked": true})

// B-H1: Move ownership check BEFORE any DB resource lookup
if clientID != allocationObj.OwnerID {
    return nil, common.NewError("invalid_operation", "Operation needs to be performed by the owner")
}
// Now safe to call GetLimitedRefFieldsByLookupHash

// B-H2/H3: Add client_id = '' filter
Where("file_path_hash = ? AND client_id = '' AND revoked = ?", filePathHash, false)
```

**gosdk**
```go
// G-C2: Remove double-send — only send to errCh in the outer block, not inside the callback
err := zboxutil.HttpDo(a.ctx, a.ctxCancelF, httpreq, func(resp *http.Response, err error) error {
    if err != nil {
        return err  // do NOT send to channel here
    }
    return nil
})
if err != nil {
    errCh <- err  // only here
}

// G-C3: Guard type assertion
if v, ok := data["status"].(float64); ok && v == float64(http.StatusNotFound) {
    notFound <- 1
}

// G-H1: Use Consensus struct instead of 100%
consensus := Consensus{
    RWMutex: &sync.RWMutex{},
    consensus: len(success),
    consensusThresh: a.DataShards,
    fullconsensus:   a.fullconsensus,
}
if !consensus.isConsensusOk() {
    return errors.New("", "consensus not reached")
}
```

**0box**
```go
// Z-C1: Remove fmt.Println; check error first
publicShare, err := s.shareInfoRepo.GetPublicShareByOwner(ctx, clientID, lookupHash, appType)
if err != nil {
    logger.Logger0box.Error("Error getting public share", zap.Error(err))
    return nil, common.NewError("share_not_found", "Public share not found")
}

// Z-C2: Restore receiverID derivation from auth ticket
if authTicketClientID == "" {
    shareType = modelV2.Public
    receiverID = authHeader.ClientID  // requester is recipient for public shares
} else {
    shareType = modelV2.Private
    receiverID = authTicketClientID
}

// Z-C3: Derive shareType from ticket; reject conflicting caller-supplied type
// Do not: shareType := modelV2.ToShareType(shareinfo.ShareInfoType)

// Z-H2: Return error from ReplaceGroupMembers
if err != nil {
    return nil, common.NewError("500", "failed to replace group members")
}
```

**web-apps**
```js
// W-C1: Move actual share execution to handleShare button click
// On suggestion click: only add to pendingGroups list
const addGroup = (group) => {
    setPendingGroups(prev => [...prev, group])
}
// In handleShare: execute the actual sharing
const handleShare = async () => {
    for (const group of pendingGroups) { await shareGroupMembers(group) }
}

// W-H3: Fix handleRemoveUser to match member_name field
const handleRemoveUser = userName => {
    setSelectedUsers(selectedUsers.filter(u => (u.user_name || u.member_name) !== userName))
}

// W-H4: Fix DELETE_GROUP reducer — pass groupId via assets
case types.DELETE_GROUP_SUCCESS:
    return {
        ...state,
        myGroups: state.myGroups.filter(group => group.id !== action.assets?.groupId),
    }
```

---

## Blobber PR #1551

### Critical

#### B-C1 — `DeletePublicShareInfo` revokes private shares
**File:** `reference/shareinfo.go:726`
**Dimension:** Correctness / Security

Public and private shares are distinguished solely by `client_id = ''` (empty for public, non-empty for private). The new `DeletePublicShareInfo` function only filters on `file_path_hash`:

```go
// Current — revokes ALL shares (public and private) for the path
result := db.Model(&ShareInfo{}).
    Where(&ShareInfo{
        FilePathHash: shareInfo.FilePathHash,
        Revoked:      false,
    }).
    Updates(ShareInfo{Revoked: true})
```

Revoking a public share silently revokes every private (targeted) share on the same file path.

**Fix:**
```go
result := db.Model(&ShareInfo{}).
    Where("file_path_hash = ? AND client_id = '' AND revoked = ?", shareInfo.FilePathHash, false).
    Updates(map[string]interface{}{"revoked": true})
```

---

#### B-C2 — GORM zero-value silently drops `revoked = false` from WHERE
**File:** `reference/shareinfo.go:706,726`
**Dimension:** Correctness

GORM v2 ignores zero-value struct fields in `Where(&ShareInfo{..., Revoked: false})`. `bool false` is Go's zero value, so the generated SQL omits the `revoked = false` predicate entirely. Already-revoked records are re-matched. When `RowsAffected=0` results, the code returns `gorm.ErrRecordNotFound` — a misleading error that misattributes the cause.

**Fix:** Use explicit string WHERE conditions for any boolean field:
```go
Where("client_id = ? AND file_path_hash = ? AND revoked = ?", clientID, hash, false).
    Updates(map[string]interface{}{"revoked": true})
```

---

### High

#### B-H1 — Ownership check ordered after file-existence DB lookup
**File:** `handler/handler.go` (all 4 new handlers)
**Dimension:** Security

Handler flow in all 4 new endpoints:
1. Verify allocation
2. Verify signature
3. `GetLimitedRefFieldsByLookupHash(...)` ← **DB lookup that reveals path existence**
4. `if clientID != allocationObj.OwnerID { return error }`

If any future authentication bypass is introduced, a non-owner can probe whether file paths exist by observing the error message difference between "Invalid file path" (path does not exist) and "Operation needs to be performed by the owner" (path exists).

**Fix:** Move the ownership check to immediately after signature verification, before any DB resource lookup.

---

#### B-H2 — `CheckPublicShareExists` counts private shares — wrong semantics
**File:** `reference/shareinfo.go:675`
**Dimension:** Correctness / Security

```go
err := db.Model(&ShareInfo{}).
    Where("file_path_hash = ? AND revoked = ?", filePathHash, false).
    Count(&count).Error
```

No `client_id = ''` filter. The API semantics promise "check if a *public* share exists" but the implementation returns `true` if *any* share (public or private) exists for that hash. An owner is incorrectly told their file is publicly shared when it was only shared privately with one recipient.

**Fix:** Add `AND client_id = ''` to the WHERE clause.

---

#### B-H3 — `GetPublicShareRecipients` returns private recipients + re-encryption keys
**File:** `reference/shareinfo.go:691`
**Dimension:** Security

Without `client_id = ''`, private share recipients appear in the public recipients listing. Full `ShareInfo` rows are returned including `re_encryption_key` and `client_encryption_public_key` — sensitive proxy re-encryption keys that must not be exposed publicly.

**Fix:**
```go
err := db.Model(&ShareInfo{}).
    Select("id", "owner_id", "client_id", "file_path_hash", "expiry_at", "available_at", "revoked").
    Where("owner_id = ? AND file_path_hash = ? AND client_id = '' AND revoked = ?",
        ownerID, filePathHash, false).
    Find(&recipients).Error
```

---

### Medium

#### B-M1 — Parameter name mismatch: `recipientClientId` vs swagger `recipient_client_id`
**File:** `handler/handler.go:~1943`
**Dimension:** Correctness

Handler reads `common.GetField(r, "recipientClientId")` (camelCase `d`) but swagger documents `recipient_client_id` (snake_case). Any client following the spec gets "Invalid recipient client ID". For DELETE requests, `TryParseForm` only processes POST/PUT/PATCH bodies, so the parameter must be a URL query string — this is undocumented.

---

#### B-M2 — HTTP 200 returned for not-found and no-content conditions
**File:** `handler/handler.go` (all 4 handlers)
**Dimension:** Correctness

All 4 handlers embed a status code in the JSON body (`{"status": 404}`) while returning HTTP 200. The existing `RevokeShare` returns proper HTTP 404. This breaks REST convention and forces callers to parse JSON to detect errors.

---

#### B-M3 — `GetPublicShareRecipients` is unbounded — no LIMIT/pagination
**File:** `reference/shareinfo.go:691`
**Dimension:** Scalability / Speed

`.Find(&recipients)` with no `LIMIT` clause. A widely-shared file returns an unbounded result set consuming arbitrary memory and producing a large JSON payload. Contrast with `ListShareInfoClientID` in the same file which correctly applies `Limit` and `Offset`.

---

#### B-M4 — No `file_path_hash` index for new query patterns — full table scan
**File:** `reference/shareinfo.go` + migrations
**Dimension:** Scalability

`CheckPublicShareExists` and `DeletePublicShareInfo` filter with `file_path_hash` as the leading column. Existing indexes are `(owner_id, file_path_hash)` and `(client_id, file_path_hash)` — neither covers a `file_path_hash`-first query.

**Fix:**
```sql
-- +goose Up
CREATE INDEX idx_marketplace_share_info_file_path_hash
    ON marketplace_share_info (file_path_hash)
    WHERE revoked = false;
```

---

#### B-M5 — Zero test coverage for all 4 new handlers
**File:** `handler/handler_share_test.go`
**Dimension:** Correctness

`TestHandlers_Share` covers only `InsertShare`, `RevokeShare`, and `ListShare`. None of the 4 new handlers have unit or integration tests.

---

### Low

#### B-L1 — `//nolint:unused` suppresses dead code instead of removing it
Various non-handler files. Inconsistent spacing (`//nolint` vs `// nolint`) matters — `golangci-lint` requires no-space form.

#### B-L2 — Shell scripts missing trailing newline
`docker.local/bin/build.*.sh` — POSIX requires text files to end with a newline.

---

## gosdk PR #1792

### Critical

#### G-C1 — Goroutine leak on early return from request-construction loop
**File:** `zboxcore/sdk/allocation.go` — `RevokePublicShare`, `RemovePublicShareRecipient`
**Dimension:** Correctness

If `NewRevokePublicShareRequest` fails for blobber N, the function returns immediately while goroutines already launched for blobbers 0..N-1 are still running. `wg.Wait()` is never called. Partial revokes are silently discarded.

**Fix:** Pre-build all requests and check all errors before launching any goroutines. Or move request construction inside each goroutine.

---

#### G-C2 — Double-send to buffered error channel causes goroutine deadlock
**File:** `zboxcore/sdk/allocation.go` — `CheckPublicShareExists`, `GetPublicShareRecipients`
**Dimension:** Correctness

Each failing goroutine sends to the `errors` channel twice: once inside the `HttpDo` callback, and again in the outer `if err != nil` block (because `HttpDo` returns the callback's return value). With N blobbers all failing, up to 2N sends are attempted on a capacity-N channel. Goroutines block forever inside `wg.Wait()` — permanent deadlock.

**Fix:** Remove the error-channel sends from inside the callback. Only send in the outer `if err != nil` block.

---

#### G-C3 — Unguarded type assertion panics in goroutines
**File:** `zboxcore/sdk/allocation.go` — `RevokePublicShare`, `RemovePublicShareRecipient`
**Dimension:** Correctness / Reliability

```go
// Panics if "status" key is absent (nil) or a non-float64 type
if data["status"].(float64) == http.StatusNotFound {
```

A missing key returns `nil`; `nil.(float64)` panics and crashes the process. This executes inside goroutines with no `recover`.

**Fix:**
```go
if v, ok := data["status"].(float64); ok && v == float64(http.StatusNotFound) {
    notFound <- 1
}
```

---

### High

#### G-H1 — 100% consensus hard-coded for write operations
**File:** `zboxcore/sdk/allocation.go` — `RevokePublicShare`, `RemovePublicShareRecipient`
**Dimension:** Correctness / Security

```go
if len(success) == len(a.Blobbers) { ... }
return errors.New("", "consensus not reached")
```

A single temporarily-offline blobber makes public share revocation permanently impossible — a security issue since the share can never be fully revoked. Every other mutation in the SDK uses the `Consensus` struct with `consensusThresh = a.DataShards`.

**Fix:** Replace with the `Consensus` struct consistent with the rest of the SDK.

---

#### G-H2 — Any single blobber failure short-circuits read operations
**File:** `zboxcore/sdk/allocation.go` — `CheckPublicShareExists`, `GetPublicShareRecipients`
**Dimension:** Correctness

```go
select {
case err := <-errors:
    return false, err  // fails if ANY single blobber errors
default:
}
```

One failing blobber causes the entire operation to return an error even if N-1 blobbers succeeded. Apply a configurable threshold.

---

#### G-H3 — `fmt.Errorf(string(respbody))` — format string injection
**File:** `zboxcore/sdk/allocation.go` — all 4 new methods
**Dimension:** Correctness

If a blobber response body contains `%` characters (e.g., `"50% disk full"`), `fmt.Errorf` interprets them as format verbs and produces garbled output like `"50%!d(MISSING)isk full"`. The first argument to `fmt.Errorf` must be a literal format string.

**Fix:** `fmt.Errorf("%s", string(respbody))` or `errors.New("", string(respbody))`.

---

#### G-H4 — `errors` channel variable shadows the `errors` import package
**File:** `zboxcore/sdk/allocation.go` — `CheckPublicShareExists`, `GetPublicShareRecipients`
**Dimension:** Correctness

```go
errors := make(chan error, len(a.Blobbers))
```

Shadows `github.com/0chain/errors` for the entire function scope. Any future call to `errors.New(...)` inside these functions would silently operate on the channel instead of the package.

**Fix:** Rename to `errCh`.

---

### Medium

#### G-M1 — URL construction must be confirmed against blobber routing
**File:** `zboxcore/zboxutil/http.go`
All 4 new request constructors use trailing-slash constants (`"/v1/marketplace/shareinfo/public/"`) combined with `allocationTx` as a path segment via `joinUrl`. Confirm the blobber routes accept `DELETE /v1/marketplace/shareinfo/public/{tx}` not a query parameter form.

---

#### G-M2 — `ShareInfo` struct defined locally — will silently drift from blobber schema
**File:** `zboxcore/sdk/allocation.go`
A local copy of the blobber `ShareInfo` struct. Field additions, renames, or JSON tag changes on the blobber side will produce silent partial data with no compile-time error.

---

#### G-M3 — `GetPublicShareRecipients` returns first responding blobber with no consistency check
**File:** `zboxcore/sdk/allocation.go`
```go
select {
case recipients := <-success:
    return recipients, nil  // non-deterministic, no majority validation
default: ...
}
```
One Byzantine or lagging blobber controls the result. Implement majority-vote or hash-based comparison.

---

#### G-M4 — No per-blobber request timeout — indefinite blocking possible
**File:** `zboxcore/sdk/allocation.go` — all 4 new methods
`a.ctx` has no deadline. A hanging TCP connection holds the goroutine and `wg.Wait()` forever.

**Fix:**
```go
reqCtx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
defer cancel()
err := zboxutil.HttpDo(reqCtx, cancel, httpreq, ...)
```

---

### Low

#### G-L1 — New methods not exported to mobilesdk or winsdk
`mobilesdk/zbox/storage.go` and `winsdk/storage.go` export `RevokeShare` but not the 4 new public-share variants.

#### G-L2 — Zero tests for all 4 new SDK methods
No new `*_test.go` files. `CheckPublicShareExists` in particular has novel consensus logic that needs test coverage for all-success, partial-failure, all-failure, and empty-blobbers cases.

#### G-L3 — Three inconsistent consensus models across four related methods
`RevokePublicShare` = 100%, `CheckPublicShareExists` = simple majority, `GetPublicShareRecipients` = first-wins. Undocumented and inconsistent with the `Consensus` struct used elsewhere.

#### G-L4 — `notFound` channel logic silently swallows partial not-found
A blobber returning HTTP 200 + `{"status": 404}` is counted as both success and not-found. If 1 of 3 blobbers reports not-found, `nil` is returned — the partial not-found is swallowed. This may be intentional but is undocumented.

---

## 0box feat/public-share-with-revoke

**Changed files:** `entity/shareinfo.go` (+108/-5), `handler/shareinfo.go` (+262/-9), `entity/group.go` (+422 new), `handler/group.go` (+832 new), `router/group.go` (+608 new), `router/shareinfo.go` (+215/-1), `router/handler.go` (+22), `goose/migrations/1736000000_create_groups.sql` (+101 new).

### Critical

#### Z-C1 — `fmt.Println` before error check — debug leak + nil-dereference risk
**File:** `handler/shareinfo.go:523–524`
**Dimension:** Correctness / Reliability

```go
publicShare, err := s.shareInfoRepo.GetPublicShareByOwner(ctx, clientID, lookupHash, appType)
fmt.Println("Error getting public share", zap.Error(err), publicShare.ShareInfoType.String())
if err != nil { ... }
```

`fmt.Println` runs before the error check on every call — including success paths — leaking internal state to stdout. If the return contract of `GetPublicShareByOwner` is ever changed to return `nil` on error (idiomatic Go), `publicShare.ShareInfoType.String()` dereferences a nil pointer.

**Fix:**
```go
publicShare, err := s.shareInfoRepo.GetPublicShareByOwner(ctx, clientID, lookupHash, appType)
if err != nil {
    logger.Logger0box.Error("Error getting public share", zap.Error(err))
    return nil, common.NewError("share_not_found", "Public share not found")
}
```

---

#### Z-C2 — Public share creation completely broken
**File:** `handler/shareinfo.go:111–138`
**Dimension:** Correctness

Staging (correct):
```go
shareType := modelV2.Public
receiverID := clientID          // requester is recipient for public shares
if authTicketClientID != "" {
    shareType = modelV2.Private
    receiverID = authTicketClientID
}
```

This branch (broken):
```go
receiverID := authTicketClientID   // "" for public auth tickets
```

For public auth tickets, `decodedAuthTicket.ClientID` is `""`. The code calls `ownerRepo.GetByClientID(ctx, "")` which fails every time. Even if it somehow succeeded, `ShareInfoEntity.Receiver = ""` violates the `receiver_client_id NOT NULL` DB constraint. **All public share creation attempts fail.**

**Fix:** Restore the staging derivation logic.

---

#### Z-C3 — Caller-controlled `share_info_type` allows type escalation
**File:** `handler/shareinfo.go:95`
**Dimension:** Security

```go
shareType := modelV2.ToShareType(shareinfo.ShareInfoType)
```

`ShareInfoType` is a user-supplied form field. A caller can pass `share_info_type=public` with a private auth ticket (creating a public-tagged record with a specific recipient) or `share_info_type=private` with a public ticket. If omitted, `ToShareType("")` returns `UndefinedShare(0)`, which bypasses `ShareInfoExists` checks and inserts a row with `share_info_type=0` — an invalid state no query handles.

**Fix:** Derive `shareType` from the auth ticket's `ClientID` field (not from caller input). Optionally reject if caller-supplied type conflicts with the derived type.

---

### High

#### Z-H1 — Auth ticket signature never cryptographically verified
**File:** `handler/shareinfo.go:626–635`
**Dimension:** Security (pre-existing, worsened by this branch)

`ConvertToAuthTicket` decodes and unmarshals the ticket but never verifies `Signature`. Combined with Z-C3 (caller-controlled share type) and Z-C2 (receiverID from ticket), a forged ticket can claim ownership of any file or redirect a share to a victim.

**Fix:** Verify the ticket signature against the blobber's known public key. At minimum, cross-check the ticket's `OwnerID` against the `senderID` already retrieved from the DB.

---

#### Z-H2 — `ReplaceGroupMembers` swallows errors — DELETE committed, INSERT not — all members silently deleted
**File:** `handler/group.go:424–428`
**Dimension:** Correctness / Data Integrity

```go
_, err := h.groupMemberRepo.ReplaceGroupMembers(ctx, groupID, members)
if err != nil {
    logger.Logger0box.Error("Cannot add group members", zap.Error(err))
    // we don't return error here, group is still created
}
```

`ReplaceGroupMembers` deletes all existing members then inserts the new list in a single transaction. If the INSERT fails, the handler ignores the error and returns HTTP 200. `WithConnection` middleware sees 200 and commits the transaction — members deleted, none inserted.

**Fix:** Return the error so the middleware rolls back:
```go
if err != nil {
    return nil, common.NewError("500", "failed to replace group members")
}
```

---

#### Z-H3 — `MaxMembers` limit not enforced during bulk operations
**File:** `handler/group.go:217–224, 423–429`
**Dimension:** Correctness

`AddMember` (single) correctly checks the limit. `CreateGroupWithMembers` calls `AddGroupMembers` (bulk) without any capacity check, and `UpdateGroupWithMembers` calls `ReplaceGroupMembers` without any capacity check. A caller can bypass the per-group limit with any request body.

**Fix:**
```go
if len(members) > group.MaxMembers {
    return nil, common.NewError("400", fmt.Sprintf("member count %d exceeds maximum %d", len(members), group.MaxMembers))
}
```

---

#### Z-H4 — Group owner can self-remove, orphaning the group
**File:** `handler/group.go:696–712`
**Dimension:** Correctness

`RemoveMember` guard only prevents admins from removing the owner; the owner can remove themselves. After self-removal the group has no owner, `DeleteGroup` cannot find the group by `owner.ID`, and the group becomes permanently unmanageable.

**Fix:** Block self-removal if the caller is the owner:
```go
if targetMember.Role == modelV2.GroupRoleOwner {
    return nil, common.NewError("403", "cannot remove group owner; transfer ownership first")
}
```

---

### Scalability

#### Z-S1 — N+1 queries in all group listing endpoints
**File:** `handler/group.go`
**Dimension:** Scalability

Every listing endpoint (`GetMyGroups`, `GetPublicGroups`, `SearchGroups`, `GetUserGroups`) calls `GetTotalMemberCount` per group in a loop. 50 groups = 51 DB round-trips per request. Under load: 100 concurrent users = 5100 queries/sec for a single endpoint.

**Fix:** Batch with a single `COUNT(*) GROUP BY group_id` query for all group IDs in the page.

---

#### Z-S2 — `ILIKE '%query%'` prevents index use — full table scan on every search
**File:** `entity/group.go:228–238`
**Dimension:** Scalability / Speed

A leading `%` wildcard prevents use of any btree index. The migration creates no GIN/GiST index for full-text search.

**Fix:**
```sql
CREATE INDEX idx_groups_fts ON public.groups
    USING gin(to_tsvector('english', group_name || ' ' || coalesce(description, '')));
```
```go
db.Where("to_tsvector('english', group_name || ' ' || coalesce(description, '')) @@ plainto_tsquery(?)", query)
```

---

#### Z-S3 — Missing composite index on `shareinfo(client_id, lookup_hash, app_type, share_info_type)`
**Dimension:** Scalability

All 5 new share queries filter on this combination. Existing indexes are single-column. Add a migration:
```sql
CREATE INDEX idx_shareinfo_owner_lookup_type
    ON public.shareinfo USING btree (client_id, lookup_hash, app_type, share_info_type)
    WHERE deleted IS NULL;
```

---

#### Z-S4 — `idx_groups_owner_id` not partial — index bloat from soft deletes
**Dimension:** Scalability (Low)

The unique index `groups_owner_name` correctly excludes soft-deleted rows with `WHERE deleted IS NULL`, but the plain `idx_groups_owner_id` does not. Delete-heavy workloads cause index bloat and slower scans.

---

### Speed

#### Z-P1 — `GetGroupWithMembers` loads all members + redundant count call
**File:** `entity/group.go`
**Dimension:** Speed

`Preload("Members")` can load up to 1000 `GroupMemberEntity` rows into memory. `GetGroup` then calls `GetTotalMemberCount` separately, ignoring that `len(group.Members)` is already available.

**Fix:** Use `len(group.Members)` instead of the extra query. Add pagination to the Preload or switch to a separate paginated member query.

---

#### Z-P2 — `GetPublicShareRecipients` fetches full rows including large `auth_ticket text`
**File:** `entity/shareinfo.go`
**Dimension:** Speed (Low)

`SELECT *` via implicit GORM `.Find()` returns the large `auth_ticket text` column on every row. Add `.Select(...)` to limit columns to what the response actually uses.

---

### Code Quality / Low

| ID | File | Issue |
|----|------|-------|
| Z-M1 | `handler/group.go` | `CheckGroupExists` is global; DB uniqueness is per-owner — prevents user B from creating "My Files" if user A has one |
| Z-M2 | `handler/shareinfo.go` | Duplicate dead-code auth ticket empty check (`len==0` then `==""`) |
| Z-M3 | `entity/shareinfo.go` | `GetPublicShareRecipients` has no LIMIT |
| Z-M4 | `handler/group.go` | `wallet_address` not validated against blockchain format |
| Z-M5 | `handler/shareinfo.go` | `DeleteByShareID` existence check ignores `app_type` — cross-app-type information disclosure |
| Z-M6 | `handler/shareinfo.go` | Unreachable `ShareInfoType != Public` check after `GetPublicShareByOwner` (which already filters by Public) |
| Z-L1 | `router/group.go` | `SearchMyGroups` HTTP handler defined but not registered — dead code |
| Z-L2 | `router/group.go` | Wrong Swagger `@Router /v2/groups/{id} [put]` annotation on `UpdateGroupWithMembers` |
| Z-L3 | `handler/group.go` | `MaxMembers int` without `omitempty` — `min=1` binding rejects omitted field (Go default = 0) |
| Z-L4 | `handler/group.go` | Typo `UpdateGroupWithMemebrs` in public service interface |
| Z-L5 | `handler/group.go` | `CreateGroup` service method is dead code (router calls `CreateGroupWithMembers`) |
| Z-L6 | `router/group.go` | `CreateGroupWithMembers` HTTP handler is dead code (not registered) |

---

## web-apps feat/group-pub-priv-share

**Files changed:** 18 (2 modified, 16 new)
**Scope:** Group management UI (CRUD), group-sharing integration in ShareDialog, new Redux slice.

### Critical

#### W-C1 — Share executes on suggestion click with no confirmation and no undo
**File:** `ShareDialog.js:648`
**Dimension:** Correctness / UX

```js
onClick={() => addGroup(group)}
```

Clicking a group in the autocomplete dropdown immediately calls `shareObject` + `addPrivateLink` + `updateShareInfo` for every group member — all irreversibly. The "Share" button is a no-op (`handleShare` only shows a toast and closes the dialog; the actual sharing already happened). A user who accidentally clicks the wrong group has no way to undo.

**Fix:** On suggestion click, add to a `pendingGroups` list only. Execute actual sharing when "Share" is clicked.

---

#### W-C2 — `selectedGroups` ephemeral — group shares permanently unrevokable after dialog close
**File:** `ShareDialog.js:102, 670`
**Dimension:** Correctness / Data Integrity

```js
const [selectedGroups, setSelectedGroups] = useState([])
```

Resets to `[]` every time the dialog opens (line 147–152 in the `isOpen` effect). Groups shared in a prior session never appear in the "Shared With" tab. There is also a dead state variable `sharedWithGroups` that was declared but never used — evidence that persistence was planned but not implemented.

**Fix:** Persist group share metadata to 0box via an endpoint, reload on dialog open (same pattern as `cFPrivateLinks` loaded from Redux).

---

#### W-C3 — Sequential `for-await` member revocations — severe UI blocking
**File:** `ShareDialog.js:573–591`
**Dimension:** Speed / UX

```js
for (const member of group.members) {
    await handleDelete(member.userName, ...)  // serial, one at a time
}
```

50 members × 300ms each = 15 seconds of UI freeze. `handleResetFilePermissions` compounds this by also iterating groups and users serially.

**Fix:** Use `Promise.all` for member revocations. Parallelize across groups too.

---

### High

#### W-H1 — `window.store?.getState()` anti-pattern — stale wallet data
**File:** `CreateGroupDialog.js:86`, `UpdateGroupDialog.js:123`
**Dimension:** Correctness

Reads Redux state via the global window object, bypassing React subscriptions. Returns stale data if Redux has updated since last render. Fails in test environments and SSR.

**Fix:** `const activeWallet = useSelector(state => state.wallet?.activeWallet)`

---

#### W-H2 — `member.user_id || member.member_name` — wrong field for username lookup
**File:** `ShareDialog.js:316`
**Dimension:** Correctness

Used to call `getPublicEncryptionKey`. `user_id` is an opaque UUID; the blobber username lookup requires `member_name`. Encryption key lookups fail silently.

---

#### W-H3 — `handleRemoveUser` can never remove pre-loaded members
**File:** `UpdateGroupDialog.js:48–54, 111–113`
**Dimension:** Correctness

Pre-loaded existing members are mapped without a `user_name` field. `handleRemoveUser` filters by `u.user_name !== userName`; since `u.user_name` is always `undefined` for pre-loaded members, the predicate is always `true` and the member is never removed. The "×" button silently does nothing.

**Fix:** Filter by `(u.user_name || u.member_name) !== userName`.

---

#### W-H4 — `DELETE_GROUP` reducer never prunes Redux state
**File:** `reducer.js:3119`
**Dimension:** Correctness

```js
case types.DELETE_GROUP_SUCCESS:
    return {
        ...state,
        myGroups: state.myGroups.filter(group => group.id !== action.payload),
```

`action.payload` is the raw JSON response body from the DELETE endpoint (e.g. `{}` or `null`), not the group ID. `group.id !== {}` is always `true`. Deleted groups remain in the UI until next full refresh.

**Fix:** Pass `assets: { groupId }` to `basicReqWithDispatch` and read `action.assets.groupId` in the reducer.

---

#### W-H5 — `_REQUEST` action dispatched twice per thunk
**File:** `actions.js:13, 39, 69, 106, 155, 209`
**Dimension:** Correctness

Every action thunk dispatches `{ type: actionTypes.request }` manually, then calls `basicReqWithDispatch` which also dispatches it internally. The reducer fires twice per call.

---

#### W-H6 — `getGroupMembers` missing `assets: { groupId }` — members stored under `undefined` key
**File:** `actions.js:90`, `reducer.js:3073`
**Dimension:** Correctness

Without `assets: { groupId }`, `action.assets` is `{}`. The fallback `action.payload?.[0]?.group_id` fails for paginated envelope responses. All members end up at `groupMembers[undefined]` in Redux.

---

#### W-H7 — Operator precedence bug — "already shared" message never shown
**File:** `ShareDialog.js:617`
**Dimension:** Correctness

```js
if (isAlreadyShared && !userSuggestions?.length > 0) {
```

Parses as `isAlreadyShared && ((!userSuggestions?.length) > 0)`. The "already shared" message is never rendered.

**Fix:** `if (isAlreadyShared && !userSuggestions?.length) {`

---

### Medium

| ID | File | Line | Issue |
|----|------|------|-------|
| W-M1 | `ShareDialog.js` | 480–483 | Debounce instance recreated without cancelling prior timer — stale closure callbacks fire |
| W-M2 | `reducer.js` | 3042–3097 | Single global `loading` flag for 6 async operations — all group UI shows "loading" when any one is in-flight |
| W-M3 | `ShareDialog.js` | 598–605 | `handleShare` is a no-op — share already committed on dropdown click |
| W-M4 | `ShareDialog.js` | 870–897 | Copy-link, social share, preview image selector removed from shared-with list — functional regression |
| W-M5 | All groups components | — | All UI strings hard-coded English, not going through `useTranslation` |
| W-M6 | `ShareDialog.js` | 106 | `sharedWithGroups` state declared but never used — dead state |
| W-M7 | `ShareDialog.js` | 36 | `getMyGroups` imported but never called |

---

### Low

| ID | File | Issue |
|----|------|-------|
| W-L1 | `AllGroupsSection.js:1494–1504` | `handleEdit` and `handleViewEdit` are byte-for-byte identical functions |
| W-L2 | Multiple | Array index used as `key` for mutable lists — causes incorrect DOM diffing when items removed |
| W-L3 | `actions.js:18` | Default limit of 20 groups/members with no pagination UI — users with >20 groups see truncated list silently |
| W-L4 | `CreateGroupDialog.js:1852` | No group name length or character validation |
| W-L5 | `ShareDialog.js:1013` | `isResetting` state not used to disable "Revoke All" button — double-submit possible |
| W-L6 | `AllGroupsSection.js:1532` | Grammar error: "All Group" should be "All Groups" |

---

## Cross-Cutting Issues

### No coordination between blobber and 0box revoke state
Revoking through 0box (`DELETE /v2/shareinfo/public/revoke`) only deletes the 0box record — the blobber auth ticket remains valid and the file is still downloadable. Revoking through gosdk/blobber directly has no effect on 0box records. A "revoked" file is still accessible from the blobber; a file shown as "revoked" in 0box may never have been revoked at the storage layer.

**Required:** Define a single source of truth. The 0box revoke endpoint must call the gosdk/blobber revoke, or the blobber revoke must notify 0box. The two systems are currently fully independent.

### Three inconsistent consensus models in gosdk
`RevokePublicShare` = 100%, `CheckPublicShareExists` = majority, `GetPublicShareRecipients` = first-wins. None are documented or consistent with the `Consensus` struct used elsewhere in the SDK.

### PRs are not wired together
The web-apps branch calls none of the new WASM functions added in gosdk PR #1792. The blobber endpoints are unreachable from the UI. The PR descriptions do not cross-reference each other.

### "Public share" maps to different concepts across repos
In the blobber: `client_id = ''` (empty recipient) share. In 0box/web-apps: group-visible file share. This terminology ambiguity makes reasoning across the codebases error-prone.

### Zero integration or system tests across all four repos for these flows

---

## Finding Counts by Repo

| Repo | Critical | High | Medium | Low | Total |
|------|----------|------|--------|-----|-------|
| blobber | 2 | 3 | 5 | 2 | **12** |
| gosdk | 3 | 4 | 4 | 4 | **15** |
| 0box | 3 | 4 | 10 | 7 | **24** |
| web-apps | 3 | 7 | 7 | 6 | **23** |
| **Total** | **11** | **18** | **26** | **19** | **74** |
