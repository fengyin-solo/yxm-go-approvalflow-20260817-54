# Bug Reproduction

## Bug 是什么

审批模板、申请人、审批单和审批记录的外部输入没有做一致的首尾空白规范化，导致合法的带空格输入被误判、保存值不一致，部分空白字段也无法按预期校验。

## 如何触发

在项目根目录运行：

```bash
go test ./internal/model -run TestDomainValidationNormalizesExternalInput -count=1
```

## 错误信息

```text
template with padded fields should validate: category: 业务类别不合法
```
