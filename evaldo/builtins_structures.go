package evaldo

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/refaktor/rye/env"
)

// { <key> [ .print ] }
// { <key> { <more> [ .print ] } }
// { <key> { _ [ .print ] } }
// { <key> <token> [ .print ] }

// both kinds of blocks for a key
// { <person> [ .print ] { <name> print } }

// cpath for traversing deeper into the structure
// { people/author { <name> <surname> [ .print ] } }

// cpath for traversing deeper into the structure
// { people/author { <name> <surname> keyval [ .collect-kv ] } }

// { some { <person> k,v { [1] key , [2] val } } }

// { _ { <person> { * [ -> 1 |print , -> 2 |print ] } } }

func load_structures_Dict(ps *env.ProgramState, block env.Block) (env.Dict, *env.Error) {
	var keys []string

	data := make(map[string]any)
	rmap := *env.NewDict(data)

	for block.Series.Pos() < block.Series.Len() {
		obj := block.Series.Peek()
		switch obj1 := obj.(type) {
		case env.Xword:
			// trace5("TAG")
			keys = append(keys, ps.Idx.GetWord(obj1.Index))
			block.Series.Next()
			continue
		case env.Tagword:
			keys = append(keys, "-"+ps.Idx.GetWord(obj1.Index)+"-")
			block.Series.Next()
			continue
		case env.Void:
			keys = append(keys, "")
			block.Series.Next()
			continue
		case env.Block:
			// trace5("BLO")
			block.Series.Next()
			if obj1.Mode == 1 {
				// if code assign in to keys in Dict
				if len(keys) > 0 {
					for _, k := range keys {
						rmap.Data[k] = obj1
						keys = []string{}
					}
				} else {
					rmap.Data["-start-"] = obj1
				}
			} else if obj1.Mode == 0 {
				rm, err := load_saxml_Dict(ps, obj1)
				if err != nil {
					return _emptyRM(), err
				}
				if len(keys) > 0 {
					for _, k := range keys {
						rmap.Data[k] = rm
						keys = []string{}
					}
				} else {
					return _emptyRM(), MakeBuiltinError(ps, "No selectors before tag map.", "process")
				}
			}
		default:
			// ni Dict ampak blok kode, vrni blok
			return _emptyRM(), MakeBuiltinError(ps, "Unknow type in block parsing TODO.", "process")
		}
	}
	return rmap, nil
}

func do_structures(ps *env.ProgramState, data env.Dict, rmap env.Dict) env.Object { // TODO -- make it work for List too later
	fmt.Println(rmap)
	// fmt.Println("IN DO")
	//	var stack []env.Dict
	for key, val := range data.Data {
		// fmt.Println(key)
		rval0, ok0 := rmap.Data[""]
		if ok0 {
			// trace5("ANY FOUND")
			switch obj := rval0.(type) {
			case env.Dict:
				switch val1 := val.(type) {
				case map[string]any:
					// trace5("RECURSING")
					do_structures(ps, *env.NewDict(val1), obj)
					// trace5("OUTCURSING")
				}
			case env.Block:
				//				stack = append(stack, rmap)
				ser := ps.Ser // TODO -- make helper function that "does" a block
				ps.Ser = obj.Series
				EvalBlockInj(ps, env.ToRyeValue(val), true)
				if ps.ErrorFlag {
					ps.Ser = ser
					return ps.Res
				}
				ps.Ser = ser
			}
		}
		rval, ok := rmap.Data[key]
		if ok {
			// fmt.Println("found")
			switch obj := rval.(type) {
			case env.Dict:
				switch val1 := val.(type) {
				case map[string]any:
					do_structures(ps, *env.NewDict(val1), obj)
				}
			case env.Block:
				//				stack = append(stack, rmap)
				ser := ps.Ser // TODO -- make helper function that "does" a block
				ps.Ser = obj.Series
				EvalBlockInj(ps, env.ToRyeValue(val), true)
				if ps.ErrorFlag {
					ps.Ser = ser
					return ps.Res
				}
				ps.Ser = ser
			}
		}
	}
	return nil
}

