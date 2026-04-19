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

- [ ] Garbage collector that removes old caches
- [ ] Redo tokens on agent and workspace
