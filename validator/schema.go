package validator

/*
#cgo pkg-config: libxml-2.0
#include <libxml/parser.h>
#include <libxml/xmlschemas.h>
#include <stdlib.h>

static xmlSchemaPtr compileSchema(const char* buf, int size) {
    xmlSchemaParserCtxtPtr ctxt = xmlSchemaNewMemParserCtxt(buf, size);
    if (ctxt == NULL) {
        return NULL;
    }
    xmlSchemaPtr schema = xmlSchemaParse(ctxt);
    xmlSchemaFreeParserCtxt(ctxt);
    return schema;
}

static xmlSchemaPtr compileSchemaFile(const char* filename) {
    xmlSchemaParserCtxtPtr ctxt = xmlSchemaNewParserCtxt(filename);
    if (ctxt == NULL) {
        return NULL;
    }
    xmlSchemaPtr schema = xmlSchemaParse(ctxt);
    xmlSchemaFreeParserCtxt(ctxt);
    return schema;
}

static int validateDoc(xmlSchemaPtr schema, const char* docbuf, int size) {
    int opts = XML_PARSE_NONET;
    xmlDocPtr doc = xmlReadMemory(docbuf, size, "noname.xml", NULL, opts);
    if (doc == NULL) {
        return -1;
    }
    xmlSchemaValidCtxtPtr vctxt = xmlSchemaNewValidCtxt(schema);
    int ret = xmlSchemaValidateDoc(vctxt, doc);
    xmlSchemaFreeValidCtxt(vctxt);
    xmlFreeDoc(doc);
    return ret;
}

static void freeSchema(xmlSchemaPtr schema) {
    if (schema) xmlSchemaFree(schema);
}
*/
import "C"
import (
	"errors"
	"math"
	"runtime"
	"unsafe"
)

// Schema wraps a compiled XML Schema.
type Schema struct {
	ptr *C.xmlSchema
}

// Compile compiles raw XSD bytes into a Schema.
func Compile(data []byte) (*Schema, error) {
	if len(data) == 0 {
		return nil, errors.New("empty schema")
	}
	if len(data) > math.MaxInt32 {
		return nil, errors.New("schema too large")
	}
	ptr := C.compileSchema((*C.char)(unsafe.Pointer(&data[0])), C.int(len(data)))
	if ptr == nil {
		return nil, errors.New("failed to compile schema")
	}
	s := &Schema{ptr: ptr}
	runtime.SetFinalizer(s, func(sc *Schema) { sc.Free() })
	return s, nil
}

// CompileFile compiles an XSD schema from a file path.
func CompileFile(path string) (*Schema, error) {
	if path == "" {
		return nil, errors.New("empty path")
	}
	if len(path) > math.MaxInt32 {
		return nil, errors.New("path too long")
	}
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	ptr := C.compileSchemaFile(cpath)
	if ptr == nil {
		return nil, errors.New("failed to compile schema")
	}
	s := &Schema{ptr: ptr}
	runtime.SetFinalizer(s, func(sc *Schema) { sc.Free() })
	return s, nil
}

// Validate validates XML bytes against the schema.
func (s *Schema) Validate(xml []byte) error {
	if s.ptr == nil {
		return errors.New("schema not initialized")
	}
	if len(xml) == 0 {
		return errors.New("empty xml")
	}
	if len(xml) > math.MaxInt32 {
		return errors.New("xml input too large")
	}
	ret := C.validateDoc(s.ptr, (*C.char)(unsafe.Pointer(&xml[0])), C.int(len(xml)))
	if ret != 0 {
		return errors.New("validation failed")
	}
	return nil
}

// Free releases resources associated with the schema.
func (s *Schema) Free() {
	if s.ptr != nil {
		C.freeSchema(s.ptr)
		s.ptr = nil
	}
}