// dictToStruct converts a Rye Dict to a Go struct using reflection
func dictToStruct(ps *env.ProgramState, dict env.Dict, structPtr any) env.Object {
	// Get the reflect.Value of the struct pointer
	structVal := reflect.ValueOf(structPtr)

	// Check if it's a pointer
	if structVal.Kind() != reflect.Ptr {
		return MakeBuiltinError(ps, "Second argument must be a pointer to a struct", "dict->struct")
	}

	// Get the struct value that the pointer points to
	structVal = structVal.Elem()

	// Check if it's a struct
	if structVal.Kind() != reflect.Struct {
		return MakeBuiltinError(ps, "Second argument must be a pointer to a struct", "dict->struct")
	}

	// Get the struct type
	structType := structVal.Type()

	// Iterate through the struct fields
	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Field(i)
		fieldType := structType.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Get the field name
		fieldName := fieldType.Name

		// Check for a tag that specifies the dict key
		tag := fieldType.Tag.Get("rye")
		if tag != "" {
			fieldName = tag
		}

		// Look for the field in the dict (try both original case and lowercase)
		dictValue, ok := dict.Data[fieldName]
		if !ok {
			// Try lowercase version of the field name
			dictValue, ok = dict.Data[strings.ToLower(fieldName)]
			if !ok {
				continue // Field not found in dict, skip it
			}
		}

		// Set the field value based on its type
		if err := setFieldValue(field, dictValue); err != nil {
			return MakeBuiltinError(ps, fmt.Sprintf("Error setting field %s: %s", fieldName, err.Error()), "dict->struct")
		}
	}

	// Return the native containing the struct pointer
	return *env.NewNative(ps.Idx, structPtr, "go-struct")
}

// setFieldValue sets a struct field value from a dict value
func setFieldValue(field reflect.Value, dictValue any) error {
	// Handle nil values
	if dictValue == nil {
		return nil // Skip nil values
	}

	// Handle Rye objects
	if ryeObj, ok := dictValue.(env.Object); ok {
		switch obj := ryeObj.(type) {
		case env.Integer:
			return setNumericField(field, float64(obj.Value))
		case env.Decimal:
			return setNumericField(field, obj.Value)
		case env.String:
			return setStringField(field, obj.Value)
		case env.Boolean:
			return setBoolField(field, obj.Value)
		case env.Dict:
			// If the field is a struct, recursively set its fields
			if field.Kind() == reflect.Struct {
				for k, v := range obj.Data {
					// Find the field in the struct
					structField := field.FieldByName(k)
					if structField.IsValid() && structField.CanSet() {
						if err := setFieldValue(structField, v); err != nil {
							return err
						}
					}
				}
				return nil
			}
			return fmt.Errorf("cannot set Dict to non-struct field")
		case env.List:
			// Handle list to slice/array conversion
			return setSliceField(field, obj.Data)
		default:
			return fmt.Errorf("unsupported Rye type: %T", obj)
		}
	}

	// Handle Go types
	switch value := dictValue.(type) {
	case string:
		return setStringField(field, value)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return setNumericField(field, reflect.ValueOf(value).Float())
	case bool:
		return setBoolField(field, value)
	case []any:
		return setSliceField(field, value)
	case map[string]any:
		// If the field is a struct, recursively set its fields
		if field.Kind() == reflect.Struct {
			for k, v := range value {
				// Find the field in the struct
				structField := field.FieldByName(k)
				if structField.IsValid() && structField.CanSet() {
					if err := setFieldValue(structField, v); err != nil {
						return err
					}
				}
			}
			return nil
		}
		return fmt.Errorf("cannot set map to non-struct field")
	default:
		return fmt.Errorf("unsupported Go type: %T", value)
	}
}

// setNumericField sets a numeric field value
func setNumericField(field reflect.Value, value float64) error {
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		field.SetInt(int64(value))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		field.SetUint(uint64(value))
	case reflect.Float32, reflect.Float64:
		field.SetFloat(value)
	default:
		return fmt.Errorf("cannot set numeric value to %s field", field.Kind())
	}
	return nil
}

// setStringField sets a string field value
func setStringField(field reflect.Value, value string) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("cannot set string value to %s field", field.Kind())
	}
	field.SetString(value)
	return nil
}

