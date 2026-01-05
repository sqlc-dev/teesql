package ast

// XmlCompressionOption represents an XML compression option
type XmlCompressionOption struct {
	IsCompressed    string                     // "On", "Off"
	PartitionRanges []*CompressionPartitionRange
	OptionKind      string // "XmlCompression"
}

func (x *XmlCompressionOption) node()        {}
func (x *XmlCompressionOption) tableOption() {}
func (x *XmlCompressionOption) indexOption() {}

// TableXmlCompressionOption represents a table-level XML compression option
type TableXmlCompressionOption struct {
	XmlCompressionOption *XmlCompressionOption
	OptionKind           string // "XmlCompression"
}

func (t *TableXmlCompressionOption) node()        {}
func (t *TableXmlCompressionOption) tableOption() {}
