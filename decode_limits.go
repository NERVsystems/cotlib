package cotlib

import "encoding/xml"

// limitTokenReader wraps an xml.Decoder and enforces XML security limits
// while streaming tokens. It checks element depth, element count,
// attribute/character data length, and token length as tokens are read.
type limitTokenReader struct {
	dec   *xml.Decoder
	depth int
	count int
}

func (l *limitTokenReader) Token() (xml.Token, error) {
	off := l.dec.InputOffset()
	tok, err := l.dec.RawToken()
	if err != nil {
		return tok, err
	}
	if l.dec.InputOffset()-off > currentMaxTokenLen() {
		return nil, ErrInvalidInput
	}
	switch t := tok.(type) {
	case xml.StartElement:
		l.depth++
		l.count++
		if l.depth > int(currentMaxElementDepth()) || l.count > int(currentMaxElementCount()) {
			return nil, ErrInvalidInput
		}
		for _, a := range t.Attr {
			if len(a.Value) > int(currentMaxValueLen()) {
				return nil, ErrInvalidInput
			}
		}
	case xml.EndElement:
		if l.depth > 0 {
			l.depth--
		}
	case xml.CharData:
		if len(t) > int(currentMaxValueLen()) {
			return nil, ErrInvalidInput
		}
	}
	return tok, nil
}

// decodeWithLimits decodes XML using the provided decoder while enforcing
// security limits during tokenization.
func decodeWithLimits(dec *xml.Decoder, v any) error {
	ltd := &limitTokenReader{dec: dec}
	secure := xml.NewTokenDecoder(ltd)
	return secure.Decode(v)
}
