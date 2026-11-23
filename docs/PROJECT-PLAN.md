# Distributed Object Store

Planning draft for a Linux or WSL prototype. Effort estimates cover future implementation and verification.

## Product requirements

**User and problem.** A developer learning distributed storage needs to see what happens to a real object when individual storage nodes disappear. They should be able to upload a small synthetic file, locate its data and parity shards, take two nodes offline, download the same bytes, and repair the missing shards. A local browser inspector makes each step visible.

**First milestone.** Use an isolated development store to upload a fixture through the browser, refresh the listing, download the fixture, and inspect its size, checksum, version and placement. Extend the same flow to six independent loopback node processes with durable writes, degraded reads and explicit repair. The browser identifies the active storage adapter.

**User stories**

- As a developer, I can upload a file and inspect its version, original size, SHA-256 and shard placement so I can trace where it went.
- As a developer, I can disable any two local nodes and still download the exact bytes from a committed version. I can see which nodes were unavailable.
- As a developer, I can replace the missing nodes and run repair for one version. I can check that repair preserved the version ID, bytes and committed object checksum.
- As a developer, I can see a write fail clearly when a node is unavailable, then retry safely without creating a duplicate version.
- As a developer, I can request a small portion of an object, see useful errors for out-of-range requests, and inspect state after a coordinator restart.

**Completion criteria.** A clean local environment can run the coordinator, a six-node synthetic cluster and the browser flow using a preinstalled Go toolchain and the standard library. A committed object survives all 15 possible pairs of missing nodes without changing its bytes. A failed write leaves the previous head readable or reports an ambiguous outcome that can be safely resolved by retrying. The resulting checks document their specific local failure model.

**Release boundary.** The first milestone supports one local namespace, multiple immutable versions per key, one closed byte range per request and manual repair. Authentication, IAM, Signature V4, AWS accounts, cloud storage, multipart upload, TLS, multi-host operation and automatic balancing belong to later work. S3-style HTTP examples cover only individually documented and tested operations. AWS SDK integration and general S3 compatibility each require a separate milestone.

## Technical design

### Processes and API

Use a Go HTTP coordinator and six loopback node processes with fixed IDs. Each worker owns one configured storage directory. The coordinator owns a separate metadata directory and holds an exclusive store lock. The browser uses same-origin HTML and JavaScript. Local runners use fresh application-owned stores and synthetic bytes. Constrain browser mutations to the exact configured origin and require a session nonce for browser write operations. Bind every listener to loopback.

Use paths below a fixed local namespace. Object keys are 1–128 ASCII letters, digits, dots, underscores or hyphens. Reject dot segments and separators, including percent-encoded separators, and decode only once. Use internally generated opaque version IDs and fixed shard IDs for on-disk names. Derive every shard path from those internal IDs.

| Operation | Contract |
|---|---|
| PUT object | Accept a bounded non-empty body and a client-supplied idempotency key. Return version, byte length and SHA-256 on commit. Serialize writes for the same key. |
| GET object | Resolve the current head or an explicitly supplied version. Pin one immutable manifest for the read. Return the exact bytes with an identifying version header. |
| GET a byte range | Support one closed inclusive range such as `bytes=1-2`. Clamp an end beyond the object length. Return 416 with the total size for a valid range starting past the end. Reject malformed, multiple, suffix and open-ended ranges as unsupported. |
| List and inspect | Return bounded key and version lists, manifest details and observed node availability. Mark node reachability separately from shard integrity that has been checked. |
| POST repair | Target one existing version and the unavailable, restored or corrupt roles identified for it. Report incomplete work if a target node is still offline. |

Start with an 8 MiB request-body limit, two concurrent data operations, a 2-second individual node-request deadline and a 15-second data-operation deadline. Limit incoming headers to 16 KiB. Reserve at most 32 keys, 16 committed versions per key and 64 MiB of shard data including staged bytes, using checked arithmetic and quota reservation before writing. Reject writes exceeding bounds. Release transient reservations on clean rollback and re-account persisted staging and orphan bytes on restart. Operators create fresh scratch stores for repeated fault runs and the prototype reports orphan candidates. General garbage collection belongs to later work.

### Erasure coding and integrity

