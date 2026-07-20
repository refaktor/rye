// TODO: PersistentTable removed temporarily (persistent_table.go renamed to .NOT)
// Stub is always active until the real implementation is restored.

package env

// PersistentTable is a stub — real implementation in persistent_table.go.NOT
// Without badger, persistent tables are unavailable.
type PersistentTable struct {
	Cols []string
	Kind Word
}

func NewPersistentTable(cols []string, dbPath string, tableName string) (*PersistentTable, error) {
	return nil, nil
}

func (pt *PersistentTable) Close() error              { return nil }
func (pt *PersistentTable) AddRow(row TableRow)       {}
func (pt *PersistentTable) GetRows() []TableRow       { return nil }
func (pt *PersistentTable) Length() int               { return 0 }
func (pt *PersistentTable) NRows() int                { return 0 }
func (pt *PersistentTable) GetRow(ps *ProgramState, index int) TableRow {
	return TableRow{}
}
func (pt *PersistentTable) RemoveRowByIndex(index int64)     {}
func (pt *PersistentTable) GetColumn(name string) Object     { return nil }
func (pt *PersistentTable) GetColumns() List                 { return List{} }
func (pt *PersistentTable) GetColumnIndex(column string) int { return -1 }
func (pt *PersistentTable) GetColumnNames() []string         { return pt.Cols }
func (pt *PersistentTable) SetCols(vals []string)            { pt.Cols = vals }
func (pt *PersistentTable) GetRowValue(column string, rrow TableRow) (any, error) {
	return nil, nil
}
func (pt *PersistentTable) Columns(ps *ProgramState, names []string) Object { return nil }
func (pt *PersistentTable) Type() Type                                       { return PersistentTableType }
func (pt *PersistentTable) GetKind() int                                     { return int(PersistentTableType) }
func (pt *PersistentTable) Equal(o Object) bool                              { return false }
func (pt *PersistentTable) Inspect(e Idxs) string                           { return "[PersistentTable: unavailable]" }
func (pt *PersistentTable) Print(e Idxs) string                             { return "PTable[unavailable]" }
func (pt *PersistentTable) ToHtml() string                                  { return "" }
func (pt *PersistentTable) ToTxt() string                                   { return "" }
func (pt *PersistentTable) Trace(msg string)                                 {}
func (pt *PersistentTable) Dump(e Idxs) string                              { return "persistent-table { }" }
func (pt *PersistentTable) Get(i int) Object                                { return nil }
func (pt *PersistentTable) MakeNew(data []Object) Object                    { return pt }
