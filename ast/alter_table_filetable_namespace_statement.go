package ast

// AlterTableFileTableNamespaceStatement represents ALTER TABLE ... ENABLE/DISABLE FILETABLE_NAMESPACE
type AlterTableFileTableNamespaceStatement struct {
	SchemaObjectName *SchemaObjectName `json:"SchemaObjectName,omitempty"`
	IsEnable         bool              `json:"IsEnable,omitempty"`
}

func (s *AlterTableFileTableNamespaceStatement) node()      {}
func (s *AlterTableFileTableNamespaceStatement) statement() {}
