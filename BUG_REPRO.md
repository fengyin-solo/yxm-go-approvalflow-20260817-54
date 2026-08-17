# Bug Reproduction

## Bug 是什么

模板、申请人、审批单的关键词筛选没有稳定使用大小写无关的业务标识字段，同时全局统计遗漏了被驳回审批单的金额，造成列表检索和金额报表偏差。

## 如何触发

在项目根目录运行：

```bash
go test ./internal/service -run TestFiltersAndStatsUseStableBusinessFields -count=1
```

## 错误信息

```text
template keyword should match code case-insensitively, total=0 err=<nil>
```
