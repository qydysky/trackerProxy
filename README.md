# trackerProxy
反向代理tracker并使用代理

使用原因：
- 某些节点无法访问
- qb中设置代理会影响DHT节点搜索
- 目前使用广泛的反代实现中，普遍没有支持再经http代理

bt client <-> trackerProxy <-> httpProxy <-> tracker

eg http://t.acg.rip:6699/announce -> http://192.168.31.101:10002/http_t.acg.rip_6699