package core

import (
	"reflect"
	"testing"
)

func TestReadSimpleString(t *testing.T) {
	tests := []struct {
		data []byte
		want string
		n    int
	}{
		{[]byte("+OK\r\n"), "OK", 5},
		{[]byte("+PONG\r\n"), "PONG", 7},
		{[]byte("+\r\n"), "", 3},
		{[]byte("+hello world\r\n"), "hello world", 14},
	}

	for _, tt := range tests {
		got, n, err := readSimpleString(tt.data)
		if err != nil {
			t.Errorf("readSimpleString(%q) = %v", tt.data, err)
			continue
		}
		if got != tt.want || n != tt.n {
			t.Errorf("readSimpleString(%q) = (%q, %d), want (%q, %d)", tt.data, got, n, tt.want, tt.n)
		}
	}
}

func TestReadError(t *testing.T) {
	tests := []struct {
		data []byte
		want string
		n    int
	}{
		{[]byte("-Error\r\n"), "Error", 8},
		{[]byte("-ERR unknown command\r\n"), "ERR unknown command", 22},
		{[]byte("-\r\n"), "", 3},
	}

	for _, tt := range tests {
		got, n, err := readError(tt.data)
		if err != nil {
			t.Errorf("readError(%q) = %v", tt.data, err)
			continue
		}
		if got != tt.want || n != tt.n {
			t.Errorf("readError(%q) = (%q, %d), want (%q, %d)", tt.data, got, n, tt.want, tt.n)
		}
	}
}

func TestReadInt64(t *testing.T) {
	tests := []struct {
		data []byte
		want int64
		n    int
	}{
		{[]byte(":0\r\n"), 0, 4},
		{[]byte(":1000\r\n"), 1000, 7},
		{[]byte(":42\r\n"), 42, 5},
	}

	for _, tt := range tests {
		got, n, err := readInt64(tt.data)
		if err != nil {
			t.Errorf("readInt64(%q) = %v", tt.data, err)
			continue
		}
		if got != tt.want || n != tt.n {
			t.Errorf("readInt64(%q) = (%d, %d), want (%d, %d)", tt.data, got, n, tt.want, tt.n)
		}
	}
}

func TestReadLength(t *testing.T) {
	tests := []struct {
		data   []byte
		length int
		n      int
	}{
		{[]byte("0\r\n"), 0, 3},
		{[]byte("5\r\nhello\r\n"), 5, 3},
		{[]byte("12\r\nabcdefghijkl\r\n"), 12, 4},
	}

	for _, tt := range tests {
		length, n := readLength(tt.data)
		if length != tt.length || n != tt.n {
			t.Errorf("readLength(%q) = (%d, %d), want (%d, %d)", tt.data, length, n, tt.length, tt.n)
		}
	}
}

func TestReadBulkString(t *testing.T) {
	tests := []struct {
		data []byte
		want string
		n    int
	}{
		{[]byte("$0\r\n\r\n"), "", 6},
		{[]byte("$5\r\nhello\r\n"), "hello", 11},
		{[]byte("$11\r\nhello world\r\n"), "hello world", 18},
	}

	for _, tt := range tests {
		got, n, err := readBulkString(tt.data)
		if err != nil {
			t.Errorf("readBulkString(%q) = %v", tt.data, err)
			continue
		}
		if got != tt.want || n != tt.n {
			t.Errorf("readBulkString(%q) = (%q, %d), want (%q, %d)", tt.data, got, n, tt.want, tt.n)
		}
	}
}

func TestReadArray(t *testing.T) {
	tests := []struct {
		data []byte
		want interface{}
		n    int
	}{
		{[]byte("*0\r\n"), []interface{}{}, 4},
		{[]byte("*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n"), []interface{}{"hello", "world"}, 26},
		{[]byte("*3\r\n:1\r\n:2\r\n:3\r\n"), []interface{}{int64(1), int64(2), int64(3)}, 16},
	}

	for _, tt := range tests {
		got, n, err := readArray(tt.data)
		if err != nil {
			t.Errorf("readArray(%q) = %v", tt.data, err)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) || n != tt.n {
			t.Errorf("readArray(%q) = (%v, %d), want (%v, %d)", tt.data, got, n, tt.want, tt.n)
		}
	}
}

func TestDecodeOne(t *testing.T) {
	tests := []struct {
		data []byte
		want interface{}
	}{
		{[]byte("+OK\r\n"), "OK"},
		{[]byte("-Error\r\n"), "Error"},
		{[]byte(":1000\r\n"), int64(1000)},
		{[]byte("$5\r\nhello\r\n"), "hello"},
		{[]byte("*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n"), []interface{}{"hello", "world"}},
	}

	for _, tt := range tests {
		got, _, err := DecodeOne(tt.data)
		if err != nil {
			t.Errorf("DecodeOne(%q) = %v", tt.data, err)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("DecodeOne(%q) = %v, want %v", tt.data, got, tt.want)
		}
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		data []byte
		want interface{}
	}{
		{[]byte("+OK\r\n"), "OK"},
		{[]byte("-Error\r\n"), "Error"},
		{[]byte(":1000\r\n"), int64(1000)},
		{[]byte("$5\r\nhello\r\n"), "hello"},
		{[]byte("*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n"), []interface{}{"hello", "world"}},
	}

	for _, tt := range tests {
		got, err := Decode(tt.data)
		if err != nil {
			t.Errorf("Decode(%q) = %v", tt.data, err)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Decode(%q) = %v, want %v", tt.data, got, tt.want)
		}
	}
}

func TestArrayDecode(t *testing.T) {
	cases := map[string][]interface{}{
		"*0\r\n":                                                   {},
		"*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n":                     {"hello", "world"},
		"*3\r\n:1\r\n:2\r\n:3\r\n":                                 {int64(1), int64(2), int64(3)},
		"*5\r\n:1\r\n:2\r\n:3\r\n:4\r\n$5\r\nhello\r\n":            {int64(1), int64(2), int64(3), int64(4), "hello"},
		"*2\r\n*3\r\n:1\r\n:2\r\n:3\r\n*2\r\n+Hello\r\n-World\r\n": {[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{"Hello", "World"}},
	}

	for data, want := range cases {
		got, err := Decode([]byte(data))
		if err != nil {
			t.Errorf("Decode(%q) = %v", data, err)
			continue
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Decode(%q) = %v, want %v", data, got, want)
		}
	}
}
