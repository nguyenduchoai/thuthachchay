# 0001 — Record architecture decisions

- **Status**: Accepted
- **Date**: 2026-05-22

## Context

Cần một nơi ghi lại quyết định kiến trúc kèm bối cảnh và hệ quả, để team mới
hiểu được "vì sao chọn cái này" mà không phải hỏi.

## Decision

Dùng ADR (Architecture Decision Record) lưu ở `docs/adr/NNNN-tiêu-đề.md`.
Mỗi ADR có: Status (Proposed/Accepted/Deprecated/Superseded), Context,
Decision, Consequences. Khi superseded, link sang ADR mới.

## Consequences

- Mọi quyết định "không hiển nhiên" phải có ADR trước khi merge.
- ADR là **immutable**: không sửa nội dung sau khi Accepted; thay vào đó tạo
  ADR mới supersede.
- Index ở `docs/adr/README.md`.
