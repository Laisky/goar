package utils

import (
	"encoding/binary"
	"errors"
	"github.com/everFinance/goar/types"
	"github.com/linkedin/goavro/v2"
)

func TagsEncode(tags []types.Tag) []types.Tag {
	base64Tags := []types.Tag{}

	for _, tag := range tags {
		base64Tags = append(base64Tags, types.Tag{
			Name:  Base64Encode([]byte(tag.Name)),
			Value: Base64Encode([]byte(tag.Value)),
		})
	}

	return base64Tags
}

func TagsDecode(base64Tags []types.Tag) ([]types.Tag, error) {
	tags := []types.Tag{}

	for _, bt := range base64Tags {
		bName, err := Base64Decode(bt.Name)
		if err != nil {
			return nil, err
		}

		bValue, err := Base64Decode(bt.Value)
		if err != nil {
			return nil, err
		}

		tags = append(tags, types.Tag{
			Name:  string(bName),
			Value: string(bValue),
		})
	}

	return tags, nil
}

// using bundle tx, avro serialize
func SerializeTags1(tags []types.Tag) ([]byte, error) {
	if len(tags) == 0 {
		return make([]byte, 0), nil
	}
	codec, err := goavro.NewCodec(avroTagRecordSchema)
	if err != nil {
		return nil, err
	}
	var payload []byte
	for _, tag := range tags {
		payload, err = codec.BinaryFromNative(payload, map[string]interface{}{"name": []byte(tag.Name), "value": []byte(tag.Value)})
		if err != nil {
			return nil, err
		}
	}
	// Preserve the legacy serializer's negative-count, sized array block.
	// Avro strings and bytes have identical binary field representations.
	var prefix [2 * binary.MaxVarintLen64]byte
	n := binary.PutVarint(prefix[:], -int64(len(tags)))
	n += binary.PutVarint(prefix[n:], int64(len(payload)))
	encoded := make([]byte, 0, n+len(payload)+1)
	encoded = append(encoded, prefix[:n]...)
	encoded = append(encoded, payload...)
	return append(encoded, 0), nil
}

func DeserializeTags1(data []byte) ([]types.Tag, error) {
	// Avro string and byte values share the same binary representation. Use
	// the decoder that reports truncated input rather than discarding EOF.
	return DeserializeTags(data)
}

const avroTagRecordSchema = `{"type":"record","name":"Tag","fields":[{"name":"name","type":"bytes"},{"name":"value","type":"bytes"}]}`
const avroTagSchema = `{"type":"array","items":` + avroTagRecordSchema + `}`

func SerializeTags(tags []types.Tag) ([]byte, error) {
	if len(tags) == 0 {
		return make([]byte, 0), nil
	}

	codec, err := goavro.NewCodec(avroTagSchema)
	if err != nil {
		return nil, err
	}
	avroTags := []map[string]interface{}{}
	for _, tag := range tags {
		m := map[string]interface{}{"name": []byte(tag.Name), "value": []byte(tag.Value)}
		avroTags = append(avroTags, m)
	}

	data, err := codec.BinaryFromNative(nil, avroTags)
	if err != nil {
		return nil, err
	}

	return data, err
}

func DeserializeTags(data []byte) ([]types.Tag, error) {
	codec, err := goavro.NewCodec(avroTagSchema)
	if err != nil {
		return nil, err
	}

	avroTags, remaining, err := codec.NativeFromBinary(data)
	if err != nil {
		return nil, err
	}

	if len(remaining) != 0 {
		return nil, errors.New("trailing Avro tag bytes")
	}

	tags := []types.Tag{}

	for _, v := range avroTags.([]interface{}) {
		tag := v.(map[string]interface{})
		tags = append(tags, types.Tag{Name: string(tag["name"].([]byte)), Value: string(tag["value"].([]byte))})
	}
	return tags, err
}
