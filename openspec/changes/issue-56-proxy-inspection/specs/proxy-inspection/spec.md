## ADDED Requirements

### Requirement: Per-request JSON log files
The proxy SHALL write one JSON log file per forwarded request to the configured log directory. Each file SHALL contain the full inbound request body and the full outbound response body.

#### Scenario: Successful request logged
- **WHEN** the proxy forwards a request to OpenRouter and receives a response
- **THEN** a JSON file is written to `LOG_DIR` containing `request` (full body) and `response` (full body) fields

#### Scenario: Failed upstream request logged
- **WHEN** the upstream request fails (network error, timeout)
- **THEN** a JSON file is written containing `request` and `error` fields; `response` is omitted

#### Scenario: Non-JSON response stored as string
- **WHEN** the upstream response body is not valid JSON
- **THEN** the `response` field in the log file SHALL contain the raw body as a JSON string value, not cause the log file to be invalid JSON

### Requirement: Log directory configuration
The proxy SHALL read the log directory path from the `LOG_DIR` environment variable. If `LOG_DIR` is not set, the proxy SHALL default to `/logs`. The proxy SHALL create the directory on startup if it does not exist.

#### Scenario: Default log directory used when LOG_DIR unset
- **WHEN** `LOG_DIR` is not set
- **THEN** log files are written to `/logs`

#### Scenario: Custom log directory used when LOG_DIR set
- **WHEN** `LOG_DIR=/data/logs` is set
- **THEN** log files are written to `/data/logs`

### Requirement: Log file naming
Log files SHALL be named by the UTC timestamp of the request in the format `20060102T150405.000000000Z.json`.

#### Scenario: File name is unique per request
- **WHEN** two requests arrive at different nanoseconds
- **THEN** they produce two distinct log files
