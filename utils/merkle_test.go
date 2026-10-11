package utils

import (
	"fmt"
	"github.com/everFinance/goar/types"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// func TestGenerateChunks(t *testing.T) {
// 	data, _ := Base64Decode("NzcyNg")
// 	assert.Equal(t, "z3rQGxyiqdQuOh2dxDst176oOKmW3S9MwQNTEh4DK1U", Base64Encode(GenerateChunks(data).DataRoot))
//
// 	data, err := os.ReadFile("./testfile/1mb.bin")
// 	assert.NoError(t, err)
// 	chunks := GenerateChunks(data)
// 	assert.Equal(t, "o1tTTjbC7hIZN6KbUUYjlkQoDl2k8VXNuBDcGIs52Hc", Base64Encode(chunks.DataRoot))
// }

func TestChunkData(t *testing.T) {
	data, _ := Base64Decode("NzcyNg")

	assert.Equal(t, "55w6-CA_Um7muHLnJvBlUUKTjpa35cPLv1PCIPQs6M8", Base64Encode(chunkData(data)[0].DataHash))
}

func TestGenerateLeaves(t *testing.T) {
	data, _ := Base64Decode("NzcyNg")

	assert.Equal(t, "z3rQGxyiqdQuOh2dxDst176oOKmW3S9MwQNTEh4DK1U", (Base64Encode(generateLeaves(chunkData(data))[0].ID)))
}

func TestChunkStream(t *testing.T) {
	for _, size := range []int{0, 1, types.MIN_CHUNK_SIZE - 1, types.MIN_CHUNK_SIZE, types.MAX_CHUNK_SIZE - 1, types.MAX_CHUNK_SIZE, types.MAX_CHUNK_SIZE + 1, types.MAX_CHUNK_SIZE + types.MIN_CHUNK_SIZE - 1, 2*types.MAX_CHUNK_SIZE + 1} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			data := make([]byte, size)
			for i := range data {
				data[i] = byte(i*31 + 7)
			}
			path := filepath.Join(t.TempDir(), "synthetic.bin")
			require.NoError(t, os.WriteFile(path, data, 0600))
			dataReader, err := os.Open(path)
			require.NoError(t, err)
			defer dataReader.Close()
			chunks01 := chunkData(data)
			chunks02, err := chunkStreamData(dataReader)
			require.NoError(t, err)
			require.Equal(t, len(chunks01), len(chunks02))
			for i := range chunks01 {
				assert.Equal(t, chunks01[i].MaxByteRange, chunks02[i].MaxByteRange)
				assert.Equal(t, chunks01[i].MinByteRange, chunks02[i].MinByteRange)
				assert.Equal(t, chunks01[i].DataHash, chunks02[i].DataHash)
			}
		})
	}
}

func TestDecodeEmptyString(t *testing.T) {
	data := ""
	by, err := Base64Decode(data)
	assert.NoError(t, err)
	t.Log(by)
}
