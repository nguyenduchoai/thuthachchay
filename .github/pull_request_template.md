## Tóm tắt

<!-- 1-3 câu mô tả thay đổi. -->

## Vì sao

<!-- Liên kết issue / context. Vì sao thay đổi này cần thiết. -->

## Phạm vi ảnh hưởng

- [ ] miniapp
- [ ] api
- [ ] worker
- [ ] admin
- [ ] db migration
- [ ] infra/ci

## Test plan

- [ ] Unit test pass (`make test`)
- [ ] Lint pass (`make lint`)
- [ ] Đã chạy thử local (nếu UI)
- [ ] OpenAPI/SDK update nếu thay đổi schema

## Migration / Breaking

- [ ] Có migration DB (kèm down)
- [ ] Có breaking change (đã ghi CHANGELOG + bump version)

## Bảo mật

- [ ] Không commit secret
- [ ] Validate input ở biên ngoài
- [ ] Audit log entry nếu là admin action
