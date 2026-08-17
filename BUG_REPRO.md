# Bug Reproduction

## Bug 是什么

内存存储在申请人、模板、审批单和审批记录的创建与列表读取路径上缺少互斥保护，并发读写同一批 map 时会触发 race detector 报警，严重时出现并发 map 迭代和写入崩溃。

## 如何触发

在项目根目录运行：

```bash
go test -race ./internal/store -run TestConcurrentMemoryStoreAccessIsRaceFree -count=1
```

## 错误信息

```text
WARNING: DATA RACE
fatal error: concurrent map iteration and map write
```