// setBoolField sets a boolean field value
func setBoolField(field reflect.Value, value bool) error {
	if field.Kind() != reflect.Bool {
		return fmt.Errorf("cannot set boolean value to %s field", field.Kind())
	}
	field.SetBool(value)
	return nil
}

// setSliceField sets a slice field value
func setSliceField(field reflect.Value, value []any) error {
	if field.Kind() != reflect.Slice && field.Kind() != reflect.Array {
		return fmt.Errorf("cannot set slice value to %s field", field.Kind())
	}

	// Create a new slice of the appropriate type
	sliceType := field.Type()
	newSlice := reflect.MakeSlice(sliceType, len(value), len(value))

	// Set each element in the slice
	for i, v := range value {
		elemValue := newSlice.Index(i)
		if err := setFieldValue(elemValue, v); err != nil {
			return err
		}
	}

	// Set the field to the new slice
	field.Set(newSlice)
	return nil
}

// structToDict converts a Go struct to a Rye Dict using reflection
func structToDict(ps *env.ProgramState, structPtr any) env.Object {
	// Get the reflect.Value of the struct pointer
	structVal := reflect.ValueOf(structPtr)

	// Check if it's a pointer
	if structVal.Kind() == reflect.Ptr {
		// Get the struct value that the pointer points to
		structVal = structVal.Elem()
	}

	// Check if it's a struct
	if structVal.Kind() != reflect.Struct {
		return MakeBuiltinError(ps, "Argument must be a struct or pointer to a struct", "struct->dict")
	}

	// Create a new Dict
	data := make(map[string]any)
	dict := env.NewDict(data)

	// Get the struct type
	structType := structVal.Type()

	// Iterate through the struct fields
	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Field(i)
		fieldType := structType.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Get the field name
		fieldName := fieldType.Name

		// Check for a tag that specifies the dict key
		tag := fieldType.Tag.Get("rye")
		if tag != "" {
			fieldName = tag
		}

		// Convert the field value to a Rye value
		ryeValue := fieldToRyeValue(ps, field)

		// Add the field to the Dict
		dict.Data[fieldName] = ryeValue
	}

	return *dict
}

