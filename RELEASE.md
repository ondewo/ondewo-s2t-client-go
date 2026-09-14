# Release History

*****************

## Release ONDEWO S2T Go Client 7.5.0

### New Features

* Initial release of the ONDEWO S2T (Speech-to-Text) gRPC client for Go. The module
  ships the stubs generated from the [ONDEWO S2T API](https://github.com/ondewo/ondewo-s2t-api)
  by version 5.15.1 of the
  [ONDEWO Proto Compiler](https://github.com/ondewo/ondewo-proto-compiler): one `*.pb.go` of
  messages and one `*_grpc.pb.go` of service stubs per `.proto`, below `api/ondewo/s2t/`,
  compiled against the `google.golang.org/protobuf` and `google.golang.org/grpc` runtimes pinned by
  the compiler image.
* `make build` reproduces the whole client from the two submodules — proto compiler image, stub
  generation and `go build` — and `make check_build` asserts that every `.proto` of the API
  produced a stub.
* A go test suite under `tests/`, run by `.github/workflows/ci.yml` on go 1.25 and 1.27 against the
  committed stubs. It needs no network and no ONDEWO server: gRPC calls travel over an in-memory
  `bufconn` listener, so the generated client stub, the generated server stub and a real HTTP/2
  connection are all exercised in-process. Messages round-trip through the wire format, every
  generated unary stub is called, and the two generators' views of the service (`grpc.ServiceDesc`
  and the proto descriptor) are cross-checked against each other.

### Improvements

* Version numbering follows the fleet rule: the client version matches the ONDEWO S2T API in
  major and minor version, so the first release is `7.5.0` rather than a `0.x`. The heading
  above and `ONDEWO_S2T_VERSION` in the `Makefile` now agree on it — they did not before, and
  `make build_gh_release` slices its release body out of this file by exactly that string, so the
  mismatch would have published a GitHub release with an empty body.
* `make test_coverage` gates the build on the coverage of the HAND-WRITTEN code (`auth/`, currently
  100%) and fails when it ends up measuring nothing at all; `make test_coverage_generated` reports
  the generated stubs' figure (17.1%) without ever gating on it. `make check_stubs` runs
  first in CI and fails when `api/` is absent or empty, so the pipeline cannot go green by finding
  no work to do.
* The release targets tag each release twice on the same commit: with the ONDEWO release number
  (`7.5.0`) that the rest of the fleet uses, and with the `v`-prefixed spelling (`v7.5.0`) that is
  the only tag shape the Go module resolver accepts. `make publish_go_module` then warms
  `proxy.golang.org` so the new version is immediately installable with `go get`.
