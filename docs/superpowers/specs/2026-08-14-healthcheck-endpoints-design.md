# 设计：health-check / readiness endpoints

日期：2026-08-14
状态：已批准

## 目标

为 k8s 部署提供 liveness 与 readiness 探针端点，无需鉴权。

## 端点定义

| 端点 | 类型 | 逻辑 | 响应 |
|------|------|------|------|
| `GET /healthz` | liveness | 进程存活即返回 200 | `200` + JSON `{"status":"ok"}` |
| `GET /readyz` | readiness | UserBot 已登录（`infos.Status == 3`）返回 200，否则 503 | `200`/`503` + JSON `{"status":"ready"/"not_ready"}` |

## 实现

- 在 `handleMain` 的 `switch` 中新增 `case path == "/healthz"` 与 `case path == "/readyz"`，复用现有路由分发。
- 无鉴权（探针不应携带 password/hash）。
- 通过 `infos.Status.Load()` 判断 UserBot 登录状态（值 3 = 已登录）。

## 明确不做

- 不增加 TCP 探活检查（保持简单）。
- 不改动 `/` 现有状态端点。