// fieldToRyeValue converts a reflect.Value to a Rye value
func fieldToRyeValue(ps *env.ProgramState, field reflect.Value) env.Object {
	// Handle nil values
	if !field.IsValid() || (field.Kind() == reflect.Ptr && field.IsNil()) {
		return env.Void{}
	}

	// If it's a pointer, get the value it points to
	if field.Kind() == reflect.Ptr {
		return fieldToRyeValue(ps, field.Elem())
	}

	// Convert based on the field type
	switch field.Kind() {
	case reflect.Bool:
		return *env.NewBoolean(field.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return *env.NewInteger(field.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return *env.NewInteger(int64(field.Uint()))
	case reflect.Float32, reflect.Float64:
		return *env.NewDecimal(field.Float())
	case reflect.String:
		return *env.NewString(field.String())
	case reflect.Struct:
		return structToDict(ps, field.Interface())
	case reflect.Slice, reflect.Array:
		// Create a new List with empty data
		data := make([]any, 0, field.Len())
		list := env.NewList(data)

		// Add each element to the List
		for i := 0; i < field.Len(); i++ {
			elem := field.Index(i)
			ryeValue := fieldToRyeValue(ps, elem)
			list.Data = append(list.Data, ryeValue)
		}

		return *list
	case reflect.Map:
		// Create a new Dict
		data := make(map[string]any)
		dict := env.NewDict(data)

		// Add each key-value pair to the Dict
		for _, key := range field.MapKeys() {
			// Only support string keys
			if key.Kind() == reflect.String {
				value := field.MapIndex(key)
				ryeValue := fieldToRyeValue(ps, value)
				dict.Data[key.String()] = ryeValue
			}
		}

		return *dict
	default:
		// For unsupported types, return a string representation
		return *env.NewString(fmt.Sprintf("%v", field.Interface()))
	}
}

// EmptyRM creates an empty Dict
func EmptyRM() env.Dict {
	data := make(map[string]any)
	return *env.NewDict(data)
}

// _emptyRM is kept for backward compatibility
func _emptyRM() env.Dict {
	return EmptyRM()
}

// createStructType creates a new struct type at runtime based on a Rye Dict
func createStructType(ps *env.ProgramState, dict env.Dict, structName string) (reflect.Type, error) {
	// Create a map of field definitions
	fields := make([]reflect.StructField, 0, len(dict.Data))

	// We don't need to specify a package path for anonymous structs

	// Process each key-value pair in the Dict
	for key, value := range dict.Data {
		// Skip keys that are not valid Go identifiers
		if !isValidGoIdentifier(key) {
			continue
		}

		// Determine the field type based on the value
		var fieldType reflect.Type
		switch v := value.(type) {
		case env.Integer:
			fieldType = reflect.TypeOf(int64(0))
		case env.Decimal:
			fieldType = reflect.TypeOf(float64(0))
		case env.String:
			fieldType = reflect.TypeOf("")
		case env.Boolean:
			fieldType = reflect.TypeOf(false)
		case env.Dict:
			// For nested Dicts, recursively create a new struct type
			nestedType, err := createStructType(ps, v, key+"Struct")
			if err != nil {
				return nil, err
			}
			fieldType = nestedType
		case env.List:
			// For Lists, create a slice type based on the first element's type
			if len(v.Data) > 0 {
				var elemType reflect.Type
				// Determine element type based on the first element
				firstElem := v.Data[0]
				if _, ok := firstElem.(env.Integer); ok {
					elemType = reflect.TypeOf(int64(0))
				} else if _, ok := firstElem.(env.Decimal); ok {
					elemType = reflect.TypeOf(float64(0))
				} else if _, ok := firstElem.(env.String); ok {
					elemType = reflect.TypeOf("")
				} else if _, ok := firstElem.(env.Boolean); ok {
					elemType = reflect.TypeOf(false)
				} else {
					// Default to interface{} for complex types
					elemType = reflect.TypeOf((*any)(nil)).Elem()
				}
				fieldType = reflect.SliceOf(elemType)
			} else {
				// Empty list, default to []interface{}
				fieldType = reflect.SliceOf(reflect.TypeOf((*any)(nil)).Elem())
			}
		default:
			// Default to interface{} for unsupported types
			fieldType = reflect.TypeOf((*any)(nil)).Elem()
		}

		// Create a struct field
		field := reflect.StructField{
			Name: capitalize(key), // Ensure the field name starts with an uppercase letter
			Type: fieldType,
			Tag:  reflect.StructTag(fmt.Sprintf(`rye:"%s"`, key)),
		}

		fields = append(fields, field)
	}

	// Create a new struct type
	structType := reflect.StructOf(fields)
	return structType, nil
}

// isValidGoIdentifier checks if a string is a valid Go identifier
func isValidGoIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}

	// First character must be a letter or underscore
	if !unicode.IsLetter(rune(s[0])) && s[0] != '_' {
		return false
	}

	// Remaining characters must be letters, digits, or underscores
	for _, c := range s[1:] {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '_' {
			return false
		}
	}

	return true
}

