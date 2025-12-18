package ast

// FetchType represents the orientation for a FETCH statement.
type FetchType struct {
	Orientation string           `json:"Orientation,omitempty"`
	RowOffset   ScalarExpression `json:"RowOffset,omitempty"`
}

// DeclareCursorStatement represents DECLARE cursor_name CURSOR FOR SELECT.
type DeclareCursorStatement struct {
	Name             *Identifier       `json:"Name,omitempty"`
	CursorDefinition *CursorDefinition `json:"CursorDefinition,omitempty"`
}

func (s *DeclareCursorStatement) node()      {}
func (s *DeclareCursorStatement) statement() {}

// OpenCursorStatement represents OPEN cursor_name.
type OpenCursorStatement struct {
	Cursor *CursorId `json:"Cursor,omitempty"`
}

func (s *OpenCursorStatement) node()      {}
func (s *OpenCursorStatement) statement() {}

// CloseCursorStatement represents CLOSE cursor_name.
type CloseCursorStatement struct {
	Cursor *CursorId `json:"Cursor,omitempty"`
}

func (s *CloseCursorStatement) node()      {}
func (s *CloseCursorStatement) statement() {}

// DeallocateCursorStatement represents DEALLOCATE cursor_name.
type DeallocateCursorStatement struct {
	Cursor *CursorId `json:"Cursor,omitempty"`
}

func (s *DeallocateCursorStatement) node()      {}
func (s *DeallocateCursorStatement) statement() {}

// FetchCursorStatement represents FETCH cursor_name.
type FetchCursorStatement struct {
	FetchType     *FetchType         `json:"FetchType,omitempty"`
	Cursor        *CursorId          `json:"Cursor,omitempty"`
	IntoVariables []ScalarExpression `json:"IntoVariables,omitempty"`
}

func (s *FetchCursorStatement) node()      {}
func (s *FetchCursorStatement) statement() {}
