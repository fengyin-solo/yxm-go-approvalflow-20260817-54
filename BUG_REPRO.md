# Bug Reproduction

## Bug 是什么

审批请求和模板状态的防御性校验被放松，空标题、空事由、无节点模板激活、停用申请人提交以及重复流水号都没有被正确拒绝。

## 如何触发

在项目根目录运行：

```bash
go test ./internal/service -run TestRequestValidationAndStateGuardsRemainStrict -count=1
```

## 错误信息

```text
template without nodes must not become active
```