// capitalize returns a string with the first letter capitalized
func capitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// populateDynamicStruct populates a dynamically created struct with values from a Rye Dict
func populateDynamicStruct(ps *env.ProgramState, dict env.Dict, structValue reflect.Value) error {
	// Get the struct type
	structType := structValue.Type()

	// Iterate through the struct fields
	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		fieldType := structType.Field(i)

		// Get the field name from the tag
		tag := fieldType.Tag.Get("rye")

		// Get the value from the Dict
		dictValue, ok := dict.Data[tag]
		if !ok {
			continue // Skip fields not found in the Dict
		}

		// Set the field value based on its type
		switch field.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if intVal, ok := dictValue.(env.Integer); ok {
				field.SetInt(intVal.Value)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if intVal, ok := dictValue.(env.Integer); ok {
				field.SetUint(uint64(intVal.Value))
			}
		case reflect.Float32, reflect.Float64:
			if floatVal, ok := dictValue.(env.Decimal); ok {
				field.SetFloat(floatVal.Value)
			}
		case reflect.String:
			if strVal, ok := dictValue.(env.String); ok {
				field.SetString(strVal.Value)
			}
		case reflect.Bool:
			if boolVal, ok := dictValue.(env.Boolean); ok {
				field.SetBool(boolVal.Value)
			}
		case reflect.Struct:
			if dictVal, ok := dictValue.(env.Dict); ok {
				if err := populateDynamicStruct(ps, dictVal, field); err != nil {
					return err
				}
			}
		case reflect.Slice:
			if listVal, ok := dictValue.(env.List); ok {
				// Create a new slice
				sliceType := field.Type()
				elemType := sliceType.Elem()
				slice := reflect.MakeSlice(sliceType, len(listVal.Data), len(listVal.Data))

				// Populate the slice
				for j, elem := range listVal.Data {
					elemValue := slice.Index(j)

					switch elemType.Kind() {
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						if intVal, ok := elem.(env.Integer); ok {
							elemValue.SetInt(intVal.Value)
						}
					case reflect.Float32, reflect.Float64:
						if floatVal, ok := elem.(env.Decimal); ok {
							elemValue.SetFloat(floatVal.Value)
						}
					case reflect.String:
						if strVal, ok := elem.(env.String); ok {
							elemValue.SetString(strVal.Value)
						}
					case reflect.Bool:
						if boolVal, ok := elem.(env.Boolean); ok {
							elemValue.SetBool(boolVal.Value)
						}
					}
				}

				// Set the field to the new slice
				field.Set(slice)
			}
		}
	}

	return nil
}

// dictToNewStruct creates a new struct type at runtime and populates it with values from a Rye Dict
func dictToNewStruct(ps *env.ProgramState, dict env.Dict, structName string) (any, error) {
	// Create a new struct type
	structType, err := createStructType(ps, dict, structName)
	if err != nil {
		return nil, err
	}

	// Create a new instance of the struct
	structPtr := reflect.New(structType).Interface()

	// Populate the struct with values from the Dict
	if err := populateDynamicStruct(ps, dict, reflect.ValueOf(structPtr).Elem()); err != nil {
		return nil, err
	}

	return structPtr, nil
}

var Builtins_structures = map[string]*env.Builtin{

	"dict->struct": {
		Argsn: 2,
		Doc:   "Converts a Rye Dict to a Go struct using reflection. Takes a Dict and a Native containing a pointer to a struct.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Check if the first argument is a Dict
			switch dict := arg0.(type) {
			case env.Dict:
				// Check if the second argument is a Native
				switch native := arg1.(type) {
				case env.Native:
					// Convert the Dict to a struct
					return dictToStruct(ps, dict, native.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.NativeType}, "dict->struct")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.DictType}, "dict->struct")
			}
		},
	},

	"dict->new-struct": {
		Argsn: 2,
		Doc:   "Creates a new Go struct type at runtime based on a Rye Dict and populates it with values. Takes a Dict and a string for the struct name.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Check if the first argument is a Dict
			switch dict := arg0.(type) {
			case env.Dict:
				// Check if the second argument is a String
				switch nameObj := arg1.(type) {
				case env.String:
					// Create a new struct type and populate it
					structPtr, err := dictToNewStruct(ps, dict, nameObj.Value)
					if err != nil {
						return MakeBuiltinError(ps, err.Error(), "dict->new-struct")
					}

					// Return the new struct as a Native
					return *env.NewNative(ps.Idx, structPtr, nameObj.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "dict->new-struct")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.DictType}, "dict->new-struct")
			}
		},
	},

	"struct->dict": {
		Argsn: 1,
		Doc:   "Converts a Go struct to a Rye Dict using reflection. Takes a Native containing a struct or pointer to a struct.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Check if the argument is a Native
			switch native := arg0.(type) {
			case env.Native:
				// Convert the struct to a Dict
				return structToDict(ps, native.Value)
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "struct->dict")
			}
		},
	},

	"process": {
		Argsn: 2,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			rm, err := load_structures_Dict(ps, arg1.(env.Block))
			//fmt.Println(rm)
			if err != nil {
				ps.FailureFlag = true
				return err
			}
			switch data := arg0.(type) {
			case env.Dict:
				return do_structures(ps, data, rm)
			}
			return nil
		},
	},
}
