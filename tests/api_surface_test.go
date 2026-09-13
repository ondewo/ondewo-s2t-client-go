// Copyright 2020-2026 ONDEWO GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Everything in this file names the ONDEWO S2T API specifically: its service, its messages, one of
// its enums. It is the ONLY file that differs from the sibling go clients - generated_code_test.go
// and auth_test.go are product agnostic and are copied over unchanged.
package tests

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	s2t "github.com/ondewo/ondewo-s2t-client-go/api/ondewo/s2t"
)

// protoFileCount is the number of .proto files below ondewo-s2t-api/ondewo that the compiler
// consumed - S2T ships a single service proto, ondewo/s2t/speech-to-text.proto. Every one of them
// has to end up in the global descriptor registry when this package is linked; a proto that
// silently stopped being compiled is otherwise invisible until a consumer misses a type.
const protoFileCount = 1

// services is every gRPC service this product exposes, keyed by the fully qualified proto name
// the ServiceDesc must declare.
var services = map[string]*grpc.ServiceDesc{
	"ondewo.s2t.Speech2Text": &s2t.Speech2Text_ServiceDesc,
}

// clientConstructors is the generated New<Service>Client of every service above. A client SDK
// that compiles but whose constructors are missing is useless, and the two generators that
// produce them (protoc-gen-go, protoc-gen-go-grpc) can disagree - so both halves are listed.
var clientConstructors = map[string]func(grpc.ClientConnInterface) any{
	"ondewo.s2t.Speech2Text": func(cc grpc.ClientConnInterface) any { return s2t.NewSpeech2TextClient(cc) },
}

// expectedMethods pins RPCs by name. The descriptor cross-check in generated_code_test.go proves
// the two generators agree with each other; it cannot notice an RPC that was renamed upstream,
// because both halves would be renamed together. These are spelled out so that a rename is a
// failing test rather than a silently broken consumer.
//
// The names are the ones in the .proto - `GetS2tPipeline`, not the `GetS2TPipeline` that
// protoc-gen-go-grpc capitalises the go method to. Both halves of the generated code carry the
// proto spelling in the ServiceDesc, which is what the wire uses.
var expectedMethods = map[string][]string{
	// TranscribeStream is bidirectional, so it lives in ServiceDesc.Streams, not .Methods - the
	// lookup has to consider both.
	"ondewo.s2t.Speech2Text": {
		"TranscribeFile",
		"TranscribeStream",
		"GetS2tPipeline",
		"CreateS2tPipeline",
		"UpdateS2tPipeline",
		"DeleteS2tPipeline",
		"ListS2tPipelines",
		"ListS2tLanguages",
		"ListS2tDomains",
		"GetServiceInfo",
	},
}

