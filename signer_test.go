package huaweicloud

import (
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestHmacsha256(t *testing.T) {
	tests := []struct {
		name     string
		key      []byte
		data     string
		expected string
	}{
		{
			name:     "empty_string",
			key:      []byte("test_key"),
			data:     "",
			expected: "d056b2b640f407a9daeba0b13c3b3966e5b69e84283ec3c7fa0cac56a02208a7",
		},
		{
			name:     "simple_string",
			key:      []byte("test_key"),
			data:     "test_data",
			expected: "46a5b27b7e6672271c998f4d79ed460ff03c88cacd31355ffc161539e1657824",
		},
		{
			name:     "special_chars",
			key:      []byte("test_key"),
			data:     "!@#$%^&*()",
			expected: "b6713c4273ea051f45164f94af0f4bcd9783310375e347203ac64dc55fb3ba6b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := hmacsha256(tt.key, tt.data)
			if err != nil {
				t.Errorf("hmacsha256() error = %v", err)
				return
			}
			if got := fmt.Sprintf("%x", result); got != tt.expected {
				t.Errorf("hmacsha256() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCanonicalURI(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "simple_path",
			path:     "/test/path",
			expected: "/test/path/",
		},
		{
			name:     "path_with_special_chars",
			path:     "/test/path with spaces",
			expected: "/test/path%20with%20spaces/",
		},
		{
			name:     "empty_path",
			path:     "",
			expected: "/",
		},
		{
			name:     "root_path",
			path:     "/",
			expected: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://example.com"+tt.path, nil)
			if got := CanonicalURI(req); got != tt.expected {
				t.Errorf("CanonicalURI() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCanonicalQueryString(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{
			name:     "simple_query",
			query:    "a=1&b=2",
			expected: "a=1&b=2",
		},
		{
			name:     "query_with_special_chars",
			query:    "a=1 2&b=3 4",
			expected: "a=1%202&b=3%204",
		},
		{
			name:     "empty_query",
			query:    "",
			expected: "",
		},
		{
			name:     "multiple_values",
			query:    "a=1&a=2&b=3",
			expected: "a=1&a=2&b=3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://example.com?"+tt.query, nil)
			if got := CanonicalQueryString(req); got != tt.expected {
				t.Errorf("CanonicalQueryString() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCanonicalHeaders(t *testing.T) {
	tests := []struct {
		name          string
		headers       map[string]string
		signedHeaders []string
		expected      string
	}{
		{
			name: "simple_headers",
			headers: map[string]string{
				"Content-Type": "application/json",
				"Host":         "example.com",
			},
			signedHeaders: []string{"content-type", "host"},
			expected:      "content-type:application/json\nhost:example.com\n",
		},
		{
			name: "headers_with_spaces",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Custom":     "  value with spaces  ",
			},
			signedHeaders: []string{"content-type", "x-custom"},
			expected:      "content-type:application/json\nx-custom:value with spaces\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://example.com", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			got := CanonicalHeaders(req, tt.signedHeaders)
			if got != tt.expected {
				t.Errorf("CanonicalHeaders() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSignedHeaders(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected []string
	}{
		{
			name: "simple_headers",
			headers: map[string]string{
				"Content-Type": "application/json",
				"Host":         "example.com",
			},
			expected: []string{"content-type", "host"},
		},
		{
			name: "mixed_case_headers",
			headers: map[string]string{
				"Content-Type": "application/json",
				"HOST":         "example.com",
			},
			expected: []string{"content-type", "host"},
		},
		{
			name:     "empty_headers",
			headers:  map[string]string{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://example.com", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			got := SignedHeaders(req)
			sort.Strings(got)
			sort.Strings(tt.expected)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("SignedHeaders() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRequestPayload(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected string
	}{
		{
			name:     "empty_body",
			body:     "",
			expected: "",
		},
		{
			name:     "simple_body",
			body:     "test body",
			expected: "test body",
		},
		{
			name:     "json_body",
			body:     `{"key":"value"}`,
			expected: `{"key":"value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "http://example.com", strings.NewReader(tt.body))
			got, err := RequestPayload(req)
			if err != nil {
				t.Errorf("RequestPayload() error = %v", err)
				return
			}
			if string(got) != tt.expected {
				t.Errorf("RequestPayload() = %v, want %v", string(got), tt.expected)
			}
		})
	}
}

func TestStringToSign(t *testing.T) {
	tests := []struct {
		name           string
		canonicalReq   string
		timestamp      time.Time
		expectedPrefix string
	}{
		{
			name:           "simple_request",
			canonicalReq:   "GET\n/test/\n\nhost:example.com\n\nhost\ne3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			timestamp:      time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			expectedPrefix: SignAlgorithm + "\n20240101T120000Z\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StringToSign(tt.canonicalReq, tt.timestamp)
			if err != nil {
				t.Errorf("StringToSign() error = %v", err)
				return
			}
			if !strings.HasPrefix(got, tt.expectedPrefix) {
				t.Errorf("StringToSign() = %v, want prefix %v", got, tt.expectedPrefix)
			}
		})
	}
}

func TestSignStringToSign(t *testing.T) {
	tests := []struct {
		name         string
		stringToSign string
		signingKey   []byte
		expectedLen  int
	}{
		{
			name:         "simple_signature",
			stringToSign: "test string to sign",
			signingKey:   []byte("test_key"),
			expectedLen:  64, // SHA256 hex string length
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SignStringToSign(tt.stringToSign, tt.signingKey)
			if err != nil {
				t.Errorf("SignStringToSign() error = %v", err)
				return
			}
			if len(got) != tt.expectedLen {
				t.Errorf("SignStringToSign() length = %v, want %v", len(got), tt.expectedLen)
			}
		})
	}
}

func TestHexEncodeSHA256Hash(t *testing.T) {
	tests := []struct {
		name     string
		body     []byte
		expected string
	}{
		{
			name:     "empty_body",
			body:     []byte(""),
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "simple_body",
			body:     []byte("test body"),
			expected: "63efb315ed71cc7e5a1fc202434bb3aec2091e7838707e148a017faebb7464fe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HexEncodeSHA256Hash(tt.body)
			if err != nil {
				t.Errorf("HexEncodeSHA256Hash() error = %v", err)
				return
			}
			if got != tt.expected {
				t.Errorf("HexEncodeSHA256Hash() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAuthHeaderValue(t *testing.T) {
	tests := []struct {
		name           string
		signature      string
		accessKey      string
		signedHeaders  []string
		expectedPrefix string
	}{
		{
			name:           "simple_auth",
			signature:      "test_signature",
			accessKey:      "test_key",
			signedHeaders:  []string{"host", "content-type"},
			expectedPrefix: SignAlgorithm + " Access=test_key, SignedHeaders=host;content-type, Signature=test_signature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AuthHeaderValue(tt.signature, tt.accessKey, tt.signedHeaders)
			if got != tt.expectedPrefix {
				t.Errorf("AuthHeaderValue() = %v, want %v", got, tt.expectedPrefix)
			}
		})
	}
}

func TestSigner_Sign(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		secret  string
		request *http.Request
		wantErr bool
	}{
		{
			name:   "simple_request",
			key:    "test_key",
			secret: "test_secret",
			request: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com/test", nil)
				req.Header.Set("Host", "example.com")
				return req
			}(),
			wantErr: false,
		},
		{
			name:   "request_with_body",
			key:    "test_key",
			secret: "test_secret",
			request: func() *http.Request {
				req, _ := http.NewRequest("POST", "http://example.com/test", strings.NewReader(`{"key":"value"}`))
				req.Header.Set("Host", "example.com")
				req.Header.Set("Content-Type", "application/json")
				return req
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Signer{
				Key:    tt.key,
				Secret: tt.secret,
			}
			if err := s.Sign(tt.request); (err != nil) != tt.wantErr {
				t.Errorf("Signer.Sign() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if auth := tt.request.Header.Get(HeaderXAuthorization); auth == "" {
					t.Error("Signer.Sign() did not set Authorization header")
				}
				if date := tt.request.Header.Get(HeaderXDateTime); date == "" {
					t.Error("Signer.Sign() did not set X-Sdk-Date header")
				}
			}
		})
	}
}

// 基准测试
func BenchmarkSigner_Sign(b *testing.B) {
	s := &Signer{
		Key:    "test_key",
		Secret: "test_secret",
	}
	req, _ := http.NewRequest("GET", "http://example.com/test", nil)
	req.Header.Set("Host", "example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Sign(req)
	}
}

func BenchmarkHmacsha256(b *testing.B) {
	key := []byte("test_key")
	data := "test data"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hmacsha256(key, data)
	}
}

func BenchmarkHexEncodeSHA256Hash(b *testing.B) {
	data := []byte("test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HexEncodeSHA256Hash(data)
	}
}
