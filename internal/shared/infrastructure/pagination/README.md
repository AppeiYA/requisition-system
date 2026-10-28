# Pagination

This module centralizes pagination defaults and parsing.

## Usage
- In HTTP handlers: use `pagination.Parse(c *gin.Context)` which reads `limit` and `offset`, validates them, writes errors via `response.Error`, and returns `(Query, bool)`.
- In use cases/services: use `shareddomain.Clamp(limit, offset int)` or `shareddomain.NewPagination(...)` to normalize values defensively (transport-independent). The infrastructure package is for HTTP parsing.
- Constants: `DefaultLimit` (100), `MaxLimit` (100). These can be adjusted centrally.

## Notes
- HTTP adapter is responsible for request validation and returning 400s to clients.
- Use cases should remain defensive and clamp values to avoid unexpected behavior across adapters.