// TestMessageRoundTripsThroughTheWire is the core assertion about generated message code: a value
// built in go, serialized and parsed back is the same value. The message chosen covers every
// field kind the S2T generator has to get right at once - a bytes scalar, a nested message, an
// enum, a oneof member, a proto3 `optional` scalar and a well-known google.protobuf.Struct - so a
// generator that mis-numbers a field or loses a nested type fails here.
func TestMessageRoundTripsThroughTheWire(t *testing.T) {
	t.Parallel()

	// A Struct carries the service credentials of a cloud S2T provider. Built from the fixed
	// reference instant so the value compares exactly instead of racing the clock.
	serviceConfig, err := structpb.NewStruct(map[string]any{
		"api_key":    "8f14e45fceea167a5a36dedd4bea2543",
		"region":     "eu-central-1",
		"issued_at":  referenceTime.Format("2006-01-02T15:04:05Z07:00"),
		"max_alerts": float64(3),
	})
	if err != nil {
		t.Fatalf("structpb.NewStruct failed: %v", err)
	}

	original := &s2t.TranscribeFileRequest{
		AudioFile: []byte{0x52, 0x49, 0x46, 0x46, 0x00, 0x01, 0x02, 0x03},
		Config: &s2t.TranscribeRequestConfig{
			S2TPipelineId: "default_german_1",
			Decoding:      s2t.Decoding_BEAM_SEARCH_WITH_LM,
			OneofLanguageModelName: &s2t.TranscribeRequestConfig_LanguageModelName{
				LanguageModelName: "de_lm_v3",
			},
			Language:         proto.String("de-DE"),
			Task:             proto.String("transcribe"),
			S2TServiceConfig: serviceConfig,
		},
	}

	wire, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("proto.Marshal(%T) failed: %v", original, err)
	}
	if len(wire) == 0 {
		t.Fatal("proto.Marshal produced 0 bytes for a fully populated message")
	}

	parsed := &s2t.TranscribeFileRequest{}
	if err := proto.Unmarshal(wire, parsed); err != nil {
		t.Fatalf("proto.Unmarshal failed: %v", err)
	}

	if !proto.Equal(original, parsed) {
		t.Fatalf("round trip changed the message:\n original = %v\n  parsed = %v", original, parsed)
	}
	if got, want := parsed.GetConfig().GetDecoding(), s2t.Decoding_BEAM_SEARCH_WITH_LM; got != want {
		t.Errorf("enum after round trip = %v, want %v", got, want)
	}
	if got, want := parsed.GetConfig().GetLanguageModelName(), "de_lm_v3"; got != want {
		t.Errorf("oneof member after round trip = %q, want %q", got, want)
	}
	if got, want := parsed.GetConfig().GetS2TServiceConfig().GetFields()["region"].GetStringValue(), "eu-central-1"; got != want {
		t.Errorf("google.protobuf.Struct field after round trip = %q, want %q", got, want)
	}
}

// TestRepeatedNestedMessagesRoundTrip covers the response side, where S2T nests three levels of
// repeated messages: a transcription holds words, and every word holds its alternatives. A
// generator that loses a nested type or mis-numbers a repeated field fails here rather than in a
// consumer reading an empty list.
func TestRepeatedNestedMessagesRoundTrip(t *testing.T) {
	t.Parallel()

	original := &s2t.TranscribeFileResponse{
		Transcriptions: []*s2t.Transcription{
			{
				Transcription:   "guten morgen",
				ConfidenceScore: 0.93,
				Words: []*s2t.WordDetail{
					{
						StartTime:  0.25,
						EndTime:    0.61,
						Word:       "guten",
						Confidence: 0.97,
						WordAlternatives: []*s2t.WordAlternative{
							{Word: "gute", Confidence: 0.02},
							{Word: "guter", Confidence: 0.01},
						},
					},
				},
				Alternatives: []*s2t.TranscriptionAlternative{
					{Transcript: "guten morgen!", Confidence: 0.41},
				},
			},
		},
		Time:      1.75,
		AudioUuid: "0f3a4a53-9a4c-1b0a-5a6f-7c8d6b1d0c3e",
	}

	wire, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("proto.Marshal(%T) failed: %v", original, err)
	}

	parsed := &s2t.TranscribeFileResponse{}
	if err := proto.Unmarshal(wire, parsed); err != nil {
		t.Fatalf("proto.Unmarshal failed: %v", err)
	}

	if !proto.Equal(original, parsed) {
		t.Fatalf("round trip changed the message:\n original = %v\n  parsed = %v", original, parsed)
	}
	if got, want := len(parsed.GetTranscriptions()[0].GetWords()[0].GetWordAlternatives()), 2; got != want {
		t.Errorf("third-level repeated message has %d entries after the round trip, want %d", got, want)
	}
	if got, want := parsed.GetTranscriptions()[0].GetWords()[0].GetWord(), "guten"; got != want {
		t.Errorf("nested word after round trip = %q, want %q", got, want)
	}
}

