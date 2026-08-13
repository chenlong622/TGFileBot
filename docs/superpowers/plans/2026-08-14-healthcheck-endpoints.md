# Healthcheck Endpoints Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add unauthenticated `/livez` (liveness, with `/healthz` as deprecated compatibility alias) and `/readyz` (readiness) endpoints for k8s probes.

**Architecture:** Two new `case` branches in the existing `handleMain` router in `http.go`. `/livez` and its alias `/healthz` always return 200; `/readyz` returns 200 when UserBot is logged in (`infos.Status.Load() == 3`), otherwise 503. Both return a small JSON body and require no auth.

**Tech Stack:** Go standard library (`net/http`, `encoding/json`). No new dependencies.

## Global Constraints

- No auth on probe endpoints (k8s probes cannot carry `password`/`hash`).
- Do not modify the existing `/` status endpoint.
- Use `infos.Status.Load()` (value 3 = logged in) for readiness.
- Match existing code style in `http.go` (JSON via `json.NewEncoder`, `log.Printf` on error, Chinese comments).

---

### Task 1: Add `/livez` (with `/healthz` alias) and `/readyz` routes

**Files:**
- Modify: `http.go:30-76` (the `switch` in `handleMain`)

**Interfaces:**
- Consumes: `infos *Infos` global; `infos.Status` is `atomic.Int32` (value 3 = logged in).
- Produces: `GET /livez` (and `/healthz`) → 200 `{"status":"ok"}`; `GET /readyz` → 200 `{"status":"ready"}` when logged in, else 503 `{"status":"not_ready"}`.

- [ ] **Step 1: Add the route cases**

Insert two new `case` branches before the `/pic` case in `handleMain` (`http.go`, after the `/` case at line 46):

```go
	case path == "/livez", path == "/healthz":
		// k8s liveness 探针: 进程存活即返回 200, 无需鉴权
		// (healthz 已自 Kubernetes v1.16 弃用, 保留作兼容别名)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
			log.Printf("发送网页失败: %+v", err)
		}
		return
	case path == "/readyz":
		// k8s readiness 探针: UserBot 已登录(状态 3)才算就绪, 无需鉴权
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if infos.Status.Load() == 3 {
			if err := json.NewEncoder(w).Encode(map[string]string{"status": "ready"}); err != nil {
				log.Printf("发送网页失败: %+v", err)
			}
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "not_ready"}); err != nil {
			log.Printf("发送网页失败: %+v", err)
		}
		return
```

- [ ] **Step 2: Build to verify**

Run: `go build ./...`
Expected: no output, exit code 0.

- [ ] **Step 3: Run go vet**

Run: `go vet ./...`
Expected: no output, exit code 0.

- [ ] **Step 4: Manual smoke test**

Start server, then:
- `curl http://localhost:8080/livez` → `200` `{"status":"ok"}`
- `curl http://localhost:8080/healthz` → `200` `{"status":"ok"}` (兼容别名)
- `curl http://localhost:8080/readyz` → `503` `{"status":"not_ready"}` (UserBot not logged in at startup), or `200` `{"status":"ready"}` after login.

- [ ] **Step 5: Commit**

```bash
git add http.go
git commit -m "feat: add /livez and /readyz probe endpoints"
```