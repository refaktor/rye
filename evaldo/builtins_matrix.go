package evaldo

import (
	"math"
	"math/rand"

	"github.com/drewlanenga/govector"
	"github.com/refaktor/rye/env"
)

var Builtins_matrix = map[string]*env.Builtin{

	//
	// ##### Matrix ##### "Matrix operations for 2D numerical data"
	//

	// Tests:
	// equal { matrix\zeros 3 4 |type? } 'matrix
	// equal { matrix\zeros 2 3 |rows? } 2
	// equal { matrix\zeros 2 3 |cols? } 3
	// Args:
	// * rows: number of rows
	// * cols: number of columns
	// Returns:
	// * a new zero-initialized matrix
	"matrix\\zeros": {
		Argsn: 2,
		Doc:   "Creates a new matrix with the given dimensions, initialized to zeros.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch rows := arg0.(type) {
			case env.Integer:
				switch cols := arg1.(type) {
				case env.Integer:
					if rows.Value <= 0 || cols.Value <= 0 {
						return MakeBuiltinError(ps, "Matrix dimensions must be positive", "matrix")
					}
					return *env.NewMatrix(int(rows.Value), int(cols.Value))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "matrix")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "matrix")
			}
		},
	},

	// Tests:
	// equal { matrix { 2 3 } { 1.0 2.0 3.0 4.0 5.0 6.0 } |mat-get 1 2 } 6.0
	// Args:
	// * shape: block with { rows cols }
	// * data: block of decimal values in row-major order
	// Returns:
	// * a new matrix with the given data
	"matrix": {
		Argsn: 2,
		Doc:   "Creates a matrix from a shape block and a data block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch shape := arg0.(type) {
			case env.Block:
				if shape.Series.Len() != 2 {
					return MakeBuiltinError(ps, "Shape must be { rows cols }", "matrix\\from")
				}
				rowsObj := shape.Series.Get(0)
				colsObj := shape.Series.Get(1)
				rowsInt, ok1 := rowsObj.(env.Integer)
				colsInt, ok2 := colsObj.(env.Integer)
				if !ok1 || !ok2 {
					return MakeBuiltinError(ps, "Shape must contain integers", "matrix\\from")
				}
				rows := int(rowsInt.Value)
				cols := int(colsInt.Value)

				switch data := arg1.(type) {
				case env.Block:
					if data.Series.Len() != rows*cols {
						return MakeBuiltinError(ps, "Data length must match rows*cols", "matrix\\from")
					}
					floats := make([]float64, rows*cols)
					for i := 0; i < data.Series.Len(); i++ {
						switch v := data.Series.Get(i).(type) {
						case env.Decimal:
							floats[i] = v.Value
						case env.Integer:
							floats[i] = float64(v.Value)
						default:
							return MakeBuiltinError(ps, "Data must contain numbers", "matrix\\from")
						}
					}
					return *env.NewMatrixWithData(rows, cols, floats)
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "matrix\\from")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "matrix\\from")
			}
		},
	},

	// Tests:
	// equal { matrix\randn 3 4 |rows? } 3
	// equal { matrix\randn 3 4 |cols? } 4
	// Args:
	// * rows: number of rows
	// * cols: number of columns
	// Returns:
	// * a new matrix with random normal values (mean 0, std 1)
	"matrix\\randn": {
		Argsn: 2,
		Doc:   "Creates a matrix with random normal values (mean 0, std 1).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch rows := arg0.(type) {
			case env.Integer:
				switch cols := arg1.(type) {
				case env.Integer:
					if rows.Value <= 0 || cols.Value <= 0 {
						return MakeBuiltinError(ps, "Matrix dimensions must be positive", "matrix\\randn")
					}
					r := int(rows.Value)
					c := int(cols.Value)
					data := make([]float64, r*c)
					for i := range data {
						data[i] = rand.NormFloat64()
					}
					return *env.NewMatrixWithData(r, c, data)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "matrix\\randn")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "matrix\\randn")
			}
		},
	},

	// Tests:
	// equal { matrix\ones 2 3 |mat-get 0 0 } 1.0
	// equal { matrix\ones 2 3 |mat-get 1 2 } 1.0
	// Args:
	// * rows: number of rows
	// * cols: number of columns
	// Returns:
	// * a new matrix filled with ones
	"matrix\\ones": {
		Argsn: 2,
		Doc:   "Creates a matrix filled with ones.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch rows := arg0.(type) {
			case env.Integer:
				switch cols := arg1.(type) {
				case env.Integer:
					if rows.Value <= 0 || cols.Value <= 0 {
						return MakeBuiltinError(ps, "Matrix dimensions must be positive", "matrix\\ones")
					}
					r := int(rows.Value)
					c := int(cols.Value)
					data := make([]float64, r*c)
					for i := range data {
						data[i] = 1.0
					}
					return *env.NewMatrixWithData(r, c, data)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "matrix\\ones")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "matrix\\ones")
			}
		},
	},

	// Tests:
	// equal { matrix\eye 3 |mat-get 0 0 } 1.0
	// equal { matrix\eye 3 |mat-get 0 1 } 0.0
	// equal { matrix\eye 3 |mat-get 1 1 } 1.0
	// Args:
	// * n: size of the identity matrix
	// Returns:
	// * an n×n identity matrix
	"matrix\\eye": {
		Argsn: 1,
		Doc:   "Creates an identity matrix of size n×n.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch n := arg0.(type) {
			case env.Integer:
				if n.Value <= 0 {
					return MakeBuiltinError(ps, "Matrix size must be positive", "matrix\\eye")
				}
				size := int(n.Value)
				data := make([]float64, size*size)
				for i := 0; i < size; i++ {
					data[i*size+i] = 1.0
				}
				return *env.NewMatrixWithData(size, size, data)
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "matrix\\eye")
			}
		},
	},

	//
	// ##### Matrix Access ##### "Functions to access matrix properties and elements"
	//

	// Tests:
	// equal { matrix 3 4 |shape? } { 3 4 }
	// Args:
	// * mat: matrix
	// Returns:
	// * block with { rows cols }
	"shape?": {
		Argsn: 1,
		Doc:   "Returns the shape of a matrix as { rows cols }.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch m := arg0.(type) {
			case env.Matrix:
				return *env.NewBlock(*env.NewTSeries([]env.Object{
					*env.NewInteger(int64(m.Rows)),
					*env.NewInteger(int64(m.Cols)),
				}))
			default:
				return MakeArgError(ps, 1, []env.Type{env.MatrixType}, "shape?")
			}
		},
	},

	// Tests:
	// equal { matrix 3 4 |rows? } 3
	// Args:
	// * mat: matrix
	// Returns:
	// * number of rows
	"rows?": {
		Argsn: 1,
		Doc:   "Returns the number of rows in a matrix.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch m := arg0.(type) {
			case env.Matrix:
				return *env.NewInteger(int64(m.Rows))
			default:
				return MakeArgError(ps, 1, []env.Type{env.MatrixType}, "rows?")
			}
		},
	},

	// Tests:
	// equal { matrix 3 4 |cols? } 4
	// Args:
	// * mat: matrix
	// Returns:
	// * number of columns
	"cols?": {
		Argsn: 1,
		Doc:   "Returns the number of columns in a matrix.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch m := arg0.(type) {
			case env.Matrix:
				return *env.NewInteger(int64(m.Cols))
			default:
				return MakeArgError(ps, 1, []env.Type{env.MatrixType}, "cols?")
			}
		},
	},

	// Tests:
	// equal { matrix\from { 2 3 } { 1.0 2.0 3.0 4.0 5.0 6.0 } |mat-get 0 1 } 2.0
	// equal { matrix\from { 2 3 } { 1.0 2.0 3.0 4.0 5.0 6.0 } |mat-get 1 0 } 4.0
	// Args:
	// * mat: matrix
	// * row: row index (0-based)
	// * col: column index (0-based)
	// Returns:
	// * the element at (row, col)
	"mat-get": {
		Argsn: 3,
		Doc:   "Gets the element at the specified row and column (0-indexed).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch m := arg0.(type) {
			case env.Matrix:
				switch row := arg1.(type) {
				case env.Integer:
					switch col := arg2.(type) {
					case env.Integer:
						r := int(row.Value)
						c := int(col.Value)
						if r < 0 || r >= m.Rows || c < 0 || c >= m.Cols {
							return MakeBuiltinError(ps, "Index out of bounds", "mat-get")
						}
						return *env.NewDecimal(m.Get(r, c))
					default:
						return MakeArgError(ps, 3, []env.Type{env.IntegerType}, "mat-get")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "mat-get")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.MatrixType}, "mat-get")
			}
		},
	},

	// Tests:
	// equal { m: matrix 2 2 , mat-set! m 0 1 5.0 , mat-get m 0 1 } 5.0
	// Args:
	// * mat: matrix (modified in place)
	// * row: row index (0-based)
	// * col: column index (0-based)
	// * val: value to set
	// Returns:
	// * the modified matrix
	"mat-set!": {
		Argsn: 4,
		Doc:   "Sets the element at the specified row and column (0-indexed). Modifies matrix in place.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch m := arg0.(type) {
			case env.Matrix:
				switch row := arg1.(type) {
				case env.Integer:
					switch col := arg2.(type) {
					case env.Integer:
						r := int(row.Value)
						c := int(col.Value)
						if r < 0 || r >= m.Rows || c < 0 || c >= m.Cols {
							return MakeBuiltinError(ps, "Index out of bounds", "mat-set!")
						}
						var val float64
						switch v := arg3.(type) {
						case env.Decimal:
							val = v.Value
						case env.Integer:
							val = float64(v.Value)
						default:
							return MakeArgError(ps, 4, []env.Type{env.DecimalType, env.IntegerType}, "mat-set!")
						}
						m.Set(r, c, val)
						return m
					default:
						return MakeArgError(ps, 3, []env.Type{env.IntegerType}, "mat-set!")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "mat-set!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.MatrixType}, "mat-set!")
			}
		},
	},

	// Tests:
	// equal { matrix\from { 2 3 } { 1.0 2.0 3.0 4.0 5.0 6.0 } |mat-row 0 |type? } 'vector
	// Args:
	// * mat: matrix
	// * row: row index (0-based)
	// Returns:
	// * the row as a vector
	"mat-row": {
		Argsn: 2,
		Doc:   "Returns a row of the matrix as a vector.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch m := arg0.(type) {
			case env.Matrix:
				switch row := arg1.(type) {
				case env.Integer:
					r := int(row.Value)
					if r < 0 || r >= m.Rows {
						return MakeBuiltinError(ps, "Row index out of bounds", "mat-row")
					}
					data := make(govector.Vector, m.Cols)
					for j := 0; j < m.Cols; j++ {
						data[j] = m.Get(r, j)
					}
					return *env.NewVector(data)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "mat-row")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.MatrixType}, "mat-row")
			}
		},
	},

	// Tests:
	// equal { matrix\from { 2 3 } { 1.0 2.0 3.0 4.0 5.0 6.0 } |mat-col 1 |type? } 'vector
	// Args:
	// * mat: matrix
	// * col: column index (0-based)
	// Returns:
	// * the column as a vector
	"mat-col": {
		Argsn: 2,
		Doc:   "Returns a column of the matrix as a vector.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch m := arg0.(type) {
			case env.Matrix:
				switch col := arg1.(type) {
				case env.Integer:
					c := int(col.Value)
					if c < 0 || c >= m.Cols {
						return MakeBuiltinError(ps, "Column index out of bounds", "mat-col")
					}
					data := make(govector.Vector, m.Rows)
					for i := 0; i < m.Rows; i++ {
						data[i] = m.Get(i, c)
					}
					return *env.NewVector(data)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "mat-col")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.MatrixType}, "mat-col")
			}
		},
	},

	//
	// ##### Matrix Operations ##### "Matrix transformation and arithmetic operations"
	//

	// Tests:
	// equal { matrix\from { 2 3 } { 1.0 2.0 3.0 4.0 5.0 6.0 } |mat-transpose |shape? } { 3 2 }
	// equal { matrix\from { 2 3 } { 1.0 2.0 3.0 4.0 5.0 6.0 } |mat-transpose |mat-get 0 1 } 4.0
	// Args:
	// * mat: matrix
	// Returns:
	// * transposed matrix
	"mat-transpose": {
		Argsn: 1,
		Doc:   "Returns the transpose of a matrix.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch m := arg0.(type) {
			case env.Matrix:
				data := make([]float64, m.Rows*m.Cols)
				for i := 0; i < m.Rows; i++ {
					for j := 0; j < m.Cols; j++ {
						// Transposed: new[j][i] = old[i][j]
						data[j*m.Rows+i] = m.Get(i, j)
					}
				}
				return *env.NewMatrixWithData(m.Cols, m.Rows, data)
			default:
				return MakeArgError(ps, 1, []env.Type{env.MatrixType}, "mat-transpose")
			}
		},
	},

	// Tests:
	// equal { A: matrix\from { 2 3 } { 1.0 2.0 3.0 4.0 5.0 6.0 } , B: matrix\from { 3 2 } { 1.0 2.0 3.0 4.0 5.0 6.0 } , mat-mul A B |shape? } { 2 2 }
	// equal { A: matrix\from { 2 3 } { 1.0 2.0 3.0 4.0 5.0 6.0 } , B: matrix\from { 3 2 } { 1.0 2.0 3.0 4.0 5.0 6.0 } , mat-mul A B |mat-get 0 0 } 22.0
	// Args:
	// * A: left matrix (m×n)
	// * B: right matrix (n×p)
	// Returns:
	// * result matrix (m×p)
	"mat-mul": {
		Argsn: 2,
		Doc:   "Matrix multiplication. A (m×n) × B (n×p) = C (m×p).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch A := arg0.(type) {
			case env.Matrix:
				switch B := arg1.(type) {
				case env.Matrix:
					if A.Cols != B.Rows {
						return MakeBuiltinError(ps, "Matrix dimensions incompatible for multiplication", "mat-mul")
					}
					m, n, p := A.Rows, A.Cols, B.Cols
					data := make([]float64, m*p)
					for i := 0; i < m; i++ {
						for j := 0; j < p; j++ {
							var sum float64
							for k := 0; k < n; k++ {
								sum += A.Get(i, k) * B.Get(k, j)
							}
							data[i*p+j] = sum
						}
					}
					return *env.NewMatrixWithData(m, p, data)
				case env.Vector:
					// Matrix × Vector = Vector
					if A.Cols != len(B.Value) {
						return MakeBuiltinError(ps, "Matrix columns must match vector length", "mat-mul")
					}
					result := make(govector.Vector, A.Rows)
					for i := 0; i < A.Rows; i++ {
						var sum float64
						for j := 0; j < A.Cols; j++ {
							sum += A.Get(i, j) * B.Value[j]
						}
						result[i] = sum
					}
					return *env.NewVector(result)
				default:
					return MakeArgError(ps, 2, []env.Type{env.MatrixType, env.VectorType}, "mat-mul")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.MatrixType}, "mat-mul")
			}
		},
	},

	// Note: In the future, element-wise multiplication could have a dedicated operator like .*
	// Tests:
	// equal { A: matrix\ones 2 2 , B: matrix\ones 2 2 , mat-hadamard A B |mat-get 0 0 } 1.0
	// Args:
	// * A: first matrix
	// * B: second matrix (same dimensions)
	// Returns:
	// * element-wise product
	"mat-hadamard": {
		Argsn: 2,
		Doc:   "Element-wise (Hadamard) multiplication of two matrices. In the future, could have a dedicated .* operator.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch A := arg0.(type) {
			case env.Matrix:
				switch B := arg1.(type) {
				case env.Matrix:
					if A.Rows != B.Rows || A.Cols != B.Cols {
						return MakeBuiltinError(ps, "Matrix dimensions must match", "mat-hadamard")
					}
					data := make([]float64, len(A.Data))
					for i := range A.Data {
						data[i] = A.Data[i] * B.Data[i]
					}
					return *env.NewMatrixWithData(A.Rows, A.Cols, data)
				default:
					return MakeArgError(ps, 2, []env.Type{env.MatrixType}, "mat-hadamard")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.MatrixType}, "mat-hadamard")
			}
		},
	},

	// Note: In the future, element-wise addition could have a dedicated operator like .+
	// Tests:
	// equal { A: matrix\ones 2 2 , B: matrix\ones 2 2 , mat-add A B |mat-get 0 0 } 2.0
	// Args:
	// * A: first matrix
	// * B: second matrix (same dimensions)
	// Returns:
	// * element-wise sum
	"mat-add": {
		Argsn: 2,
		Doc:   "Element-wise addition of two matrices. In the future, could have a dedicated .+ operator.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch A := arg0.(type) {
			case env.Matrix:
				switch B := arg1.(type) {
				case env.Matrix:
					if A.Rows != B.Rows || A.Cols != B.Cols {
						return MakeBuiltinError(ps, "Matrix dimensions must match", "mat-add")
					}
					data := make([]float64, len(A.Data))
					for i := range A.Data {
						data[i] = A.Data[i] + B.Data[i]
					}
					return *env.NewMatrixWithData(A.Rows, A.Cols, data)
				default:
					return MakeArgError(ps, 2, []env.Type{env.MatrixType}, "mat-add")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.MatrixType}, "mat-add")
			}
		},
	},

	// Note: In the future, element-wise subtraction could have a dedicated operator like .-
	// Tests:
	// equal { A: matrix\ones 2 2 , B: matrix\ones 2 2 , mat-sub A B |mat-get 0 0 } 0.0
	// Args:
	// * A: first matrix
	// * B: second matrix (same dimensions)
	// Returns:
	// * element-wise difference
	"mat-sub": {
		Argsn: 2,
		Doc:   "Element-wise subtraction of two matrices. In the future, could have a dedicated .- operator.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch A := arg0.(type) {
			case env.Matrix:
				switch B := arg1.(type) {
				case env.Matrix:
					if A.Rows != B.Rows || A.Cols != B.Cols {
						return MakeBuiltinError(ps, "Matrix dimensions must match", "mat-sub")
					}
					data := make([]float64, len(A.Data))
					for i := range A.Data {
						data[i] = A.Data[i] - B.Data[i]
					}
					return *env.NewMatrixWithData(A.Rows, A.Cols, data)
				default:
					return MakeArgError(ps, 2, []env.Type{env.MatrixType}, "mat-sub")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.MatrixType}, "mat-sub")
			}
		},
	},

	// Tests:
	// equal { matrix\ones 2 2 |mat-scale 3.0 |mat-get 0 0 } 3.0
	// Args:
	// * mat: matrix
	// * scalar: number to multiply by
	// Returns:
	// * scaled matrix
	"mat-scale": {
		Argsn: 2,
		Doc:   "Multiplies all elements of a matrix by a scalar.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch m := arg0.(type) {
			case env.Matrix:
				var scalar float64
				switch s := arg1.(type) {
				case env.Decimal:
					scalar = s.Value
				case env.Integer:
					scalar = float64(s.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.DecimalType, env.IntegerType}, "mat-scale")
				}
				data := make([]float64, len(m.Data))
				for i := range m.Data {
					data[i] = m.Data[i] * scalar
				}
				return *env.NewMatrixWithData(m.Rows, m.Cols, data)
			default:
				return MakeArgError(ps, 1, []env.Type{env.MatrixType}, "mat-scale")
			}
		},
	},

	//
	// ##### Reductions ##### "Aggregate matrix values"
	//

	// Tests:
	// equal { matrix\from { 2 3 } { 1.0 2.0 3.0 4.0 5.0 6.0 } |mat-sum-rows |type? } 'vector
	// Args:
	// * mat: matrix
	// Returns:
	// * vector with sum of each row
	"mat-sum-rows": {
		Argsn: 1,
		Doc:   "Returns a vector containing the sum of each row.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch m := arg0.(type) {
			case env.Matrix:
				result := make(govector.Vector, m.Rows)
				for i := 0; i < m.Rows; i++ {
					var sum float64
					for j := 0; j < m.Cols; j++ {
						sum += m.Get(i, j)
					}
					result[i] = sum
				}
				return *env.NewVector(result)
			default:
				return MakeArgError(ps, 1, []env.Type{env.MatrixType}, "mat-sum-rows")
			}
		},
	},

	// Tests:
	// equal { matrix\from { 2 3 } { 1.0 2.0 3.0 4.0 5.0 6.0 } |mat-sum-cols |type? } 'vector
	// Args:
	// * mat: matrix
	// Returns:
	// * vector with sum of each column
	"mat-sum-cols": {
		Argsn: 1,
		Doc:   "Returns a vector containing the sum of each column.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch m := arg0.(type) {
			case env.Matrix:
				result := make(govector.Vector, m.Cols)
				for j := 0; j < m.Cols; j++ {
					var sum float64
					for i := 0; i < m.Rows; i++ {
						sum += m.Get(i, j)
					}
					result[j] = sum
				}
				return *env.NewVector(result)
			default:
				return MakeArgError(ps, 1, []env.Type{env.MatrixType}, "mat-sum-cols")
			}
		},
	},



	// Tests:
	// equal { matrix\from { 2 3 } { 1.0 2.0 3.0 4.0 5.0 6.0 } |mat-max-rows |type? } 'vector
	// Args:
	// * mat: matrix
	// Returns:
	// * vector with max of each row
	"mat-max-rows": {
		Argsn: 1,
		Doc:   "Returns a vector containing the maximum of each row.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch m := arg0.(type) {
			case env.Matrix:
				result := make(govector.Vector, m.Rows)
				for i := 0; i < m.Rows; i++ {
					maxVal := m.Get(i, 0)
					for j := 1; j < m.Cols; j++ {
						if v := m.Get(i, j); v > maxVal {
							maxVal = v
						}
					}
					result[i] = maxVal
				}
				return *env.NewVector(result)
			default:
				return MakeArgError(ps, 1, []env.Type{env.MatrixType}, "mat-max-rows")
			}
		},
	},

	// Tests:
	// equal { matrix\from { 2 3 } { 1.0 2.0 3.0 4.0 5.0 6.0 } |mat-max-cols |type? } 'vector
	// Args:
	// * mat: matrix
	// Returns:
	// * vector with max of each column
	"mat-max-cols": {
		Argsn: 1,
		Doc:   "Returns a vector containing the maximum of each column.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch m := arg0.(type) {
			case env.Matrix:
				result := make(govector.Vector, m.Cols)
				for j := 0; j < m.Cols; j++ {
					maxVal := m.Get(0, j)
					for i := 1; i < m.Rows; i++ {
						if v := m.Get(i, j); v > maxVal {
							maxVal = v
						}
					}
					result[j] = maxVal
				}
				return *env.NewVector(result)
			default:
				return MakeArgError(ps, 1, []env.Type{env.MatrixType}, "mat-max-cols")
			}
		},
	},



	//
	// ##### Conversion ##### "Convert between matrix and other types"
	//

	// Tests:
	// equal { matrix\from { 2 2 } { 1.0 2.0 3.0 4.0 } |mat-to-block |length? } 2
	// Args:
	// * mat: matrix
	// Returns:
	// * block of blocks (each inner block is a row)
	"mat-to-block": {
		Argsn: 1,
		Doc:   "Converts a matrix to a block of blocks (each row is a block).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch m := arg0.(type) {
			case env.Matrix:
				rows := make([]env.Object, m.Rows)
				for i := 0; i < m.Rows; i++ {
					row := make([]env.Object, m.Cols)
					for j := 0; j < m.Cols; j++ {
						row[j] = *env.NewDecimal(m.Get(i, j))
					}
					rows[i] = *env.NewBlock(*env.NewTSeries(row))
				}
				return *env.NewBlock(*env.NewTSeries(rows))
			default:
				return MakeArgError(ps, 1, []env.Type{env.MatrixType}, "mat-to-block")
			}
		},
	},

	// Tests:
	// equal { { { 1.0 2.0 } { 3.0 4.0 } } |block-to-mat |shape? } { 2 2 }
	// Args:
	// * block: block of blocks (each inner block is a row)
	// Returns:
	// * matrix
	"block-to-mat": {
		Argsn: 1,
		Doc:   "Converts a block of blocks (rows) to a matrix.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch blk := arg0.(type) {
			case env.Block:
				if blk.Series.Len() == 0 {
					return MakeBuiltinError(ps, "Block is empty", "block-to-mat")
				}
				// Get first row to determine columns
				firstRow, ok := blk.Series.Get(0).(env.Block)
				if !ok {
					return MakeBuiltinError(ps, "Block must contain blocks (rows)", "block-to-mat")
				}
				rows := blk.Series.Len()
				cols := firstRow.Series.Len()
				data := make([]float64, rows*cols)

				for i := 0; i < rows; i++ {
					row, ok := blk.Series.Get(i).(env.Block)
					if !ok {
						return MakeBuiltinError(ps, "Block must contain blocks (rows)", "block-to-mat")
					}
					if row.Series.Len() != cols {
						return MakeBuiltinError(ps, "All rows must have the same length", "block-to-mat")
					}
					for j := 0; j < cols; j++ {
						switch v := row.Series.Get(j).(type) {
						case env.Decimal:
							data[i*cols+j] = v.Value
						case env.Integer:
							data[i*cols+j] = float64(v.Value)
						default:
							return MakeBuiltinError(ps, "Rows must contain numbers", "block-to-mat")
						}
					}
				}
				return *env.NewMatrixWithData(rows, cols, data)
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "block-to-mat")
			}
		},
	},

	// Tests:
	// equal { vector { 1.0 2.0 3.0 } |vector-to-mat 'row |shape? } { 1 3 }
	// equal { vector { 1.0 2.0 3.0 } |vector-to-mat 'col |shape? } { 3 1 }
	// Args:
	// * vec: vector
	// * orientation: 'row or 'col
	// Returns:
	// * matrix (1×n or n×1)
	"vector-to-mat": {
		Argsn: 2,
		Doc:   "Converts a vector to a matrix. Use 'row for 1×n or 'col for n×1.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch v := arg0.(type) {
			case env.Vector:
				switch orient := arg1.(type) {
				case env.Word:
					word := ps.Idx.GetWord(orient.Index)
					n := len(v.Value)
					data := make([]float64, n)
					copy(data, v.Value)
					if word == "row" {
						return *env.NewMatrixWithData(1, n, data)
					} else if word == "col" {
						return *env.NewMatrixWithData(n, 1, data)
					}
					return MakeBuiltinError(ps, "Orientation must be 'row or 'col", "vector-to-mat")
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "vector-to-mat")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.VectorType}, "vector-to-mat")
			}
		},
	},

	// Tests:
	// equal { matrix\from { 2 3 } { 1.0 2.0 3.0 4.0 5.0 6.0 } |mat-slice { 0 1 } { 1 2 } |shape? } { 2 2 }
	// Args:
	// * mat: matrix
	// * row-range: { start end } (inclusive)
	// * col-range: { start end } (inclusive)
	// Returns:
	// * sub-matrix
	"mat-slice": {
		Argsn: 3,
		Doc:   "Extracts a sub-matrix. Row and column ranges are { start end } (inclusive, 0-indexed).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch m := arg0.(type) {
			case env.Matrix:
				switch rowRange := arg1.(type) {
				case env.Block:
					if rowRange.Series.Len() != 2 {
						return MakeBuiltinError(ps, "Row range must be { start end }", "mat-slice")
					}
					switch colRange := arg2.(type) {
					case env.Block:
						if colRange.Series.Len() != 2 {
							return MakeBuiltinError(ps, "Column range must be { start end }", "mat-slice")
						}
						r0, ok1 := rowRange.Series.Get(0).(env.Integer)
						r1, ok2 := rowRange.Series.Get(1).(env.Integer)
						c0, ok3 := colRange.Series.Get(0).(env.Integer)
						c1, ok4 := colRange.Series.Get(1).(env.Integer)
						if !ok1 || !ok2 || !ok3 || !ok4 {
							return MakeBuiltinError(ps, "Ranges must contain integers", "mat-slice")
						}
						rowStart, rowEnd := int(r0.Value), int(r1.Value)
						colStart, colEnd := int(c0.Value), int(c1.Value)
						if rowStart < 0 || rowEnd >= m.Rows || rowStart > rowEnd {
							return MakeBuiltinError(ps, "Invalid row range", "mat-slice")
						}
						if colStart < 0 || colEnd >= m.Cols || colStart > colEnd {
							return MakeBuiltinError(ps, "Invalid column range", "mat-slice")
						}
						newRows := rowEnd - rowStart + 1
						newCols := colEnd - colStart + 1
						data := make([]float64, newRows*newCols)
						for i := 0; i < newRows; i++ {
							for j := 0; j < newCols; j++ {
								data[i*newCols+j] = m.Get(rowStart+i, colStart+j)
							}
						}
						return *env.NewMatrixWithData(newRows, newCols, data)
					default:
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "mat-slice")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "mat-slice")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.MatrixType}, "mat-slice")
			}
		},
	},

	// Tests:
	// equal { matrix { 2 2 } { 1.0 2.0 3.0 4.0 } |mat-softmax |mat-get 0 0 |> 0.1 } true
	// Args:
	// * matrix: input matrix
	// Returns:
	// * a new matrix with softmax applied to each column
	"mat-softmax": {
		Argsn: 1,
		Doc:   "Applies softmax to each column of the matrix. Returns a new matrix.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch m := arg0.(type) {
			case env.Matrix:
				data := make([]float64, m.Rows*m.Cols)
				for j := 0; j < m.Cols; j++ {
					// Find max in column for numerical stability
					maxVal := m.Get(0, j)
					for i := 1; i < m.Rows; i++ {
						if v := m.Get(i, j); v > maxVal {
							maxVal = v
						}
					}
					// Compute exp(x - max) and sum
					sum := 0.0
					for i := 0; i < m.Rows; i++ {
						exp := math.Exp(m.Get(i, j) - maxVal)
						data[i*m.Cols+j] = exp
						sum += exp
					}
					// Normalize
					for i := 0; i < m.Rows; i++ {
						data[i*m.Cols+j] /= sum
					}
				}
				return *env.NewMatrixWithData(m.Rows, m.Cols, data)
			default:
				return MakeArgError(ps, 1, []env.Type{env.MatrixType}, "mat-softmax")
			}
		},
	},

	// Tests:
	// equal { m: matrix\zeros 3 2 , mat-set-col! m 0 vector { 1.0 2.0 3.0 } , mat-get m 1 0 } 2.0
	// Args:
	// * matrix: matrix to modify
	// * col: column index (0-based)
	// * values: vector or block of values
	// Returns:
	// * the modified matrix
	"mat-set-col!": {
		Argsn: 3,
		Doc:   "Sets a column of the matrix from a vector or block. Mutates in place.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch m := arg0.(type) {
			case env.Matrix:
				switch col := arg1.(type) {
				case env.Integer:
					c := int(col.Value)
					if c < 0 || c >= m.Cols {
						return MakeBuiltinError(ps, "Column index out of bounds", "mat-set-col!")
					}
					switch vals := arg2.(type) {
					case env.Vector:
						if len(vals.Value) != m.Rows {
							return MakeBuiltinError(ps, "Vector length must match matrix rows", "mat-set-col!")
						}
						for i := 0; i < m.Rows; i++ {
							m.Set(i, c, vals.Value[i])
						}
						return m
					case env.Block:
						if vals.Series.Len() != m.Rows {
							return MakeBuiltinError(ps, "Block length must match matrix rows", "mat-set-col!")
						}
						for i := 0; i < m.Rows; i++ {
							switch v := vals.Series.Get(i).(type) {
							case env.Decimal:
								m.Set(i, c, v.Value)
							case env.Integer:
								m.Set(i, c, float64(v.Value))
							default:
								return MakeBuiltinError(ps, "Block must contain numbers", "mat-set-col!")
							}
						}
						return m
					default:
						return MakeArgError(ps, 3, []env.Type{env.VectorType, env.BlockType}, "mat-set-col!")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "mat-set-col!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.MatrixType}, "mat-set-col!")
			}
		},
	},

	// Tests:
	// equal { vec-add vector { 1.0 2.0 } vector { 3.0 4.0 } |first } 4.0
	// Args:
	// * v1: first vector
	// * v2: second vector
	// Returns:
	// * a new vector with element-wise sum
	"vec-add": {
		Argsn: 2,
		Doc:   "Adds two vectors element-wise. Returns a new vector.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch v1 := arg0.(type) {
			case env.Vector:
				switch v2 := arg1.(type) {
				case env.Vector:
					if len(v1.Value) != len(v2.Value) {
						return MakeBuiltinError(ps, "Vectors must have same length", "vec-add")
					}
					data := make(govector.Vector, len(v1.Value))
					for i := range v1.Value {
						data[i] = v1.Value[i] + v2.Value[i]
					}
					return *env.NewVector(data)
				default:
					return MakeArgError(ps, 2, []env.Type{env.VectorType}, "vec-add")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.VectorType}, "vec-add")
			}
		},
	},

	// Tests:
	// equal { vec-sub vector { 5.0 3.0 } vector { 1.0 2.0 } |first } 4.0
	// Args:
	// * v1: first vector
	// * v2: second vector
	// Returns:
	// * a new vector with element-wise difference (v1 - v2)
	"vec-sub": {
		Argsn: 2,
		Doc:   "Subtracts two vectors element-wise (v1 - v2). Returns a new vector.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch v1 := arg0.(type) {
			case env.Vector:
				switch v2 := arg1.(type) {
				case env.Vector:
					if len(v1.Value) != len(v2.Value) {
						return MakeBuiltinError(ps, "Vectors must have same length", "vec-sub")
					}
					data := make(govector.Vector, len(v1.Value))
					for i := range v1.Value {
						data[i] = v1.Value[i] - v2.Value[i]
					}
					return *env.NewVector(data)
				default:
					return MakeArgError(ps, 2, []env.Type{env.VectorType}, "vec-sub")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.VectorType}, "vec-sub")
			}
		},
	},

	// Tests:
	// equal { vec-to-block vector { 1.0 2.0 3.0 } |second } 2.0
	// Args:
	// * v: vector to convert
	// Returns:
	// * a block containing the vector values as decimals
	"vec-to-block": {
		Argsn: 1,
		Doc:   "Converts a vector to a block of decimal values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch v := arg0.(type) {
			case env.Vector:
				items := make([]env.Object, len(v.Value))
				for i, val := range v.Value {
					items[i] = *env.NewDecimal(val)
				}
				return *env.NewBlock(*env.NewTSeries(items))
			default:
				return MakeArgError(ps, 1, []env.Type{env.VectorType}, "vec-to-block")
			}
		},
	},
}
