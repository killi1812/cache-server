# go cache server for nix

This project is a reimplementation of
[cache-server](https://github.com/xkriza08/cache-server) with minor improvement
It is a replacement for [cachix server](https://www.cachix.org/)
compatible with cachix client

## Implementation track

- [x] listen
- [x] stop
- [x] agent
  - [x] add
  - [x] remove
  - [x] list
  - [x] info
- [x] cache
  - [x] create
  - [x] start
  - [x] stop
  - [x] delete
  - [x] update
  - [x] list
  - [x] info
- [x] store-path
  - [x] list
  - [x] delete
  - [x] info
- [x] workspace
  - [x] create
  - [x] delete
  - [x] list
  - [x] info
  - [x] cache

- [x] endpoints
  - [x] cache
  - [x] narinfo
  - [x] multipar-nar
  - [x] deployment
  - [x] active
  - [x] nix-cache-nfo
  - [x] nar
- [x] websockets

- [x] Garbage collector that removes old caches
- [x] Redo tokens on agent and workspace


## BUGs

* there is a wrong endpoint hit

> [GIN] 2026/05/08 - 17:23:58 | 404 |       1.181µs |             ::1 | GET      "/nix-cache-info"
> [GIN] 2026/05/08 - 17:23:58 | 404 |       1.178µs |             ::1 | PUT      "/nix-cache-info"

Why is it hitting the enpint 12345 when it needs to hit cache endpoint
