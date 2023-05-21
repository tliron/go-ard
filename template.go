package ard

import (
	"bytes"
	templatepkg "text/template"
)

func DecodeYAMLTemplate(code string, data any) (Value, error) {
	if template, err := templatepkg.New("").Parse(code); err == nil {
		var buffer bytes.Buffer
		if err := template.Execute(&buffer, data); err == nil {
			value, _, err := ReadYAML(&buffer, false)
			return value, err
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}
