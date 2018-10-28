package utils

import (
	"fmt"
	"os"
	"reflect"
)

func Merge(dest interface{}, src interface{}) error {
	vSrc := reflect.ValueOf(src)

	vDst := reflect.ValueOf(dest)
	if vDst.Kind() == reflect.Ptr {
		vDst = vDst.Elem()
	}
	return merge(vDst, vSrc)
}

func merge(dest reflect.Value, src reflect.Value) error {
	switch src.Kind() {
	case reflect.Func:
		if !dest.CanSet() {
			return nil
		}
		src = src.Call([]reflect.Value{})[0]
		if src.Kind() == reflect.Ptr {
			src = src.Elem()
		}
		if err := merge(dest, src); err != nil {
			return err
		}
	case reflect.Struct:
		// try to set the struct
		if src.Type() == dest.Type() {
			if !dest.CanSet() {
				return nil
			}

			dest.Set(src)
			return nil
		}

		for i := 0; i < src.NumMethod(); i++ {
			tMethod := src.Type().Method(i)

			df := dest.FieldByName(tMethod.Name)
			if df.Kind() == 0 {
				continue
			}

			if err := merge(df, src.Method(i)); err != nil {
				return err
			}
		}

		for i := 0; i < src.NumField(); i++ {
			tField := src.Type().Field(i)

			df := dest.FieldByName(tField.Name)
			if df.Kind() == 0 {
				continue
			}

			if err := merge(df, src.Field(i)); err != nil {
				return err
			}
		}

	case reflect.Map:
		x := reflect.MakeMap(dest.Type())
		for _, k := range src.MapKeys() {
			x.SetMapIndex(k, src.MapIndex(k))
		}
		dest.Set(x)
	case reflect.Slice:
		x := reflect.MakeSlice(dest.Type(), src.Len(), src.Len())
		for j := 0; j < src.Len(); j++ {
			merge(x.Index(j), src.Index(j))
		}
		dest.Set(x)
	case reflect.Chan:
	case reflect.Ptr:
		if !src.IsNil() && dest.CanSet() && src.Type() == dest.Type() {
			fmt.Fprintf(os.Stderr, "%#v %s\n", src.Type(), src.Type().Name())
			fmt.Fprintf(os.Stderr, "%#v %s\n", dest.Type(), dest.Type().Name())
			x := reflect.New(dest.Type().Elem())
			merge(x.Elem(), src.Elem())
			dest.Set(x)
		}
	default:
		if !dest.CanSet() {
			return nil
		}
		dest.Set(src)
	}

	return nil
}
