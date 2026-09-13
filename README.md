<div align="center">
  <table>
    <tr>
      <td>
        <a href="https://ondewo.com/">
            <img width="400px" src="https://raw.githubusercontent.com/ondewo/ondewo-logos/master/ondewo_we_automate_your_phone_calls.png"/>
        </a>
      </td>
    </tr>
    <tr>
       <td align="center">
          <a href="https://www.linkedin.com/company/ondewo"><img width="40px" src="https://cdn-icons-png.flaticon.com/512/3536/3536505.png"></a>
          <a href="https://www.facebook.com/ondewo"><img width="40px" src="https://cdn-icons-png.flaticon.com/512/733/733547.png"></a>
          <a href="https://twitter.com/ondewo"><img width="40px" src="https://cdn-icons-png.flaticon.com/512/733/733579.png"></a>
          <a href="https://www.instagram.com/ondewo.ai/"><img width="40px" src="https://cdn-icons-png.flaticon.com/512/174/174855.png"></a>
       </td>
    </tr>
  </table>
  <h1 align="center">
    ONDEWO S2T Client Go
  </h1>
</div>

## Overview

`ondewo-s2t-client-go` is the Go gRPC client of the ONDEWO S2T (Speech-to-Text) API. It is a
compiled version of the [ONDEWO S2T API](https://github.com/ondewo/ondewo-s2t-api),
generated with the [ONDEWO PROTO COMPILER](https://github.com/ondewo/ondewo-proto-compiler).

ONDEWO APIs use [Protocol Buffers](https://github.com/protocolbuffers/protobuf) version 3 (proto3)
as their Interface Definition Language (IDL) to define the API interface and the structure of the
payload messages. The same interface definition is used for the gRPC version of the API in all
languages, so the service and message names below are the ones documented for the API itself.

Everything under `api/` is **generated**. It is nevertheless committed to this repository, because a
Go module is resolved straight from its version control tree — there is no build step between
`go get` and the consumer's compiler. Hand-written code therefore lives *outside* `api/`.

## Installation

```shell
go get github.com/ondewo/ondewo-s2t-client-go
```

Then import the package of the service you need:

```go
import s2tpb "github.com/ondewo/ondewo-s2t-client-go/api/ondewo/s2t"
```

> **Major versions.** From major version 2 on, a Go module path carries its major version as a
> `/vN` suffix (see [the module reference](https://go.dev/ref/mod#major-version-suffixes)), so the
> import path of release `7.5.0` is `github.com/ondewo/ondewo-s2t-client-go/v7/api/ondewo/s2t`. The
> `Makefile` derives the suffix from `ONDEWO_S2T_VERSION`; run `make TEST` to print the
> exact module path of the current release.

To work on the client itself:

```shell
git clone https://github.com/ondewo/ondewo-s2t-client-go.git   ## Clone the repository
cd ondewo-s2t-client-go                                        ## Change into the repo directory
make setup_developer_environment_locally              ## Check out submodules, install pre-commit hooks
```

## Usage

```go
package main

import (
    "context"
    "crypto/tls"
    "log"
    "os"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"

    s2tpb "github.com/ondewo/ondewo-s2t-client-go/api/ondewo/s2t"
    "github.com/ondewo/ondewo-s2t-client-go/auth"
)

func main() {
    // auth.WithBearerToken sends the Keycloak access token as `authorization: Bearer <token>` on
    // every call of this connection, exactly as the other ONDEWO clients do. It refuses to attach
    // itself to a plaintext connection, so it is paired with transport credentials here.
    conn, err := grpc.NewClient(
        "grpc-s2t.ondewo.com:443",
        grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
        auth.WithBearerToken(os.Getenv("ONDEWO_S2T_ACCESS_TOKEN")),
    )
    if err != nil {
        log.Fatalf("could not connect: %v", err)
    }
    defer conn.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // The API has a single service, Speech2Text, with its generated NewSpeech2TextClient
    // constructor. Browse api/ondewo/s2t/ for the RPCs it exposes.
    client := s2tpb.NewSpeech2TextClient(conn)

    response, err := client.ListS2TPipelines(ctx, &s2tpb.ListS2TPipelinesRequest{})
    if err != nil {
        log.Fatalf("rpc failed: %v", err)
    }
    log.Printf("pipelines: %v", response.GetPipelineConfigs())
}
```

## Repository structure

```
.
├── api                                    <----- GENERATED - do not edit, `make generate_ondewo_protos` rewrites it
│   └── ondewo
│       └── s2t
│           ├── *.pb.go                    <----- messages (protoc-gen-go)
│           └── *_grpc.pb.go               <----- service stubs (protoc-gen-go-grpc)
├── auth                                   <----- HAND WRITTEN - the `authorization: Bearer` credential
├── tests                                  <----- HAND WRITTEN - the go test suite (see Testing below)
├── ondewo-s2t-api                             <----- submodule @ https://github.com/ondewo/ondewo-s2t-api
├── ondewo-proto-compiler                  <----- submodule @ https://github.com/ondewo/ondewo-proto-compiler
├── .github
│   └── workflows
│       └── ci.yml
├── CONTRIBUTING.md
├── go.mod                                 <----- module manifest, written by the compiler on the first run
├── go.sum
├── LICENSE
├── Makefile
├── README.md
└── RELEASE.md
```

## Regenerating the stubs

The generated code is produced by the `ondewo-go-proto-compiler` docker image, which is built from
the `ondewo-proto-compiler` submodule. The fixed image tag is the only contract between the two
repositories.

```shell
make update_submodules                      ## git submodule update --init --recursive
make checkout_defined_submodule_versions    ## check out the pins from the Makefile Variables chapter
make build_compiler                         ## build ondewo-go-proto-compiler:latest from the submodule
make generate_ondewo_protos                 ## generate api/ from ondewo-s2t-api/ondewo
make check_build                            ## assert every .proto produced a *.pb.go
make go_build                               ## compile the module
```

`make build` runs the whole chain in that order.

A few properties of the generation worth knowing:

* The image copies the mounted input volume into a temporary directory inside the container and
  compiles there, so the protos of `ondewo-s2t-api` are never modified.
* `api/` is **wiped** before every run, so a proto that was renamed or deleted upstream leaves no
  orphaned stub behind. Never put hand-written code below it.
* The Go import path is baked into every generated file by `protoc-gen-go`, so the module path is
  passed to the image as the third positional argument — it cannot be fixed up afterwards.
* `go.mod` and `go.sum` are written by the image **only when this repository has neither**. Once
  they are committed they are yours to maintain; the image prints the module requirements of the
  generated stubs at the end of every run so a drift is visible.
* Generation needs no network: every module the stubs are compiled against was pre-downloaded when
  the image was built.

## Testing

```shell
make check_stubs            ## assert the generated stubs are committed
make test                   ## go test over every package
make test_coverage          ## the same suite under -race, plus the hand-written coverage gate
make test_coverage_generated ## report (never gate) how much of api/ the suite exercises
```

The suite lives in `tests/` — never below `api/`, which is wiped on every regeneration — and needs
neither a network nor a running ONDEWO server. gRPC connections are made over an in-memory
`bufconn` listener, so a client stub, a server stub and a real HTTP/2 connection are exercised
in-process.

What it asserts about the **generated** code:

* `TranscribeFileRequest` — a bytes scalar, a nested `TranscribeRequestConfig`, an enum, a oneof
  member, a proto3 `optional` field and a well-known `google.protobuf.Struct` — survives
  `proto.Marshal` → `proto.Unmarshal` unchanged, and a truncated payload is rejected;
* the three levels of repeated messages a `TranscribeFileResponse` carries (transcription → word →
  word alternative) survive the same round trip;
* a proto3 `optional` scalar keeps its explicit presence — an explicitly set `""` is transmitted and
  arrives as a non-nil pointer, while an unset field stays `nil`. This is the distinction the
  angular target of the same compiler once lost, which made `false`/`0`/`""` unsendable;
* the enum zero value is pinned. S2T's `Decoding` is **not** an `*_UNSPECIFIED` enum — its zero
  member is `DEFAULT`, a real choice — so the test pins that member by name rather than assuming
  the convention; `ServiceTier`, which does follow it, is checked too;
* the `grpc.ServiceDesc` of `ondewo.s2t.Speech2Text` (from `protoc-gen-go-grpc`) lists exactly the
  RPCs its proto descriptor (from `protoc-gen-go`) does — the two plugins run separately and each
  half compiles on its own, so a disagreement is otherwise invisible;
* the generated `NewSpeech2TextClient` binds to a connection, and every generated **unary** stub is
  actually called over the wire and has to come back as `codes.Unimplemented` — 15 of them for this
  product, which proves each one marshals its request and builds a method name the transport
  accepts. The one bidirectional RPC, `TranscribeStream`, is covered by its own test;
* an RPC answered by a fake server round-trips its response, and one the server leaves to the
  generated `UnimplementedSpeech2TextServer` base type reports `codes.Unimplemented`;
* the compiled `.proto` file is registered in the global descriptor registry as proto3.

**Coverage.** The threshold (`COVERAGE_THRESHOLD` in the `Makefile`, currently **100%**) is
enforced over the hand-written packages only — `auth/` — because everything below `api/` is machine
output: gating on it would measure how much of protoc's output a test happens to walk. The stubs
are still exercised for real, as listed above; `make test_coverage_generated` prints their figure
(**17.1%** of generated statements at the time of writing) for the record. `make test_coverage` also
fails if it ends up measuring no hand-written function at all, so a deleted package cannot turn the
gate into a green no-op.

`.github/workflows/ci.yml` runs exactly these targets on `ubuntu-latest` against the go directive of
`go.mod` and the toolchain the compiler image generates with. It does **not** build the compiler
image or check out the submodules: it builds and tests the committed stubs, which is what a
consumer of the module gets.

## Release

The release is driven entirely by the `Makefile` — see `make help` for the full list of targets.

```shell
make ondewo_release                         ## credentials from the devops-accounts repo, then `make release`
```

`make release` builds, commits, creates the release branch, pushes **two** tags for the same commit
— the ONDEWO release tag (`7.5.0`) and the `v`-prefixed tag Go tooling requires (`v7.5.0`) — creates
the GitHub release from the matching `RELEASE.md` entry, and asks the public module proxy to fetch
the new version.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

Apache License 2.0 — see [LICENSE](LICENSE).
