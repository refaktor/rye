//go:build b_smtpd_MORE_RAW
// +build b_smtpd_MORE_RAW

package evaldo

import (
	// "bytes"
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"strings"

	"github.com/refaktor/rye/env"
	// "github.com/jinzhu/copier"
)

type EmailBody struct {
	Text       []string
	HTML       []string
	Attachment []string
}

func retrieveEmailBodiesAndAttachments(email *mail.Message) (*EmailBody, []string, error) {
	emailBodies := &EmailBody{}
	var attachments []string

	mediaType, params, err := mime.ParseMediaType(email.Header.Get("Content-Type"))
	if err != nil {
		return nil, nil, err
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(email.Body, params["boundary"])
		for {
			p, err := mr.NextPart()
			if err != nil {
				break
			}

			partContentType := p.Header.Get("Content-Type")
			switch {
			case strings.HasPrefix(partContentType, "text/plain"):
				body, err := readBody(p)
				if err != nil {
					return nil, nil, err
				}
				emailBodies.Text = append(emailBodies.Text, body)

			case strings.HasPrefix(partContentType, "text/html"):
				body, err := readBody(p)
				if err != nil {
					return nil, nil, err
				}
				emailBodies.HTML = append(emailBodies.HTML, body)

			default:
				// Treat it as an attachment
				attachmentFilename, err := decodeRFC2047(p.FileName())
				if err != nil {
					return nil, nil, err
				}
				attachments = append(attachments, attachmentFilename)
			}
		}
	} else if strings.HasPrefix(mediaType, "text/plain") {
		body, err := readBodyFull(email.Body, mediaType)
		if err != nil {
			return nil, nil, err
		}
		emailBodies.Text = append(emailBodies.Text, body)
	} else if strings.HasPrefix(mediaType, "text/html") {
		body, err := readBodyFull(email.Body, mediaType)
		if err != nil {
			return nil, nil, err
		}
		emailBodies.HTML = append(emailBodies.HTML, body)
	}

	return emailBodies, attachments, nil
}

func readBody(part *multipart.Part) (string, error) {
	bodyBytes, err := ioutil.ReadAll(part)
	if err != nil {
		return "", err
	}

	mediaType, _, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
	if err != nil {
		return "", err
	}

	decodedBody, err := decodeBody(bodyBytes, mediaType)
	if err != nil {
		return "", err
	}

	return decodedBody, nil
}

func readBodyFull(reader io.Reader, mediaType string) (string, error) {
	bodyBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}

	decodedBody, err := decodeBody(bodyBytes, mediaType)
	if err != nil {
		return "", err
	}

	return decodedBody, nil
}

func decodeBody(body []byte, mediaType string) (string, error) {
	// Check if the body is encoded with a transfer encoding
	if transferEncoding := strings.ToLower(mediaType); transferEncoding != "" {
		body, err := decodeTransferEncoding(body, transferEncoding)
		if err != nil {
			return "", err
		}
		return string(body), nil
	}

	// If no transfer encoding, assume UTF-8 and return the body as is
	return string(body), nil
}

func decodeTransferEncoding(body []byte, transferEncoding string) ([]byte, error) {
	switch transferEncoding {
	case "quoted-printable":
		return ioutil.ReadAll(quotedPrintableReader(strings.NewReader(string(body))))

	case "base64":
		return ioutil.ReadAll(base64Reader(strings.NewReader(string(body))))

	default:
		return nil, fmt.Errorf("unsupported transfer encoding: %s", transferEncoding)
	}
}

func quotedPrintableReader(r io.Reader) io.Reader {
	return bufio.NewReader(quotedprintable.NewReader(r))
}

func base64Reader(r io.Reader) io.Reader {
	return base64.NewDecoder(base64.StdEncoding, r)
}

func decodeRFC2047(s string) (string, error) {
	dec := new(mime.WordDecoder)
	return dec.DecodeHeader(s)
}

var Builtins_mimeMail = map[string]*env.Builtin{

	"mail-message//header?": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, arg0, "smtpd")
		},
	},

	"mail-message//body?": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, arg0, "smtpd")
		},
	},

	"parse-media-type": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, arg0, "smtpd")
		},
	},

	"mail-message//for-parts": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, arg0, "smtpd")
		},
	},
}

// todo - NAUK PO
// * msfg.header.Get(subject)
// .... attachment , text, gXSXS
