package types

import (
	"testing"
)

func TestRequestBinarySerialization(t *testing.T) {
	// 测试用例
	testCases := []struct {
		name     string
		request  Request
		expected string
	}{
		{
			name: "empty data",
			request: Request{
				Type: "ping",
				Data: nil,
			},
			expected: "ping",
		},
		{
			name: "small data",
			request: Request{
				Type: "get_block",
				Data: []byte("block123"),
			},
			expected: "get_block",
		},
		{
			name: "large data",
			request: Request{
				Type: "sync",
				Data: make([]byte, 1000),
			},
			expected: "sync",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 序列化
			serialized, err := tc.request.Serialize()
			if err != nil {
				t.Fatalf("serialization failed: %v", err)
			}

			// 反序列化
			var deserialized Request
			err = deserialized.Deserialize(serialized)
			if err != nil {
				t.Fatalf("deserialization failed: %v", err)
			}

			// 验证结果
			if deserialized.Type != tc.request.Type {
				t.Errorf("Type mismatch: expected %s, got %s", tc.request.Type, deserialized.Type)
			}

			if len(deserialized.Data) != len(tc.request.Data) {
				t.Errorf("Data length mismatch: expected %d, got %d", len(tc.request.Data), len(deserialized.Data))
			}

			if len(tc.request.Data) > 0 {
				for i := range tc.request.Data {
					if deserialized.Data[i] != tc.request.Data[i] {
						t.Errorf("Data content mismatch at position %d: expected %d, got %d", i, tc.request.Data[i], deserialized.Data[i])
						break
					}
				}
			}
		})
	}
}

func TestResponseBinarySerialization(t *testing.T) {
	// 测试用例
	testCases := []struct {
		name     string
		response Response
		expected string
	}{
		{
			name: "empty data",
			response: Response{
				Type: "pong",
				Data: nil,
			},
			expected: "pong",
		},
		{
			name: "small data",
			response: Response{
				Type: "block_response",
				Data: []byte("block_data"),
			},
			expected: "block_response",
		},
		{
			name: "large data",
			response: Response{
				Type: "sync_response",
				Data: make([]byte, 2000),
			},
			expected: "sync_response",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 序列化
			serialized, err := tc.response.Serialize()
			if err != nil {
				t.Fatalf("serialization failed: %v", err)
			}

			// 反序列化
			var deserialized Response
			err = deserialized.Deserialize(serialized)
			if err != nil {
				t.Fatalf("deserialization failed: %v", err)
			}

			// 验证结果
			if deserialized.Type != tc.response.Type {
				t.Errorf("Type mismatch: expected %s, got %s", tc.response.Type, deserialized.Type)
			}

			if len(deserialized.Data) != len(tc.response.Data) {
				t.Errorf("Data length mismatch: expected %d, got %d", len(tc.response.Data), len(deserialized.Data))
			}

			if len(tc.response.Data) > 0 {
				for i := range tc.response.Data {
					if deserialized.Data[i] != tc.response.Data[i] {
						t.Errorf("Data content mismatch at position %d: expected %d, got %d", i, tc.response.Data[i], deserialized.Data[i])
						break
					}
				}
			}
		})
	}
}

func TestBinaryFormatEfficiency(t *testing.T) {
	// 测试二进制格式的效率
	request := Request{
		Type: "get_block",
		Data: []byte("block_data_123"),
	}

	// 序列化
	serialized, err := request.Serialize()
	if err != nil {
		t.Fatalf("serialization failed: %v", err)
	}

	// 验证二进制格式的大小
	// 期望大小: 4 bytes (header) + 9 bytes (type) + 13 bytes (data) = 26 bytes
	expectedSize := 4 + len(request.Type) + len(request.Data)
	if len(serialized) != expectedSize {
		t.Errorf("serialization size mismatch: expected %d, got %d", expectedSize, len(serialized))
	}

	t.Logf("二进制格式大小: %d bytes", len(serialized))
	t.Logf("原始数据大小: type=%d bytes, data=%d bytes", len(request.Type), len(request.Data))
}

func TestBinaryFormatLimits(t *testing.T) {
	// 测试长度限制
	largeType := string(make([]byte, 65536)) // 超过 int16 限制
	largeData := make([]byte, 65536)         // 超过 int16 限制

	// 测试过大的 type
	request := Request{
		Type: largeType,
		Data: []byte("test"),
	}
	_, err := request.Serialize()
	if err == nil {
		t.Error("should reject oversized type")
	}

	// 测试过大的 data
	request = Request{
		Type: "test",
		Data: largeData,
	}
	_, err = request.Serialize()
	if err == nil {
		t.Error("should reject oversized data")
	}
}
