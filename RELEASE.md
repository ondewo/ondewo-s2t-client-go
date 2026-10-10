# Release History

*****************

## Release ONDEWO S2T Go Client 7.5.1

### New Features

* `client.NewChannel(cfg client.Config, opts ...grpc.DialOption)` opens the gRPC connection for
  plaintext, TLS or mutual TLS, following the TLS contract of the python SDKs
  (`ondewo-client-utils` 4.1.1), so one set of certificates works with every ONDEWO client.
  `client.Config` carries `Host`, `Port`, `Insecure`, an optional `Logger` and the PEM **content**
  (never a file path) of the CA (`GrpcCert`, empty = system roots) and of an optional client
  identity (`GrpcClientCert` / `GrpcClientKey`).
* Refused with an error before gRPC sees them: half a client identity, a certificate and key that
  do not form a pair, a `GrpcCert` that holds no PEM certificate (typically a path), and
  `Insecure: true` combined with a client identity. No error message contains a PEM, a key or the
  whole `Config`; they name the field and `host:port`.
* A plaintext connection logs a warning naming `host:port` through `log/slog` (`Config.Logger`, or
  `slog.Default()`); the package never configures logging. Every `fmt` verb and `slog` rendering of
  a `Config` redacts the client key.
* Bare IPv6 literal hosts are bracketed (`::1` -> `[::1]:<port>`); bracketed hosts and hosts with a
  scheme are used as they are. PEMs with CRLF line endings work.
* Connection defaults of the python SDKs that grpc-go exposes: keepalive pings every 30 s with a
  20 s timeout, only while an RPC is active; 5 s maximum reconnect backoff; 2^31-1 byte maximum
  message size in both directions. Every `grpc.DialOption` passed in is applied after them and wins.
  Documented gaps: grpc-go has no `http2.max_pings_without_data` and only one keepalive timeout, and
  no per-method retry policy is configured (only gRPC's transparent retries apply).

### Improvements

* `tests/tls_test.go` runs real TLS and mutual-TLS handshakes against an in-process server with a
  PKI generated per run; the 100% coverage gate now spans `auth/` and `client/`.
* `tests/release_notes_test.go` pins the `RELEASE.md` slice the GitHub release body is built from:
  the Makefile's perl range, the spelling of every release heading, the `*****` separator closing
  every section, and non-empty notes for the current version.
* README: new section "TLS, mutual TLS and certificates" (modes, connection defaults, a test PKI with
  openssl, TLS security notes, troubleshooting).
* The ONDEWO proto compiler submodule is pinned to 5.15.2 (was 5.15.1).

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

*****************
