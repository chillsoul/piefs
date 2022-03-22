# Pie File System

A simple file system based on [Facebook Haystack Paper](https://www.usenix.org/legacy/event/osdi10/tech/full_papers/Beaver.pdf).

Only for learning, **NOT RECOMMEND** to use for production environment (see [SeaweedFS](https://github.com/chrislusf/seaweedfs) instead).

---

### TODO List
- Master
    - [ ] Web UI
    - HTTP API
      - [x] Put Needle
      - [x] Get Needle
    - gRPC Service
      - [x] Heartbeat
      - [x] Delete Needle
- Storage
  - [ ] Cache
  - [x] Directory
    - [x] LevelDB store file index
  - [x] Volume
  - [x] Needle
  - [x] Heartbeat
  - HTTP API
    - [x] Get Needle
  - gRPC Service
    - [x] Add Volume
    - [x] Put Needle
    - [x] Delete Needle
---

## Document
### Outline design
```mermaid
sequenceDiagram
autonumber
    note over Client,Master:gRPC-gateway
    note over Master,Storage:gRPC
    Client->>+Master: Put/Del Needle HTTP Request
    Master->>+Storage: RPC Request
    Storage-->>-Master: RPC Response
    Master-->>-Client: HTTP Response
```
```mermaid
sequenceDiagram
autonumber
    Client->>+Master: Get Needle HTTP Request
    Master-->>-Client:Redirect to someone Storage
    note over Master,Client: Redirect can be omitted if Storage URL is known
    Client->>+Storage: Get Neelde HTTP Request
    Storage-->>-Client: Response
```
### Directory

```mermaid
classDiagram
class Directory{
<<interface>>
no fields
+Get()
+Set()
+Has()
+Del()
+GetVolumeMap()
}
```

A directory use Key-Value database (LevelDB now) to store the mapping relationship of one volume between needle id and needle metadata (map[vid]LevelDB<nid,n metadata> in short). 

### Volume
```mermaid
classDiagram
class Volume{
<<physical file struct>>
Current Index
Needle 1(offset 8,size x)
Needle 2(offset 8+x,size y)
.......(offset m,size n)
(End of file)
}
```
Each Volume file's first 8 bytes is its current offset, which means storage server can store data from here.

### Needle

```mermaid
classDiagram
class Needle{
	+ID : uint64
    +Size : uint64
    +Offset : uint64
    +FileExt : string
    +UploadTime : time.Time
    -currentIndex : uint64
    +File : *os.File
}
```
The `currentIndex` is used for implementing `Reader` and `Writer` interfaces.
### Usage
```shell
#Upload File By Master
> curl -F 'file=@./resources/test/gofactory.jpg' 'localhost:8080/PutNeedle'
{"vid":1647868700846810400,"nid":1647920354875133400}
```
The master will auto hand off this file to several storage servers for replication.
```shell
#Get File By Master
GET localhost:8080/GetNeedle?vid=1647868700846810400&nid=1647920354875133400
```
Then the master will auto redirect you to physical file URL.
```shell
#Delete File By Master
> curl -d 'vid=1647868700846810400&nid=1647920354875133400' 'localhost:8080/DelNeedle'
success
```
Only delete this record from directory db.It will be physically delete when compaction operation start.

## Reference

This repository references many great project or paper (including but not limited to code and design ideas), especially following:

[Facebook Haystack Paper](https://www.usenix.org/legacy/event/osdi10/tech/full_papers/Beaver.pdf)

[AlexanderChiuluvB/xiaoyaoFS - Github](https://github.com/AlexanderChiuluvB/xiaoyaoFS)

[chrislusf/seaweedfs - Github](https://github.com/chrislusf/seaweedfs)

[hmli/simplefs - Github](https://github.com/hmli/simplefs)

[030io/whalefs - Github](https://github.com/030io/whalefs)

Really thanks!