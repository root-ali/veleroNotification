# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-03-19

### Added

- Initial release of Velero Reporter.
- Real-time monitoring of Velero Backup resources using Kubernetes informers.
- Mattermost notifications for backup success and failure.
- Configurable Mattermost URL and webhook token via environment variables or helm values.
- Configurable kubernetes configuration (`in-cluster` or `out-of-cluster`).
- Configurable http server host and port.
- Configurable log level.
- ConfigMap management for tracking the last processed resource version.
- Robust error handling for Kubernetes API and Mattermost interactions.
- Health check: check if the kubernetes api is ready.
- deployment via helm chart.

### Changed

- Refactored `NewKubernetesClient` to remove unused context and simplify context management.
- Improved `HealthCheck` for better error handling and logging.
- Corrected the `VeleroBackupWatch` function to use a single informer, preventing rate-limiting issues.
- Improved `updateConfigMap` function to use `Update` instead of `Patch`.
- Added `getBackupStatus` function for parsing Velero `BackupStatus`.
- Corrected the Mattermost message format to match the API requirements.
- Improved error handling throughout the codebase.
- Update README file.

### Added in the future

- add more tests.
- add metrics.
- add multi notification channels(mail,...) support.
