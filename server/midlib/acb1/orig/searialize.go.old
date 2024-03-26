package Acb1

import "bytes"

// SearializeDocumentType searalizes a documentType for hashing.
func SearializeDocumentType(dt documentType) []byte {
	var buf bytes.Buffer
	// binary.Write(&buf, binary.BigEndian, bk.Index)
	buf.Write([]byte(dt.Title))
	buf.Write([]byte(dt.Desc))
	buf.Write([]byte(dt.Category))
	buf.Write([]byte(dt.Tags))
	buf.Write([]byte(dt.ImageList))
	return buf.Bytes()
}

/*
type MetaDocumentType struct {
	Document      DocumentType
	DocumentID    string
	CategoryID    string
	OverallHash   string
	MerkleHash    string
	ImageListHash []string
	LeafHash      []string
}
*/

// SearializeMetaDocumentType searalizes a metaDocumentType for hashing.
func SearializeMetaDocument(doc metaDocumentType) []byte {
	var buf bytes.Buffer
	buf.Write(SearializeDocumentType(doc.Document))
	buf.Write([]byte(doc.MerkleHash))
	buf.Write([]byte(doc.PdfHash))
	return buf.Bytes()
}

/* vim: set noai ts=4 sw=4: */
