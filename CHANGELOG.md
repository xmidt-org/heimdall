# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]
- Patch failing Dockerfile, fix linter issues [#54](https://github.com/xmidt-org/heimdall/pull/54)

## [v0.3.1]
- Migrate to github actions, normalize analysis tools, Dockerfiles and Makefiles. [#27](https://github.com/xmidt-org/heimdall/pull/27)
- Updated references to the main branch [#23](https://github.com/xmidt-org/heimdall/pull/23)
- bumped bascule to remove security vulnerability. [#31](https://github.com/xmidt-org/heimdall/pull/31)
- Updated spec file and rpkg version macro to be able to choose when the 'v' is included in the version. [#32](https://github.com/xmidt-org/heimdall/pull/32)
- Fix security vulns
  - https://github.com/xmidt-org/heimdall/issues/44
  - https://github.com/xmidt-org/heimdall/issues/45
  - https://github.com/xmidt-org/heimdall/issues/46
  - https://github.com/xmidt-org/heimdall/issues/47
  - https://github.com/xmidt-org/heimdall/issues/49


## [v0.3.0]
- improve logging
- improve configuration
- improve pool metrics
- improve filling pool with random devices
- Fix loading of XMiDT SAT [#21](https://github.com/xmidt-org/heimdall/pull/21)

## [v0.2.0]
- bumped bascule version to v0.7.0
- bumped webpa-common to v1.5.1
- leverage capacityset in shuffle package
- switch db from cockroach to yugabyte
- bump codex-db to v0.4.0
- update time window
- Updated release pipeline to use travis

## [v0.1.6]

## [v0.1.5]
- updated urls and imports

## [v0.1.3]
- metrics fix

## [v0.1.2]
- metrics fix

## [v0.1.1]
- setting up pipeline

## [v0.1.0]
- Initial Code added.
- Added metric labels.
- Bumped codex for `db` package updates.

[Unreleased]: https://github.com/xmidt-org/heimdall/compare/0.3.1...HEAD
[v0.3.1]: https://github.com/xmidt-org/heimdall/compare/0.3.0...v0.3.1
[v0.3.0]: https://github.com/xmidt-org/heimdall/compare/0.2.0...v0.3.0
[v0.2.0]: https://github.com/xmidt-org/heimdall/compare/0.1.6...v0.2.0
[v0.1.6]: https://github.com/xmidt-org/heimdall/compare/0.1.5...v0.1.6
[v0.1.5]: https://github.com/xmidt-org/heimdall/compare/0.1.4...v0.1.5
[v0.1.4]: https://github.com/xmidt-org/heimdall/compare/0.1.3...v0.1.4
[v0.1.3]: https://github.com/xmidt-org/heimdall/compare/0.1.2...v0.1.3
[v0.1.2]: https://github.com/xmidt-org/heimdall/compare/0.1.1...v0.1.2
[v0.1.1]: https://github.com/xmidt-org/heimdall/compare/0.1.0...v0.1.1
[v0.1.0]: https://github.com/xmidt-org/heimdall/compare/0.0.0...v0.1.0
