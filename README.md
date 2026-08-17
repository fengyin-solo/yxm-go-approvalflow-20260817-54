# Approval Flow — 审批流中心

纯 Go 标准库实现的审批流后端服务，零第三方依赖，开箱即跑。

## 业务说明

管理企业审批的完整生命周期：**模板编排 → 激活 → 发起审批 → 逐级审批 → 通过/驳回/撤回 → 统计报表**。

- **审批模板**：定义一类审批（请假/报销/采购/通用）的节点编排，状态机 `draft → active → archived`。
- **申请人**：员工主体，发起审批单；可被配置为审批节点的审批人。
- **审批节点**：模板下的顺序节点（Seq 1..10），指定审批人，按序逐级审批。
- **审批单**：申请人基于模板发起，状态机 `pending → approved / rejected / canceled`。
- **审批记录**：每次同意/驳回/撤回的不可变留痕，支持按单查询时间线。

> 涉及金额字段（Amount）单位为人民币「分」（int64），0 表示不涉及金额。

## 运行

```bash
cd origin
go run ./cmd/server
# 默认监听 :8080，可通过 PORT / ADDR 环境变量修改
```

环境变量：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| PORT | 8080 | 监听端口 |
| ADDR | :PORT | 完整监听地址（优先于 PORT） |
| MAX_PAGE_SIZE | 100 | 分页最大条数 |
| LOG_LEVEL | info | 日志级别：debug/info/warn/error |

## API 一览

### 审批模板

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/templates | 创建模板（草稿态） |
| GET | /api/templates | 列表（支持 status/category/keyword 筛选 + 分页） |
| GET | /api/templates/{id} | 详情 |
| PUT | /api/templates/{id} | 更新可编辑字段 |
| DELETE | /api/templates/{id} | 删除（已有审批单则拒绝，级联删节点） |
| POST | /api/templates/{id}/transition | 状态流转（激活前须有节点） |

### 审批节点

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/templates/{id}/nodes | 为草稿模板追加节点 |
| GET | /api/templates/{id}/nodes | 查询模板节点（按 Seq 升序） |
| DELETE | /api/nodes/{id} | 删除草稿模板的节点 |

### 申请人

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/applicants | 创建申请人 |
| GET | /api/applicants | 列表（支持 department/status/keyword 筛选 + 分页） |
| GET | /api/applicants/{id} | 详情 |
| PUT | /api/applicants/{id} | 更新 |
| DELETE | /api/applicants/{id} | 删除（有审批单或作为审批人则拒绝） |
| GET | /api/applicants/{id}/pending | 查询该审批人的待办审批单 |

### 审批单

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/requests | 发起审批 |
| GET | /api/requests | 列表（支持 template_id/applicant_id/status/keyword 筛选 + 分页） |
| GET | /api/requests/{id} | 详情 |
| GET | /api/requests/{id}/timeline | 审批时间线 |
| POST | /api/requests/{id}/approve | 当前节点同意（末节点则通过） |
| POST | /api/requests/{id}/reject | 当前节点驳回（必须填意见，直接终态） |
| POST | /api/requests/{id}/cancel | 申请人撤回 |

### 审批记录

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/records | 列表（支持 request_id/operator_id/action 筛选 + 分页） |
| GET | /api/records/{id} | 详情 |

### 统计

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/stats/overview | 全局概览（单量/各状态计数/金额/通过率） |
| GET | /api/stats/by-template | 按模板分组统计 |
| GET | /api/stats/by-department | 按申请人部门分组统计 |
| GET | /api/stats/top-approvers | 审批人工作量排行（?n=10） |

### 健康检查

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /healthz | 健康检查 |

## 统一响应格式

```json
{"code": 0, "message": "ok", "data": ...}
```

错误码映射：400 参数校验失败 / 404 记录不存在 / 409 状态冲突或唯一性冲突 / 500 内部错误。

## 工程结构

```
origin/
├── go.mod
├── README.md
├── cmd/server/main.go          # 入口：配置加载、依赖装配、优雅关闭
├── internal/
│   ├── app/app.go              # 依赖装配 store -> service -> handler
│   ├── config/config.go        # 环境变量配置
│   ├── model/                  # 领域模型 + 状态机 + 校验
│   ├── store/                  # Store 接口 + 内存实现
│   ├── service/                # 业务逻辑 + 统计
│   └── handler/                # HTTP 路由 + 处理器
└── pkg/
    ├── httpx/httpx.go          # 统一响应、分页、JSON 解析
    ├── idgen/idgen.go          # Hex ID + base62 短码
    └── logger/logger.go        # 分级日志
```

## 测试

```bash
go test ./...
```
