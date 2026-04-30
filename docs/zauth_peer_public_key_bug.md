# ZAuth `wallet is not connected` Bug on `fix/pre-reuse-encrypted-key`

**Date**: 2026-04-26
**Branches affected**: `fix/pre-reuse-encrypted-key`, `fix/pre-reuse-folder-entropy`
**Commit at fault**: `f79678dd` (gosdk) — *"fix: use PeerPublicKey instead of Keys[0].PublicKey for zauth X-Peer-Public-Key header"*
**Branches NOT affected**: `fix/lfb-aware-sharder-selection` (uses `c.Keys[0].PublicKey`, which is correct)

## Symptom

Every split-key file upload fails with HTTP 500 from blobbers:

```
uploadObject err: upload_failed: Upload failed.
unexpected status code: 500, res: err happened wallet is not connected
```

The error originates in zauth-server: it can't find a row in `split_wallets` for the
`(client_id, peer_public_key)` pair sent by the client, so split-key signing fails and the
blobber upload aborts.

(Side bug in zauth-server: `errors.Unwrap(err) == stores.ErrWalletNotConnected` is always
false because `ErrWalletNotConnected` is returned unwrapped — should be `errors.Is`. That
turns a clean 404 into a 500. Out of scope here, but worth tracking.)

## Investigation

Reproduced in Playwright on the local dev server (`http://localhost:3430/`) hitting
`https://zauth.test.zus.network/sign/msg`. Captured every `/sign/msg` request:

| Source | `X-Peer-Public-Key` header sent | Response |
|---|---|---|
| Web Worker (per-blobber WASM) | `0cfdd78f…` (= `c.Keys[0].PublicKey`) | **200 OK** |
| Main thread WASM              | `82b5e564…` (= `c.PeerPublicKey`)   | **500 wallet not connected** |

Test wallet:
- `client_id`: `fa7c124f…`
- `Keys[0].PublicKey`: `0cfdd78f…`
- `PeerPublicKey`:    `82b5e564…`

ZAuth's DB row for this wallet: `split_wallets WHERE client_id='fa7c124f…' AND peer_public_key='0cfdd78f…'`.
**The `peer_public_key` column stores the user's pubkey** (from zauth's perspective the user
is the "peer"). So the header must equal `Keys[0].PublicKey`, not `PeerPublicKey`.

This is proven empirically: a request with `0cfdd78f` succeeds; a request with `82b5e564`
gets "wallet not connected" because no row matches.

## Root Cause

Commit `f79678dd` (Apr 5 2026, guruhubb) on `fix/pre-reuse-encrypted-key`:

```
fix: use PeerPublicKey instead of Keys[0].PublicKey for zauth X-Peer-Public-Key header

Keys[0].PublicKey is the user's split key public key (c0), but zauth
stores and looks up by the peer's public key (c1). Using the wrong key
causes 'wallet is not connected' errors on every split-key sign request.

Bug introduced in df4dbdf4 (client id support in zauth sign).
```

The reasoning is **inverted**. The "peer" in `zauth.split_wallets.peer_public_key` is the
user (from zauth's point of view), not zauth itself. The author confused the field name's
perspective.

The commit changed both `ZauthSignTxn` and `ZauthAuthCommon`:

```diff
 c := GetClient()
-pubkey := c.Keys[0].PublicKey   // CORRECT
+pubkey := c.PeerPublicKey       // WRONG: this is zauth's pubkey, not the user's
 if len(keys) > 0 {
     c = GetWalletByKey(keys[0])
     ...
-    pubkey = c.Keys[0].PublicKey
+    pubkey = c.PeerPublicKey
 }
 req.Header.Set("X-Peer-Public-Key", pubkey)
```

Reference for the correct version: commit `88fff0c5` on `fix/lfb-aware-sharder-selection`.

## Field naming reference

The naming flips between client-side gosdk, zvault, and zauth. For our test wallet:

| Location                                         | Field            | Value      |
|---|---|---|
| gosdk `client.Wallet.Keys[0].PublicKey`          | user's pubkey    | `0cfdd78f` |
| gosdk `client.Wallet.PeerPublicKey`              | zauth's pubkey   | `82b5e564` |
| zvault DB `split_keys.public_key`                | user's pubkey    | `0cfdd78f` |
| zvault DB `split_keys.peer_public_key`           | zauth's pubkey   | `82b5e564` |
| zauth DB `split_wallets.public_key`              | zauth's pubkey   | `82b5e564` |
| zauth DB `split_wallets.peer_public_key`         | **user's pubkey**| `0cfdd78f` |

Each side calls its own key `public_key` and the other side's key `peer_public_key`. When
the client sends a request to zauth, the `X-Peer-Public-Key` header must match
`zauth.split_wallets.peer_public_key`, which is the user's key — i.e. `c.Keys[0].PublicKey`
in client-side gosdk terminology.

## The Fix

Revert `f79678dd` on `fix/pre-reuse-encrypted-key` (and `fix/pre-reuse-folder-entropy`):

```diff
 // gosdk/core/client/zauth.go — both ZauthSignTxn and ZauthAuthCommon

 c := GetClient()
-pubkey := c.PeerPublicKey
+pubkey := c.Keys[0].PublicKey
 if len(keys) > 0 {
     c = GetWalletByKey(keys[0])
     if c == nil { return "", errors.Errorf(...) }
-    pubkey = c.PeerPublicKey
+    pubkey = c.Keys[0].PublicKey
 }
 req.Header.Set("X-Peer-Public-Key", pubkey)
```

Then rebuild the WASM and copy to:
- `packages/shared/public/zcn.wasm`
- `packages/vult/public/zcn.wasm`

Restart the local dev server (port 3430) so the fresh WASM is served.

## Verification

After applying the fix and rebuilding:

1. Open the app, sign in with a split-key (KMS) wallet.
2. Trigger an upload to a paid storage allocation.
3. Watch the network panel for `POST https://zauth.test.zus.network/sign/msg`:
   - Header `X-Peer-Public-Key` should equal the user's pubkey (e.g. `0cfdd78f…`), NOT zauth's pubkey (`82b5e564…`).
   - Response should be **200 OK** with `{"sig": "..."}`.
4. Upload completes; no `wallet is not connected` errors.

## Upstream coordination

Two things need to go upstream after local verification:

1. **Push the revert to `fix/pre-reuse-encrypted-key`** (and to
   `fix/pre-reuse-folder-entropy` if it's still in active use), with a commit message that
   explains why `f79678dd`'s premise was wrong. Cite the empirical evidence (workers vs.
   main thread `/sign/msg` outcomes) so a future reader doesn't re-introduce the same
   "fix".
2. **Reach out to the `f79678dd` author** (guruhubb) so the misunderstanding doesn't
   persist. The naming reference table above should clarify why `Keys[0].PublicKey` is the
   right value.

## References

- Bug commit (introduced this regression): `f79678dd`
- Correct version (kept `Keys[0].PublicKey`): `88fff0c5` on `fix/lfb-aware-sharder-selection`
- Earliest commit that added the `clientIds` parameter the bug commit blamed: `df4dbdf4`
  (Aug 19 2025, *"client id support in zauth sign"*) — but `df4dbdf4` was correct; it
  always used `c.Keys[0].PublicKey`.
