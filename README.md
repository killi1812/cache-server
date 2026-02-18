# go cache server for nix

This project is a reimplementation of [cache-server]() with minor improvement
It is a replacement for [cachix server]() compatible with cachix client

## Implementation track

- [x] listen
- [x] stop
- [ ] agent
  - [ ] add
  - [ ] remove
  - [ ] list
  - [ ] info
- [ ] cache
  - [x] create
  - [ ] start
  - [ ] stop
  - [x] delete
  - [ ] update
  - [ ] list
  - [x] info
- [ ] store-path
  - [ ] list
  - [ ] delete
  - [ ] info
- [ ] workspace
  - [ ] create
  - [ ] delete
  - [ ] list
  - [ ] info
  - [ ] cache

- [ ] endpoints
  - [ ] cache
  - [ ] narinfo
  - [ ] multipar-nar
  - [ ] deployment
  - [ ] active
  - [ ] nix-cache-nfo
  - [ ] nar
- [ ] websockets
