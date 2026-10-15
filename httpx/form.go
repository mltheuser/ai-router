package httpx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/textproto"
)

// Form is a multipart/form-data request body: text fields and files, sent in
// the order they were added.
type Form struct {
	parts []formPart
}

type formPart struct {
	name string
	// value is a text field's value, or a file's content.
	value []byte
	// filename and contentType are set for files only.
	filename    string
	contentType string
}

// Field adds a text field.
func (f *Form) Field(name, value string) {
	f.parts = append(f.parts, formPart{name: name, value: []byte(value)})
}

// File adds a file with the given file name and content type.
func (f *Form) File(name, filename, contentType string, content []byte) {
	f.parts = append(f.parts, formPart{name: name, value: content, filename: filename, contentType: contentType})
}

// encode returns the form as a request body and its Content-Type header.
func (f *Form) encode() (body []byte, contentType string, err error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for _, p := range f.parts {
		if p.filename == "" {
			err = w.WriteField(p.name, string(p.value))
		} else {
			err = writeFile(w, p)
		}
		if err != nil {
			return nil, "", err
		}
	}
	if err := w.Close(); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), w.FormDataContentType(), nil
}

func writeFile(w *multipart.Writer, p formPart) error {
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", multipart.FileContentDisposition(p.name, p.filename))
	header.Set("Content-Type", p.contentType)
	part, err := w.CreatePart(header)
	if err != nil {
		return err
	}
	_, err = part.Write(p.value)
	return err
}

// summary renders the form for the debug log as a JSON object: text fields
// as is, files as their name, type and size. Their content, typically
// binary, would make the log unreadable.
func (f *Form) summary() []byte {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, p := range f.parts {
		if i > 0 {
			buf.WriteByte(',')
		}
		value := string(p.value)
		if p.filename != "" {
			value = fmt.Sprintf("<file %s, %s, %d bytes>", p.filename, p.contentType, len(p.value))
		}
		buf.Write(jsonString(p.name))
		buf.WriteByte(':')
		buf.Write(jsonString(value))
	}
	buf.WriteByte('}')
	return buf.Bytes()
}

// jsonString encodes s as a JSON string, leaving <, > and & unescaped so the
// log stays legible.
func jsonString(s string) []byte {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(s) // encoding a string cannot fail
	return bytes.TrimSuffix(buf.Bytes(), []byte("\n"))
}
