package ard

import (
	"bytes"
	templatepkg "text/template"
)

func DecodeTemplate(code string, data any, format string) (Value, error) {
	if template, err := templatepkg.New("").Parse(code); err == nil {
		var buffer bytes.Buffer
		if err := template.Execute(&buffer, data); err == nil {
			value, _, err := Read(&buffer, format, false)
			return value, err
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}
