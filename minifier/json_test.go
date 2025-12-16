// Author: João Pinto
// Date: 2025-12-16
// Purpose: teste unitário para o tratamento de conteúdo JSON
// License: MIT

package minifier

import (
	"testing"
)

func TestMinifyJSON(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {
            name: "Simple object",
            input: `{
                "name": "João",
                "age": 30
            }`,
            expected: `{"name":"João","age":30}`,
        },
        {
            name: "Nested arrays",
            input: `{
                "items": [1, 2, 3],
                "nested": {"a": true}
            }`,
            expected: `{"items":[1,2,3],"nested":{"a":true}}`,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := MinifyJSON(tt.input)
            if got != tt.expected {
                t.Errorf("got %q, want %q", got, tt.expected)
            }
        })
    }
}
