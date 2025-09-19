package env

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/dgraph-io/badger/v4"
)

// PersistentTable implements TableInterface with BadgerDB persistence
type PersistentTable struct {
	Cols      []string
	Kind      Word
	Indexes   map[string]map[any][]int
	db        *badger.DB
	tableName string
	mu        sync.RWMutex
	rowCount  int64
}

// SerializableRow represents a row that can be serialized to JSON
type SerializableRow struct {
	Values []interface{} `json:"values"`
	ID     int64         `json:"id"`
}

// NewPersistentTable creates a new persistent table with BadgerDB backend
func NewPersistentTable(cols []string, dbPath string, tableName string) (*PersistentTable, error) {
	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil // Disable logging

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open BadgerDB: %v", err)
	}

	pt := &PersistentTable{
		Cols:      cols,
		Indexes:   make(map[string]map[any][]int),
		db:        db,
		tableName: tableName,
		rowCount:  0,
	}

	// Load existing metadata
	err = pt.loadMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata: %v", err)
	}

	return pt, nil
}

// Close closes the database connection
func (pt *PersistentTable) Close() error {
	return pt.db.Close()
}

// loadMetadata loads table metadata from the database
func (pt *PersistentTable) loadMetadata() error {
	return pt.db.View(func(txn *badger.Txn) error {
		// Load column information
		colKey := fmt.Sprintf("%s:cols", pt.tableName)
		item, err := txn.Get([]byte(colKey))
		if err == badger.ErrKeyNotFound {
			// First time setup - save columns
			return pt.saveMetadata()
		}
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &pt.Cols)
		})
		if err != nil {
			return err
		}

		// Load row count
		countKey := fmt.Sprintf("%s:count", pt.tableName)
		item, err = txn.Get([]byte(countKey))
		if err == badger.ErrKeyNotFound {
			pt.rowCount = 0
			return nil
		}
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &pt.rowCount)
		})
	})
}

// saveMetadata saves table metadata to the database
func (pt *PersistentTable) saveMetadata() error {
	return pt.db.Update(func(txn *badger.Txn) error {
		// Save columns
		colKey := fmt.Sprintf("%s:cols", pt.tableName)
		colData, err := json.Marshal(pt.Cols)
		if err != nil {
			return err
		}
		err = txn.Set([]byte(colKey), colData)
		if err != nil {
			return err
		}

		// Save row count
		countKey := fmt.Sprintf("%s:count", pt.tableName)
		countData, err := json.Marshal(pt.rowCount)
		if err != nil {
			return err
		}
		return txn.Set([]byte(countKey), countData)
	})
}

// AddRow adds a row to the persistent table
func (pt *PersistentTable) AddRow(row TableRow) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.rowCount++
	rowID := pt.rowCount

	serRow := SerializableRow{
		Values: row.Values,
		ID:     rowID,
	}

	err := pt.db.Update(func(txn *badger.Txn) error {
		key := fmt.Sprintf("%s:row:%d", pt.tableName, rowID)
		data, err := json.Marshal(serRow)
		if err != nil {
			return err
		}
		return txn.Set([]byte(key), data)
	})

	if err != nil {
		fmt.Printf("Error adding row: %v\n", err)
		pt.rowCount-- // Rollback count on error
		return
	}

	// Update metadata
	pt.saveMetadata()
}

// GetRows retrieves all rows from the persistent table
func (pt *PersistentTable) GetRows() []TableRow {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	var rows []TableRow

	pt.db.View(func(txn *badger.Txn) error {
		prefix := fmt.Sprintf("%s:row:", pt.tableName)
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek([]byte(prefix)); it.ValidForPrefix([]byte(prefix)); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var serRow SerializableRow
				err := json.Unmarshal(val, &serRow)
				if err != nil {
					return err
				}

				row := TableRow{
					Values: serRow.Values,
					Uplink: pt,
				}
				rows = append(rows, row)
				return nil
			})
			if err != nil {
				fmt.Printf("Error reading row: %v\n", err)
			}
		}
		return nil
	})

	return rows
}

// GetRow retrieves a specific row by index
func (pt *PersistentTable) GetRow(ps *ProgramState, index int) TableRow {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	var row TableRow
	rowID := int64(index + 1) // Convert 0-based index to 1-based ID

	err := pt.db.View(func(txn *badger.Txn) error {
		key := fmt.Sprintf("%s:row:%d", pt.tableName, rowID)
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			var serRow SerializableRow
			err := json.Unmarshal(val, &serRow)
			if err != nil {
				return err
			}

			row = TableRow{
				Values: serRow.Values,
				Uplink: pt,
			}
			return nil
		})
	})

	if err != nil {
		fmt.Printf("Error getting row %d: %v\n", index, err)
		return TableRow{Values: make([]any, len(pt.Cols)), Uplink: pt}
	}

	return row
}

// RemoveRowByIndex removes a row by index
func (pt *PersistentTable) RemoveRowByIndex(index int64) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	rowID := index + 1 // Convert 0-based index to 1-based ID

	err := pt.db.Update(func(txn *badger.Txn) error {
		key := fmt.Sprintf("%s:row:%d", pt.tableName, rowID)
		return txn.Delete([]byte(key))
	})

	if err != nil {
		fmt.Printf("Error removing row %d: %v\n", index, err)
	}
}

// GetColumn retrieves a specific column
func (pt *PersistentTable) GetColumn(name string) Object {
	rows := pt.GetRows()
	col1 := make([]Object, len(rows))
	idx := slices.Index[[]string](pt.Cols, name)
	if idx > -1 {
		for i, row := range rows {
			col1[i] = ToRyeValue(row.Values[idx])
		}
		return *NewBlock(*NewTSeries(col1))
	} else {
		return NewError("Column not found")
	}
}