// TestProto3ExplicitPresenceSurvivesTheWire guards the field kind that generators get wrong: a
// proto3 `optional` scalar has to keep the difference between "set to the zero value" and "not
// set". The angular target of the same compiler lost exactly this distinction, which made a
// false/0/"" unsendable; protoc-gen-go models it as a pointer, and this asserts it stays that way.
// S2T declares 105 such fields; TranscribeRequestConfig.language is one of them.
func TestProto3ExplicitPresenceSurvivesTheWire(t *testing.T) {
	t.Parallel()

	t.Run("zero value set explicitly is transmitted", func(t *testing.T) {
		t.Parallel()

		wire, err := proto.Marshal(&s2t.TranscribeRequestConfig{Language: proto.String("")})
		if err != nil {
			t.Fatalf("proto.Marshal failed: %v", err)
		}
		if len(wire) == 0 {
			t.Fatal("an explicitly set zero value was dropped from the wire - proto3 presence is lost")
		}

		parsed := &s2t.TranscribeRequestConfig{}
		if err := proto.Unmarshal(wire, parsed); err != nil {
			t.Fatalf("proto.Unmarshal failed: %v", err)
		}
		if parsed.Language == nil {
			t.Fatal("Language is nil after the round trip, want a pointer to the empty string")
		}
		if got := *parsed.Language; got != "" {
			t.Errorf("Language = %q, want the empty string", got)
		}
	})

	t.Run("unset stays unset", func(t *testing.T) {
		t.Parallel()

		wire, err := proto.Marshal(&s2t.TranscribeRequestConfig{S2TPipelineId: "unset"})
		if err != nil {
			t.Fatalf("proto.Marshal failed: %v", err)
		}

		parsed := &s2t.TranscribeRequestConfig{}
		if err := proto.Unmarshal(wire, parsed); err != nil {
			t.Fatalf("proto.Unmarshal failed: %v", err)
		}
		if parsed.Language != nil {
			t.Errorf("Language = %q after a round trip that never set it, want nil", *parsed.Language)
		}
	})
}

// TestEnumZeroValueIsPinned pins the member at 0 and the name maps generated beside it. S2T's
// Decoding does NOT follow the *_UNSPECIFIED convention - its zero member is DEFAULT, which is a
// real choice ("use whatever the pipeline config says") - so the assertion is that the zero value
// is the one the API documents, not that it carries a particular name. Renaming or reordering the
// members silently changes what an unset request field means, which is what this catches.
func TestEnumZeroValueIsPinned(t *testing.T) {
	t.Parallel()

	var zero s2t.Decoding

	if zero != s2t.Decoding_DEFAULT {
		t.Errorf("zero value of Decoding = %v, want DEFAULT", zero)
	}
	if got, want := zero.String(), "DEFAULT"; got != want {
		t.Errorf("Decoding(0).String() = %q, want %q", got, want)
	}
	if got, want := s2t.Decoding_name[0], "DEFAULT"; got != want {
		t.Errorf("Decoding_name[0] = %q, want %q", got, want)
	}
	if got, want := s2t.Decoding_value["BEAM_SEARCH_WITH_LM"], int32(s2t.Decoding_BEAM_SEARCH_WITH_LM); got != want {
		t.Errorf("Decoding_value[BEAM_SEARCH_WITH_LM] = %d, want %d", got, want)
	}
	if got, want := int32(s2t.Decoding_BEAM_SEARCH_WITH_LM), int32(2); got != want {
		t.Errorf("BEAM_SEARCH_WITH_LM = %d, want %d", got, want)
	}

	// The enums that DO follow the convention have to keep doing so: a zero value that is a real
	// choice rather than "unspecified" is unrequestable in several of the other clients of this API.
	if got, want := s2t.ServiceTier(0), s2t.ServiceTier_SERVICE_TIER_UNSPECIFIED; got != want {
		t.Errorf("zero value of ServiceTier = %v, want %v", got, want)
	}
}

// TestUnmarshalRejectsTruncatedInput asserts the generated message reports a parse error instead
// of accepting a malformed payload: field 1 (`s2t_pipeline_id`) is announced as 5 bytes long but
// only 1 follows.
func TestUnmarshalRejectsTruncatedInput(t *testing.T) {
	t.Parallel()

	if err := proto.Unmarshal([]byte{0x0a, 0x05, 'a'}, &s2t.TranscribeRequestConfig{}); err == nil {
		t.Fatal("proto.Unmarshal accepted a truncated payload, want an error")
	}
}

