# go cache server for nix

This project is a reimplementation of
[cache-server](https://github.com/xkriza08/cache-server) with minor improvement
It is a replacement for [cachix server](https://www.cachix.org/)
compatible with cachix client

## Implementation track

- [x] listen
- [x] stop
- [ ] agent
  - [x] add
  - [ ] remove
  - [x] list
  - [x] info
- [ ] cache
  - [x] create
  - [x] start
  - [x] stop
  - [x] delete
  - [ ] update
  - [x] list
  - [x] info
- [ ] store-path
  - [ ] list
  - [ ] delete
  - [ ] info
- [x] workspace
  - [x] create
  - [x] delete
  - [x] list
  - [x] info
  - [x] cache

- [ ] endpoints
  - [ ] cache
  - [ ] narinfo
  - [ ] multipar-nar
  - [ ] deployment
  - [ ] active
  - [ ] nix-cache-nfo
  - [ ] nar
- [ ] websockets

- [ ] Garbage collector that removes old caches
- [ ] Redo tokens on agent and workspace

## Compliance Implementation Plan

To achieve full compatibility with the `cachix` client and the standard Nix
binary cache protocol, the following steps must be implemented:

### Step 1: Fix Metadata Formats [COMPLETED]
- **Standardize `GET /api/v1/cache/$name`:** Update the response to return a Cachix-compatible JSON object including `publicSigningKeys`, `isPublic` (boolean), and `preferredCompressionMethod`.
- **Implement Nix-compatible `nix-cache-info`:** Change the response from JSON to plain text (e.g., `StoreDir: /nix/store\nWantMassQuery: 1\nPriority: 30\n`).
- **Implement `.narinfo` text format:** Update `GET /:storeHash.narinfo` to return the standard signed Nix text format instead of a JSON DB model. Use the existing `GenerateNarInfo` method in `StorePathSrv`.

### Step 2: Implement Binary Data Serving [COMPLETED]
- **NAR Download Endpoint:** Create a route (e.g., `/nar/:fileHash.nar.xz`) that streams the actual binary package data from `ObjectStorage` to the client.
- **Content-Type handling:** Ensure binary downloads use `application/x-nix-archive` or `application/octet-stream`.

### Step 3: Complete Multipart NAR Protocol [IN PROGRESS]
- **Finalization Logic:** Implement the `.../complete` and `.../abort` handlers for multipart uploads. [Basic handlers added]
- **Persistence:** Ensure that once a multipart upload is completed, the corresponding `StorePath` record is updated/finalized in the database.

### Step 4: Implement Deployment Activation [IN PROGRESS]
- **Endpoint:** Connect `POST /api/v1/deploy/activate` to a new handler in `deployApi`. [Basic handler added]
- **Logic:** Parse the deployment specification and trigger notifications/tasks for the targeted agents.

### Step 5: Signing Key Management

- **Key Generation:** Add support for generating Ed25519 signing keys for each
binary cache.
- **Signature Integration:** Ensure `GenerateNarInfo` uses the cache-specific
private key to sign the output.

### Step 6: Protocol Compliance Verification

- **Test Suite:** Run the `api-test.sh` script and ensure every step (Building,
Pushing, Info retrieval, Deployment) returns "RESULT: Success".