// GetColumns returns the list of column names
func (pt *PersistentTable) GetColumns() List {
	lst := make([]any, len(pt.Cols))
	for i, v := range pt.Cols {
		lst[i] = v
	}
	return *NewList(lst)
}

// GetColumnIndex returns the index of a column
func (pt *PersistentTable) GetColumnIndex(column string) int {
	index := -1
	for i, v := range pt.Cols {
		if v == column {
			index = i
			break
		}
	}
	return index
}

// GetColumnNames returns the column names
func (pt *PersistentTable) GetColumnNames() []string {
	return pt.Cols
}

// SetCols sets the column names
func (pt *PersistentTable) SetCols(vals []string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.Cols = vals
	pt.saveMetadata()
}

// GetRowValue gets a value from a row by column name
func (pt *PersistentTable) GetRowValue(column string, rrow TableRow) (any, error) {
	index := -1
	for i, v := range pt.Cols {
		if v == column {
			index = i
			break
		}
	}
	if index < 0 {
		return "", fmt.Errorf("column %s not found", column)
	}
	return rrow.Values[index], nil
}

// Columns returns a new table with only the specified columns
func (pt *PersistentTable) Columns(ps *ProgramState, names []string) Object {
	idxs := make([]int, len(names))
	for name := range names {
		idx := slices.Index[[]string](pt.Cols, names[name])
		if idx == -1 {
			return makeError(ps, "Col not found")
		}
		idxs[name] = idx
	}

	// Create a new in-memory table for the result
	nspr := NewTable(names)
	rows := pt.GetRows()

	for _, row := range rows {
		row2 := make([]any, len(names))
		for col := range idxs {
			if len(row.Values) > col {
				row2[col] = row.Values[idxs[col]]
			}
		}
		nspr.AddRow(TableRow{row2, nspr})
	}
	return *nspr
}

// Length returns the number of rows
func (pt *PersistentTable) Length() int {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	return int(pt.rowCount)
}

// NRows returns the number of rows
func (pt *PersistentTable) NRows() int {
	return pt.Length()
}

// Type returns the type
func (pt *PersistentTable) Type() Type {
	return PersistentTableType
}

// GetKind returns the kind
func (pt *PersistentTable) GetKind() int {
	return int(PersistentTableType)
}

// Equal compares two persistent tables
func (pt *PersistentTable) Equal(o Object) bool {
	if pt.Type() != o.Type() {
		return false
	}
	other, ok := o.(*PersistentTable)
	if !ok {
		return false
	}

	if len(pt.Cols) != len(other.Cols) {
		return false
	}

	for i, col := range pt.Cols {
		if col != other.Cols[i] {
			return false
		}
	}

	return pt.tableName == other.tableName
}

// Inspect returns a string representation
func (pt *PersistentTable) Inspect(e Idxs) string {
	rows := strconv.Itoa(pt.Length())
	var kindStr string
	if pt.GetKind() != int(PersistentTableType) {
		kindStr = " of kind " + pt.Kind.Print(e)
	}
	return "[PersistentTable(" + strconv.Itoa(len(pt.Cols)) + " " + rows + ")" + kindStr + "]"
}

// Print returns a text representation
func (pt *PersistentTable) Print(e Idxs) string {
	return pt.ToTxt()
}

// ToHtml returns HTML representation
func (pt *PersistentTable) ToHtml() string {
	var bu strings.Builder
	bu.WriteString("<table>")
	rows := pt.GetRows()
	for _, row := range rows {
		bu.WriteString("<tr>")
		for _, val := range row.Values {
			bu.WriteString("<td>")
			bu.WriteString(fmt.Sprint(val))
			bu.WriteString("</td>")
		}
		bu.WriteString("</tr>")
	}
	bu.WriteString("</table>")
	return bu.String()
}

// ToTxt returns text representation
func (pt *PersistentTable) ToTxt() string {
	var bu strings.Builder
	for _, name := range pt.Cols {
		bu.WriteString(fmt.Sprint(name))
		bu.WriteString("\t|")
	}
	bu.WriteString("\n")
	rows := pt.GetRows()
	for _, row := range rows {
		for _, val := range row.Values {
			bu.WriteString(fmt.Sprint(val))
			bu.WriteString("\t|")
		}
		bu.WriteString("\n")
	}
	return bu.String()
}

// Dump returns a dump representation
func (pt *PersistentTable) Dump(e Idxs) string {
	var sb strings.Builder
	sb.WriteString("persistent-table {")

	for _, col := range pt.Cols {
		sb.WriteString(" ")
		sb.WriteString("\"")
		sb.WriteString(col)
		sb.WriteString("\"")
	}
	sb.WriteString(" } [")

	rows := pt.GetRows()
	for _, row := range rows {
		for _, val := range row.Values {
			sb.WriteString(" ")
			ryeVal := ToRyeValue(val)
			if ryeVal != nil {
				sb.WriteString(ryeVal.Dump(e))
			} else {
				sb.WriteString("_")
			}
		}
		// Fill in missing columns (if they exist) with void (_)
		for i := len(row.Values); i < len(pt.Cols); i++ {
			sb.WriteString(" _")
		}
	}
	sb.WriteString(" ]")
	return sb.String()
}

// Trace prints trace information
func (pt *PersistentTable) Trace(msg string) {
	fmt.Print(msg + " (persistent-table): ")
}

// Get returns a row by index (for Collections interface)
func (pt *PersistentTable) Get(i int) Object {
	return pt.GetRow(nil, i)
}

// MakeNew creates a new collection (for Collections interface)
func (pt *PersistentTable) MakeNew(data []Object) Object {
	return *NewBlock(*NewTSeries(data))
}
