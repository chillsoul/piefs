## Pie File System
A simple file system based on [Facebook Haystack Paper](https://www.usenix.org/legacy/event/osdi10/tech/full_papers/Beaver.pdf).

---
###TODO List
- [ ] File Merge Store
- [ ] NoSQL Support
  - [ ] Redis
- [ ] Distributed Storage
- [ ] Linux FUSE
---

###Directory

负责逻辑卷到物理卷的映射

### Store

负责存储

### Cache

未命中时再从磁盘调入