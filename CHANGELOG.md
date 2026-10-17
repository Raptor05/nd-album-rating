# Changelog

## 1.1.0

- English texts, logs and documentation
- GitHub workflows for CI builds and releases
- Flat project layout with native unit tests

## 1.0.1

- Removed reliance on the sandbox clock (fixes callback timeouts and a stuck interval)
- Work per callback is bounded by a fixed number of API calls
- Stale runs are detected through a KVStore lock with TTL

## 1.0.0

- Initial version
