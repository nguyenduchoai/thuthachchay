# Deploy

## Môi trường

| Env | Domain | DB | Redis | Storage |
|---|---|---|---|---|
| dev | localhost | Postgres docker | Redis docker | MinIO docker |
| staging | api-stg.buocvang.vn | RDS small | ElastiCache nano | S3 stg bucket |
| prod | api.buocvang.vn | RDS m6g + replica | ElastiCache m6g | S3 prod bucket |

## Pipeline

```
PR merge → main → CI green
  ├── deploy-staging.yml (auto)
  └── deploy-prod.yml (manual / on tag)
```

## Build container

```bash
docker build -f deploy/Dockerfile.api    -t buocvang/api:$(git rev-parse --short HEAD)    .
docker build -f deploy/Dockerfile.worker -t buocvang/worker:$(git rev-parse --short HEAD) .
docker build -f deploy/Dockerfile.admin  -t buocvang/admin:$(git rev-parse --short HEAD)  .
docker build -f deploy/Dockerfile.miniapp -t buocvang/miniapp:$(git rev-parse --short HEAD) .
```

## Migration trên prod

Migration được áp **trước** khi rollout binary mới. Order:

1. CI build container.
2. Job `migrate` chạy `goose up` → block nếu fail.
3. Rolling deploy `api` (2/3 readiness gate qua `GET /readyz`).
4. Rolling deploy `worker`.
5. Rolling deploy `admin`.

Mọi migration phải reversible. Không xoá cột/bảng có data — luôn 2 bước:
deprecate (giữ + ngừng đọc) → xoá sau ≥ 1 tuần.

## Rollback

```bash
# Image cũ vẫn pin trong registry — chỉ cần redeploy:
kubectl set image deploy/api api=buocvang/api:<prev-sha>
# Nếu migration đã forward và không reversible → coordinate DBA + downtime window.
```

## ZMP miniapp

Miniapp **không** deploy qua container như backend; quy trình:

1. `make miniapp` build dist/.
2. `zmp-cli login` → `zmp-cli deploy --env=staging|production`.
3. Submit lên Zalo Dev Console → chờ review (2–5 ngày).
