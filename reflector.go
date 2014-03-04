package skini

/*
Reflector -- uses reflection to add values
to target structure fields.
*/

import (
	"fmt"
    "reflect"
)

//------------------------------------------------------------
// Reflection
//------------------------------------------------------------

// Gets element pointed at by interface
func getElem(target interface{}) (elem reflect.Value, err error) {
    // Target must be a pointer
    if reflect.TypeOf(target).Kind() != reflect.Ptr {
		err = fmt.Errorf("error, must be pointer to target, not object itself")
        return
    }

    // Get element
    elem = reflect.ValueOf(target).Elem()
    return
}

//------------------------------------------------------------
// Field getters
//------------------------------------------------------------

// Gets field from the root level.
func getField(target interface{}, name string) (value string, err error) {
    // Target must be a pointer
    if reflect.TypeOf(target).Kind() != reflect.Ptr {
		err = fmt.Errorf("error, must be pointer to target, not object itself")
        return
    }

    // Get element
    elem := reflect.ValueOf(target).Elem()
    value = elem.FieldByName(name).String()
    return
}

//------------------------------------------------------------
// Field setters
//------------------------------------------------------------

// Sets plain field.
func setField(elem *reflect.Value, section, key, value string) (err error) {
    //fmt.Printf("\t[%s] SET FIELD: %s = %s\n", section, key, value)

    section = toFieldName(section)
    key = toFieldName(key)

    field, err := findField(elem, section, key)
    if err != nil {
        return err
    }

    if err = isFieldModifiable(field, key, reflect.String); err != nil {
        return err
    }

    field.SetString(value)
    return
}

// Adds item to a slice.
func addSliceItem(elem *reflect.Value, section, key, value string) (err error) {
    //fmt.Printf("\tADD SLICE ITEM: [%s] %s %s\n", section, key, value)

    section = toFieldName(section)
    key = toFieldName(key)

    field, err := findField(elem, section, key)
    if err != nil {
        return err
    }

    if err = isFieldModifiable(field, key, reflect.Slice); err != nil {
        return
    }

    // First on consecutive add ?
    if field.IsNil() {
        // First add
        field.Set(reflect.MakeSlice(field.Type(), 1, 1))
        field.Index(0).Set(//reflect.ValueOf(value))
            reflect.ValueOf(value).Convert(field.Type().Elem()))
    } else {
        // Consecutive adds copy slice and increase its size by 1
        l := field.Len()
        fieldNew := reflect.MakeSlice(field.Type(), l + 1, l + 1)
        for i := 0; i < l; i++ {
            fieldNew.Index(i).Set(field.Index(i))
        }
        fieldNew.Index(l).Set(
            reflect.ValueOf(value).Convert(field.Type().Elem()))
        field.Set(fieldNew)
    }

    return
}

// Adds item to a map. 
func addMapItem(elem *reflect.Value, topmap, submap, key, value string) (err error) {
    //fmt.Printf("\t\t    + ADD MAP ITEM: [%s | %s] : %s = %s\n", topmap, submap, key, value)

    topmap = toFieldName(topmap)

    field, err := findMap(elem, topmap)
    if err != nil {
        return err
    }

    if err = isFieldModifiable(field, topmap, reflect.Map); err != nil {
        return
    }

    // First on consecutive add ?
    if field.IsNil() {
        field.Set(reflect.MakeMap(field.Type()))
    }

    // No submap ?
    if submap == "" {
        field.SetMapIndex(reflect.ValueOf(key), 
            reflect.ValueOf(value).Convert(field.Type().Elem()))
        return
    }

    // Lookup submap as a value in top map
    subkeyval := reflect.ValueOf(submap)
    subfield := field.MapIndex(subkeyval)

    // First time add
    if ! subfield.IsValid() {
        submapval := reflect.MakeMap(reflect.MapOf(
            reflect.TypeOf(key), 
            field.Type().Elem().Elem()))

        field.SetMapIndex(subkeyval, submapval)
        submapval.SetMapIndex(reflect.ValueOf(key), 
            reflect.ValueOf(value).Convert(submapval.Type().Elem()))
        return
    }

    subfield.SetMapIndex(reflect.ValueOf(key), 
            reflect.ValueOf(value).Convert(subfield.Type().Elem()))
    return
}

//------------------------------------------------------------
// Field search functions
//------------------------------------------------------------

// Finds field by given section name and key.
//
func findField(elem *reflect.Value, section, key string) (field *reflect.Value, err error) {
    var f reflect.Value
    if section == "" {
        // Get root section element
        f = elem.FieldByName(key)
        if !f.IsValid() {
            err = fmt.Errorf("struct doesn't have field: %s", key)
        }
    } else {
        // Get inner struct element
        f = elem.FieldByName(section)
        if !f.IsValid() {
            err = fmt.Errorf("struct doesn't have nested struct: %s", section)
            return
        }
        f = f.FieldByName(key)
        if !f.IsValid() {
            err = fmt.Errorf("struct doesn't have nested struct field: %s.%s", section, key)
        }
    }
    return &f, err
}

// Finds map field that can be inside another map.
func findMap(elem *reflect.Value, name string) (field *reflect.Value, err error) {
    var f reflect.Value
    
    f = elem.FieldByName(name)
    if !f.IsValid() {
        err = fmt.Errorf("struct doesn't have map field: %s", name)
    }
    return &f, err
}

// Checks if field can be modified and 
// field kind matches give kind.
func isFieldModifiable(field *reflect.Value, name string, kind reflect.Kind) (err error) {
    if !field.IsValid() {
        return fmt.Errorf("error, field not valid: %s", name)
    }
    if !field.CanSet() {
        return fmt.Errorf("error, field cannot be set: %s", name)
    }

    if field.Kind() != kind {
        switch kind {
        case reflect.String:
            return fmt.Errorf("error, field must be string: %s", name)

        case reflect.Slice:
            return fmt.Errorf("error, field must be slice: %s", name)

        case reflect.Map:
            return fmt.Errorf("error, field must be map: %s", name)

        default:
            return fmt.Errorf("error, not yet supported type for field: %s", name)
        }
    }
    return
}