// speech2TextServer is a fake ONDEWO server: it answers GetS2TPipeline and inherits the
// "unimplemented" behaviour of the generated base type for every other RPC of the service.
type speech2TextServer struct {
	s2t.UnimplementedSpeech2TextServer
}

func (speech2TextServer) GetS2TPipeline(_ context.Context, req *s2t.S2TPipelineId) (*s2t.Speech2TextConfig, error) {
	return &s2t.Speech2TextConfig{
		Id:     req.GetId(),
		Active: true,
		Description: &s2t.S2TDescription{
			Language:      "de",
			PipelineOwner: "ondewo",
			Domain:        "general",
		},
	}, nil
}

// TestUnaryRPCRoundTripsOverAnInProcessServer drives the generated client stub, the generated
// server stub and the generated ServiceDesc against each other over a real gRPC connection - the
// request is marshalled, routed by the method name baked into the stub, and the response is
// parsed back. Nothing here is mocked except the transport, which is in memory.
func TestUnaryRPCRoundTripsOverAnInProcessServer(t *testing.T) {
	t.Parallel()

	conn := dialInProcess(t, nil, func(srv *grpc.Server) {
		s2t.RegisterSpeech2TextServer(srv, speech2TextServer{})
	})
	client := s2t.NewSpeech2TextClient(conn)

	const id = "default_german_1"
	response, err := client.GetS2TPipeline(t.Context(), &s2t.S2TPipelineId{Id: id})
	if err != nil {
		t.Fatalf("GetS2TPipeline failed: %v", err)
	}

	if got := response.GetId(); got != id {
		t.Errorf("response id = %q, want %q", got, id)
	}
	if !response.GetActive() {
		t.Error("response active = false, want true")
	}
	if got, want := response.GetDescription().GetLanguage(), "de"; got != want {
		t.Errorf("response description language = %q, want %q", got, want)
	}
}

// TestUnimplementedMethodIsReportedAsUnimplemented pins the other half of the generated server
// contract: an RPC the server does not implement must come back as codes.Unimplemented, not as a
// routing failure or a panic. It also proves the method is routed at all - a method missing from
// the ServiceDesc would surface as a different code.
func TestUnimplementedMethodIsReportedAsUnimplemented(t *testing.T) {
	t.Parallel()

	conn := dialInProcess(t, nil, func(srv *grpc.Server) {
		s2t.RegisterSpeech2TextServer(srv, speech2TextServer{})
	})
	client := s2t.NewSpeech2TextClient(conn)

	_, err := client.DeleteS2TPipeline(t.Context(), &s2t.S2TPipelineId{Id: "default_german_1"})
	if got := status.Code(err); got != codes.Unimplemented {
		t.Fatalf("DeleteS2TPipeline returned code %v (err = %v), want %v", got, err, codes.Unimplemented)
	}
}

// TestBidiStreamingStubOpensAStream covers the one RPC the unary sweep in generated_code_test.go
// deliberately skips. Opening a stream against a server that does not implement it is answered on
// the first Recv rather than at call time, so both halves are asserted here: the generated stub
// hands out a stream, and the server reports the RPC as unimplemented over it.
func TestBidiStreamingStubOpensAStream(t *testing.T) {
	t.Parallel()

	conn := dialInProcess(t, nil, func(srv *grpc.Server) {
		s2t.RegisterSpeech2TextServer(srv, speech2TextServer{})
	})
	client := s2t.NewSpeech2TextClient(conn)

	stream, err := client.TranscribeStream(t.Context())
	if err != nil {
		t.Fatalf("TranscribeStream failed to open a stream: %v", err)
	}

	// The send may or may not fail depending on when the server's status reaches the client, so
	// the assertion is made on Recv, which always observes it.
	_ = stream.Send(&s2t.TranscribeStreamRequest{})

	if _, err := stream.Recv(); status.Code(err) != codes.Unimplemented {
		t.Fatalf("TranscribeStream Recv returned code %v (err = %v), want %v", status.Code(err), err, codes.Unimplemented)
	}
}
