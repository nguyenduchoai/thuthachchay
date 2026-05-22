# ADR — Architecture Decision Records

| # | Tiêu đề | Status |
|---|---|---|
| [0001](./0001-record-architecture-decisions.md) | Record architecture decisions | Accepted |
| [0002](./0002-points-not-cash.md) | Reward = điểm + voucher, không tiền mặt P2P | Accepted |
| [0003](./0003-steps-source-strava-primary.md) | Strava là nguồn bước chân chính, ZMP phụ | Accepted |

## Cách viết ADR mới

```bash
cp docs/adr/0001-record-architecture-decisions.md docs/adr/$(printf '%04d' $((LAST+1)))-tieu-de.md
```

Mỗi ADR có 3 section: Context, Decision, Consequences. Status: Proposed → Accepted → Deprecated → Superseded by NNNN.
