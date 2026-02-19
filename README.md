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
  - [ ] info
- [ ] cache
  - [x] create
  - [ ] start
  - [ ] stop
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
