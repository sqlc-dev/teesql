package ast

// CreateSpatialIndexStatement represents a CREATE SPATIAL INDEX statement
type CreateSpatialIndexStatement struct {
	Name                  *Identifier
	Object                *SchemaObjectName
	SpatialColumnName     *Identifier
	SpatialIndexingScheme string // "None", "GeometryGrid", "GeographyGrid", "GeometryAutoGrid", "GeographyAutoGrid"
	OnFileGroup           *IdentifierOrValueExpression
	SpatialIndexOptions   []SpatialIndexOption
}

func (s *CreateSpatialIndexStatement) node()      {}
func (s *CreateSpatialIndexStatement) statement() {}

// SpatialIndexOption is an interface for spatial index options
type SpatialIndexOption interface {
	Node
	spatialIndexOption()
}

// SpatialIndexRegularOption wraps a regular IndexOption for spatial indexes
type SpatialIndexRegularOption struct {
	Option IndexOption
}

func (s *SpatialIndexRegularOption) node()               {}
func (s *SpatialIndexRegularOption) spatialIndexOption() {}

// BoundingBoxSpatialIndexOption represents a BOUNDING_BOX option
type BoundingBoxSpatialIndexOption struct {
	BoundingBoxParameters []*BoundingBoxParameter
}

func (b *BoundingBoxSpatialIndexOption) node()               {}
func (b *BoundingBoxSpatialIndexOption) spatialIndexOption() {}

// BoundingBoxParameter represents a bounding box parameter (XMIN, YMIN, XMAX, YMAX)
type BoundingBoxParameter struct {
	Parameter string // "None", "XMin", "YMin", "XMax", "YMax"
	Value     ScalarExpression
}

func (b *BoundingBoxParameter) node() {}

// GridsSpatialIndexOption represents a GRIDS option
type GridsSpatialIndexOption struct {
	GridParameters []*GridParameter
}

func (g *GridsSpatialIndexOption) node()               {}
func (g *GridsSpatialIndexOption) spatialIndexOption() {}

// GridParameter represents a grid parameter
type GridParameter struct {
	Parameter string // "None", "Level1", "Level2", "Level3", "Level4"
	Value     string // "Low", "Medium", "High"
}

func (g *GridParameter) node() {}

// CellsPerObjectSpatialIndexOption represents a CELLS_PER_OBJECT option
type CellsPerObjectSpatialIndexOption struct {
	Value ScalarExpression
}

func (c *CellsPerObjectSpatialIndexOption) node()               {}
func (c *CellsPerObjectSpatialIndexOption) spatialIndexOption() {}

// DataCompressionOption represents a DATA_COMPRESSION option for indexes
type DataCompressionOption struct {
	CompressionLevel string // "None", "Row", "Page", "ColumnStore", "ColumnStoreArchive"
	OptionKind       string // "DataCompression"
	PartitionRanges  []*CompressionPartitionRange
}

func (d *DataCompressionOption) node()            {}
func (d *DataCompressionOption) indexOption()     {}
func (d *DataCompressionOption) dropIndexOption() {}

// IgnoreDupKeyIndexOption represents the IGNORE_DUP_KEY option
type IgnoreDupKeyIndexOption struct {
	OptionState               string // "On", "Off"
	OptionKind                string // "IgnoreDupKey"
	SuppressMessagesOption    *bool  // true/false when SUPPRESS_MESSAGES specified
}

func (i *IgnoreDupKeyIndexOption) node()        {}
func (i *IgnoreDupKeyIndexOption) indexOption() {}
