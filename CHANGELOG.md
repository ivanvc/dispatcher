# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] - TBA

## Fixed
- Allow the controller manager to write events
- Improve error messages when status fails to update

## [0.4.0] - 2023-07-27
## Added
- Allow to override the default namespace for jobs triggered via the HTTP API
  endpoint
- Export Prometheus metrics from job requests and executions
- Set the Kubernetes label "app.kubernetes.io/name" and
  "app.kubernetes.io/instance" to all Kubernetes resources
- Record JobExecution events in the controller manager

## Changed
- Update the directory structure of the Kustomization configuration files. Now
  the overlays are located in the `overlays/` directory. There are three
  different overlays:
  - base: doesn't create a namespace, which plays nice with Helm
  - default: uses base, and creates the namespace on top of it
  - with-prometheus-monitoring: uses base, and enables prometheus monitoring
- Replace JobExecution Status' Phase with Conditions
- Bump CRD version to v1beta1

## [0.3.2] - 2023-03-21
## Fixed
- When executing template, only recurse if the value for the pointer is of type
  struct
- Only execute template if the field is exported

## [0.3.1] - 2023-03-16
## Fixed
- Set docker image tag to 0.3.1 in kustomization files

## [0.3.0] - 2023-03-09
## Added
- Use GitHub actions to do the CI/CD of the project

## Fixed
- Using JSON as the medium to generate the Kubernetes Job payload had unexpected
  issues. It caused expression errors with Sprig functions. So, now the
  templating happens at the struct level, executing the template straight in
  each field value

## [0.2.0] - 2023-03-03
### Added
- Allow Sprig functions to be called from `JobTemplate`s

## [0.1.0] - 2023-02-15
### Added
- Add failed phase to `JobExecution`, to track when a job has failed and avoid
  recreating it in the reconciliation loop
- Return a 404 when trying to execute a job from a `JobTemplate` that doesn't
  exist
- Apply `JobTemplate` labels when creating a `JobExecution` via the HTTP API
  endpoint

## Changed
- Adopt SemVer for versioning

### Fixed
- Remove `generateName` from `JobTemplate` sample

## [0.0.2] - 2022-11-22
### Added
- Add namespaceless Kubernetes resources, which can be used by helmfile
- Add sample payloads

### Changed
- Rename `JobTemplate`'s `template` to `jobTemplate`
- Delete `JobExecution` after the `Job` gets deleted

## [0.0.1] - 2022-10-24
### Added
- Initial release

[Unreleased]: https://github.com/ivanvc/dispatcher/compare/0.4.0...HEAD
[0.4.0]: https://github.com/ivanvc/dispatcher/compare/0.3.2...0.4.0
[0.3.2]: https://github.com/ivanvc/dispatcher/compare/0.3.1...0.3.2
[0.3.1]: https://github.com/ivanvc/dispatcher/compare/0.3.0...0.3.1
[0.3.0]: https://github.com/ivanvc/dispatcher/compare/0.2.0...0.3.0
[0.2.0]: https://github.com/ivanvc/dispatcher/compare/0.1.0...0.2.0
[0.1.0]: https://github.com/ivanvc/dispatcher/compare/0.0.2...0.1.0
[0.0.2]: https://github.com/ivanvc/dispatcher/compare/0.0.1...0.0.2
[0.0.1]: https://github.com/ivanvc/dispatcher/releases/tag/0.0.1
