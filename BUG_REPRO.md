# Bug Reproduction

## Bug 是什么

多级审批节点的模板隔离和顺序契约被破坏，同序号节点在不同模板之间错误冲突，节点列表倒序返回，审批单会从错误节点开始并在流转时找错下一节点。

## 如何触发

在项目根目录运行：

```bash
go test ./internal/service -run TestApprovalNodesKeepTemplateScopedAscendingOrder -count=1
```

## 错误信息

```text
same sequence in another template should be allowed: 记录已存在或状态冲突
```
