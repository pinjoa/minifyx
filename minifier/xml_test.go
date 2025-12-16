// Author: João Pinto
// Date: 2025-12-15
// Purpose: teste unitário para o tratamento de conteúdo XML
// License: MIT

package minifier

import (
	"testing"
)

func TestMinifyXML(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {
            name:     "Remove spaces and newlines",
            input:    "<root>\n   <child> value </child>\n</root>",
            expected: "<root><child>value</child></root>",
        },
        {
            name:     "Preserve attributes",
            input:    `<root><child id="123">text</child></root>`,
            expected: `<root><child id="123">text</child></root>`,
        },
        {
            name:     "Remove comments",
            input:    "<root><!-- comment --><child>ok</child></root>",
            expected: "<root><child>ok</child></root>",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := MinifyXML(tt.input, nil)
            if got != tt.expected {
                t.Errorf("got %q, want %q", got, tt.expected)
            }
        })
    }
}
