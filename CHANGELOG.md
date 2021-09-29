# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.1] - 2020-09-01
### Added
- The installation script can now receive a parameter to install a specific version instead of always installing the latest [#124]

### Changed
- Verbose, Insecure and Detach flags working when loading from an alias [#127]

### Deleted

## [1.0.0] - 2020-08-13
### Added
- Support for ssh remote port forwarding [#114]
- Support for authentication ssh session using ssh agent [#102]
- Add builds for ARM [#109]

### Changed
- Complete revamp of CLI user experience [#112]

### Deleted

## [0.5.0] - 2019-10-02
### Added
- Configurable connection timeout [#92]
- Keep idle connection open by sending periodic synthetic packets (-keep-alive-interval flag) [#77]

### Changed
- Reconnect to SSH Server if connection drops for any reason (-connection-retries and -retry-wait) [#95]
- SSH config file is required even if all required arguments were provided through CLI [#75]
- Missing port in remote address [#86]
- Fix persistence of insecure mode flag (-insecure) [#90]
- Better protecting keys loaded in memory [#78]

### Deleted

## [0.4.0] - 2019-06-23
### Added
- Multiple tunnels using the same ssh connection (support for multiple -remote flags) [#72]

### Changed
- Project dependencies are now managed by Go modules instead of vendor/ [#69]

### Deleted

## [0.3.0] - 05-11-2019
### Added
- Windows Support! Mole now works on windows (tested on Windows 10) [#65]
- Using Github Actions for code quality checks (e.g. unit tests, code formatting, etc.)
- Skip the host key validation by using the -insecure option [#52]
- Always use the same ssh connection if multiple clients use the same tunnel [#43]
- Run mole in background by using the -detach option [#35]
- New -aliases option added to list all configured aliases [#29]
- LocalForward option from ssh config file will be used if both -local and -remote are absent [#18]
- Developers can spawn a small local infra using docker to test their changes

### Changed
- Users will be prompted to enter the key's password if it is encrypted [#54]
- Server names can contain underscore character [#50]
- Return error if required flags are missing [#33]

### Deleted

## [0.2.0] - 2018-10-14
### Added
- Aliases can be created to reuse tunnel settings.

### Changed

### Deleted

## [0.1.0] - 2018-10-10
### Added
- Add -version option to display the current version
- New website: https://davrodpin.github.io/mole/

### Changed
- IP addresses of both local and remote are now optional

### Deleted

## [0.0.1] - 2018-10-05
### Added
- First release. No changes.

### Changed

### Deleted

