package uaclient

import "testing"

func TestParseScalar(t *testing.T) {
	tests := []struct {
		name      string
		valueType string
		raw       string
		want      interface{}
	}{
		{name: "string", valueType: "string", raw: "abc", want: "abc"},
		{name: "bool", valueType: "bool", raw: "true", want: true},
		{name: "int8", valueType: "int8", raw: "-5", want: int8(-5)},
		{name: "int16", valueType: "int16", raw: "-12", want: int16(-12)},
		{name: "uint16", valueType: "uint16", raw: "12", want: uint16(12)},
		{name: "int32", valueType: "int32", raw: "42", want: int32(42)},
		{name: "uint8", valueType: "uint8", raw: "7", want: byte(7)},
		{name: "uint32", valueType: "uint32", raw: "42", want: uint32(42)},
		{name: "float32", valueType: "float32", raw: "12.5", want: float32(12.5)},
		{name: "float64", valueType: "float64", raw: "12.5", want: float64(12.5)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseScalar(tt.valueType, tt.raw)
			if err != nil {
				t.Fatalf("parseScalar returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("parseScalar() = %#v (%T), want %#v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}

func TestParseScalarRejectsUnsupportedType(t *testing.T) {
	_, err := parseScalar("xml-da", "42")
	if err == nil {
		t.Fatal("expected unsupported scalar type error")
	}
}
