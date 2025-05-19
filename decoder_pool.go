package cotlib

import (
	"bytes"
	"encoding/xml"
	"sync"
)

// pooledDecoder wraps an xml.Decoder with a reusable bytes.Reader.
type pooledDecoder struct {
	dec *xml.Decoder
	br  *bytes.Reader
}

var decoderPool = sync.Pool{
	New: func() any {
		br := bytes.NewReader(nil)
		return &pooledDecoder{dec: xml.NewDecoder(br), br: br}
	},
}

func getDecoder(data []byte) *pooledDecoder {
	pd := decoderPool.Get().(*pooledDecoder)
	pd.br.Reset(data)
	pd.dec = xml.NewDecoder(pd.br)
	pd.dec.CharsetReader = nil
	pd.dec.Entity = nil
	return pd
}

func putDecoder(pd *pooledDecoder) {
	pd.br.Reset(nil)
	decoderPool.Put(pd)
}
