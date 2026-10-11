package utils

import (
	"encoding/hex"
	"github.com/everFinance/goar/types"
	"reflect"
	"testing"
)

func TestTagDecoderWireCompatibility(t *testing.T) {
	cases := [][]types.Tag{{{Name: "name", Value: "value"}}, {{Name: "non-ascii", Value: "中文"}, {Name: "empty", Value: ""}}}
	for _, tags := range cases {
		a, err := SerializeTags1(tags)
		if err != nil {
			t.Fatal(err)
		}
		b, err := SerializeTags(tags)
		if err != nil {
			t.Fatal(err)
		}
		if len(a) == 0 || len(b) == 0 {
			t.Fatal("non-empty tags produced empty Avro")
		}
		for _, decode := range []struct {
			name string
			fn   func([]byte) ([]types.Tag, error)
		}{{"legacy string decoder", DeserializeTags1}, {"byte decoder", DeserializeTags}} {
			for _, encoded := range [][]byte{a, b} {
				got, err := decode.fn(encoded)
				if err != nil {
					t.Fatalf("%s: %v", decode.name, err)
				}
				if !reflect.DeepEqual(got, tags) {
					t.Fatalf("%s decoded tags = %v, want %v", decode.name, got, tags)
				}
			}
		}
	}
	tags := []types.Tag{{Name: "name", Value: "value"}}
	got, err := SerializeTags1(tags)
	if err != nil {
		t.Fatal(err)
	}
	if hex.EncodeToString(got) != "0116086e616d650a76616c756500" {
		t.Fatalf("known Avro wire vector changed: %x", got)
	}
}

func TestTagDecodersRejectMalformedInput(t *testing.T) {
	for _, data := range [][]byte{nil, {2}, {2, 128}, {2, 6, 'a', 'b', 'c'}, {1, 1}, {0, 2}, {255, 255, 255, 255, 255, 255, 255, 255, 255, 1, 0, 0}} {
		for _, decode := range []struct {
			name string
			fn   func([]byte) ([]types.Tag, error)
		}{{"legacy string decoder", DeserializeTags1}, {"byte decoder", DeserializeTags}} {
			func() {
				defer func() {
					if p := recover(); p != nil {
						t.Errorf("decoder panicked for %x: %v", data, p)
					}
				}()
				if _, err := decode.fn(data); err == nil {
					t.Errorf("%s accepted malformed %x", decode.name, data)
				}
			}()
		}
	}
}

func TestTagDecodersAcceptEmptyAndMultipleBlocks(t *testing.T) {
	payload, err := hex.DecodeString("086e616d650a76616c7565")
	if err != nil {
		t.Fatal(err)
	}
	multi := append([]byte{2}, payload...)
	multi = append(multi, 1, byte(2*len(payload)))
	multi = append(multi, payload...)
	multi = append(multi, 0)
	for _, decode := range []struct {
		name string
		fn   func([]byte) ([]types.Tag, error)
	}{{"legacy", DeserializeTags1}, {"bytes", DeserializeTags}} {
		empty, err := decode.fn([]byte{0})
		if err != nil || len(empty) != 0 {
			t.Fatalf("%s empty block: %v, %v", decode.name, empty, err)
		}
		tags, err := decode.fn(multi)
		want := []types.Tag{{Name: "name", Value: "value"}, {Name: "name", Value: "value"}}
		if err != nil || !reflect.DeepEqual(tags, want) {
			t.Fatalf("%s mixed blocks = %v, %v", decode.name, tags, err)
		}
	}
}