Use a fixed `rs-4-2-v1` profile with working arithmetic in GF(2^8) and irreducible polynomial 0x11d. Each stripe holds up to 256 KiB of object bytes. Split its contiguous data into four equal columns, padding the last stripe with zeroes and recording its original length. With 1–262,144 original stripe bytes, column width is ceiling(stripe bytes / 4) and stays at or below 64 KiB. Four data columns and two parity columns occupy six fixed node roles. The maximum raw object is 32 full stripes and uses 12 MiB of shard bytes, before filesystem and metadata overhead. [IETF RFC 5510, field polynomials](https://datatracker.ietf.org/doc/html/rfc5510#section-8.1)

For each column byte, P is XOR of the data bytes. Q is the finite-field weighted sum with coefficients 1, 2, 4 and 8. With two data erasures, the two surviving parity equations have nonzero determinant because their weights differ. With one data erasure and one parity erasure, either surviving parity equation has a nonzero coefficient for the lost data. With two parity erasures, all four data columns are available. This is a local storage profile with its own layout and fixture.

Anchor the codec tests with the literal expected columns for the four-byte ASCII payload `ABCD`. Data bytes are `41 42 43 44`, P is `04` and Q is `ce` in hexadecimal. Obtain reconstruction expectations directly from the original fixture bytes. For `ABCDE`, expect two-byte columns `AB`, `CD`, `E\0` and `\0\0`, with exactly five bytes after reconstruction. Test both fixtures for every pair of absent roles. Include a multistripe fixture plus a padded tail for shard placement, length and range checks.

### Publication, reads and repair

Use immutable manifests with schema version, storage profile, key, version ID, prior version ID, idempotency key, creation time, object size, object SHA-256, per-stripe lengths, and per-shard node role, length and SHA-256. Hash original unpadded bytes for the object checksum. Each node stores shards under version-specific internal paths and reports durable acknowledgement only after writing a same-directory temporary file, syncing its contents, renaming it and syncing the containing directory. Persist any newly created directory hierarchy before acknowledging durability. Test this sequence on the supported local Linux filesystem. [Go os package, File.Sync](https://pkg.go.dev/os#File.Sync), [Go os package, Rename](https://pkg.go.dev/os#Rename)

Require six durable node acknowledgments before publishing. Persist the immutable manifest, then atomically advance the key's head pointer using the same sync and rename discipline. Report success after the head and its directory are synced. The head is the commit authority, so abandoned staging and unpublished manifests remain invisible. If a head is unreadable, surface an explicit error for that key. A failure before publication preserves the old head. A connection lost around successful publication leaves an ambiguous client outcome. A retry finds a committed manifest with the same key, idempotency key and body digest and returns that version. A reused idempotency key with different bytes returns 409.

A read treats an unavailable, length-mismatched or checksum-mismatched shard as an erasure. Each stripe requires four verified columns. Reconstruct when needed, validate reconstructed shard hashes and final object SHA-256, then send a success response or slice the verified bytes for range output. With insufficient valid columns, surface 503 before sending successful response headers. With the object size capped, buffering a verified object simplifies integrity and range handling. A four-byte fixture read with `bytes=1-2` returns `BC` and `Content-Range: bytes 1-2/4`.

Repair uses verified sources for the pinned version and generates only missing or corrupt target shards. A repair writes reconstructed bytes through the node's durable store path only when their hashes match the committed manifest. Serialize concurrent repair writes to a version and role. Keep manifest and head unchanged. Report per-role completion and allow idempotent retries. Exercise fault switches and abrupt termination through owned processes and temporary node stores. Node replacement tests can point a cluster role at a fresh empty directory to simulate loss.

## Epic · Demonstrate durable objects and recovery on a local six-node cluster

**Exit demonstration.** Upload one synthetic fixture in the browser. Inspect the committed manifest and save the original checksum. Stop two chosen node processes, download and verify the bytes, restart those roles with fresh empty stores, repair the pinned version, and download again. Confirm the byte hash and version ID are unchanged. Run the complete node-pair matrix and restart checks locally.

### 62-01 · Build a real browser put, get and inspect flow · 3–5 hours

Create the local coordinator, request validation, browser page and store interface. Use an isolated single-node development adapter with a distinct format and directory. Display the active adapter in the UI. Acceptance checks upload and download actual fixture bytes, compare SHA-256, inspect metadata and refresh after restart. Invalid keys, empty files and oversized uploads have visible errors. Use a manual browser check plus Go HTTP tests against a temporary root.

### 62-02 · Implement and verify the 4+2 stripe codec · 4–6 hours

Implement pure fixed-profile encoding and recovery behind a stable stripe interface. Acceptance checks the literal `ABCD` parity bytes, every erasure pair, padded tails, shard length mismatches and insufficient inputs. Reconstruct original bytes without using the encoder output as the only oracle. The codec sequence and fixture constants are documented.

### 62-03 · Publish durable immutable versions · 4–6 hours

Connect the six node processes, stage shards, implement version allocation and publish the manifest and head only after six durable acknowledgments. Acceptance covers first write, overwrite, version pinning, node failure before publication, identical retry, conflicting retry and restart on both sides of publication. Introduce test-only fault points scoped to temporary storage.

### 62-04 · Serve integrity-checked degraded and bounded reads · 4–6 hours

Implement hash verification, reconstruction, closed byte ranges and capacity checks. Acceptance compares downloads with independent fixture bytes for all 15 failed-node pairs, including mixed missing and corrupted columns. Include short, padded and multistripe fixtures, over-budget writes, request timeouts, and a three-node loss resulting in 503. Vary node availability before shard selection and during an unfinished read, using deterministic test gates.

### 62-05 · Repair versions and explain health in the browser · 3–5 hours

Add explicit repair and per-role status. Acceptance starts two roles on empty scratch stores, repairs an existing version, checks the two stored shard digests and verifies current and pinned version reads. A failed or interrupted repair can be rerun. Repeat a successful repair with no new head, no new version and no change in bytes.

### 62-06 · Run the recovery matrix and document limitations · 3–5 hours

Create one reproducible local verification command with isolated temporary stores. Acceptance covers clean boot, abrupt coordinator restart around head publication, six worker acknowledgments, the all-pairs loss matrix, repair and quota rejection, plus the final manual browser walkthrough. Record reported results with commands and environments. State the local-process and local-filesystem limits of the demonstration.

## Sequencing and estimate

Start 62-01 and 62-02 in parallel once the store and codec interfaces are agreed. Merge their work before 62-03 integrates the cluster. After 62-03, run the 62-04 read work alongside 62-05 browser status work against the agreed manifest contract. Connect and verify actual repair after degraded reads work. Finish 62-06 after the features land. The six tickets total **21–33 engineering hours** with a conservative lower bound of 21 hours. Parallel work can shorten calendar time.
