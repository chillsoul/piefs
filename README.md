# Pie File System

A simple file system based on [Facebook Haystack Paper](https://www.usenix.org/legacy/event/osdi10/tech/full_papers/Beaver.pdf).

Only for learning, **NOT RECOMMEND** to use for production environment (see [SeaweedFS](https://github.com/chrislusf/seaweedfs) instead).

---
### Planning
Planning to refactor architecture. May use gRPC or go-micro, etc., to make this system more simple, Especially HTTP part.

### TODO List(before)
- Master
    - [ ] Web UI
    - [x] Heartbeat Monitor
    - HTTP API
      - [x] Get needle physical URL
      - [x] Upload and Hand off needle
      - [ ] Delete needle
- Storage
  - [ ] Cache
  - [x] Directory
    - [x] LevelDB store file index
  - [x] Volume
  - [x] Needle
  - [x] Heartbeat
  - HTTP API
    - [x] Add volume
    - [x] Get needle
    - [x] Put needle
    - [x] Delete needle
---

## Document

### Directory

A directory use Key-Value database (LevelDB now) to store the mapping relationship between volume id,needle id and needle metadata (store <<vid,nid>,n metadata> in short). 

### Volume

Each Volume file's first 8 bytes is its current offset, which means storage server can store data from here.

### Needle

```go
// piefs/storage/needle/needle.go
type Needle struct {
	ID           uint64    //unique ID 64bits; stored
	Size         uint64    //size of body 64bits; stored
	Offset       uint64    //offset of body 64bits; stored
	FileExt      string    //file extension; stored
	UploadTime   time.Time //upload time; stored
	File         *os.File  //volume file; memory only
	currentIndex uint64    //current index for IO read and write
}
```

The fields which be commented with 'stored' means it's a needle's metadata, and these will be stored in physical volume file before needle data as a header.

The `currentIndex` is used for implementing `Reader` and `Writer` interfaces.

## Reference

This repository references many great project or paper (including but not limited to code and design ideas), especially following:

[Facebook Haystack Paper](https://www.usenix.org/legacy/event/osdi10/tech/full_papers/Beaver.pdf)

[AlexanderChiuluvB/xiaoyaoFS - Github](https://github.com/AlexanderChiuluvB/xiaoyaoFS)

[chrislusf/seaweedfs - Github](https://github.com/chrislusf/seaweedfs)

[hmli/simplefs - Github](https://github.com/hmli/simplefs)

[030io/whalefs - Github](https://github.com/030io/whalefs)

Really thanks!