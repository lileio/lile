package protocopy

import (
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

func ToProto(src, dest interface{}) error {
	in := reflect.ValueOf(src)
	out := reflect.ValueOf(dest)

	if reflect.TypeOf(src) == reflect.TypeOf(reflect.Value{}) {
		in = src.(reflect.Value)
	}

	if reflect.TypeOf(dest) == reflect.TypeOf(reflect.Value{}) {
		out = dest.(reflect.Value)
	}

	return setValue(in, out)
}

func toProtoStruct(in, out reflect.Value) error {
	if out.IsNil() {
		out.Set(reflect.New(out.Type().Elem()))
	}

	if out.Type().String() == "*timestamppb.Timestamp" {
		if in.Kind() == reflect.Ptr {
			in = reflect.Indirect(in)
		}

		pbd, err := ptypes.TimestampProto(in.Interface().(time.Time))
		if err != nil {
			return err
		}

		out.Set(reflect.ValueOf(pbd))

		return nil
	}

FIELD_LOOP:
	for i := 0; i < in.NumField(); i++ {
		field := in.Type().Field(i)

		// Private fields are skipped
		if field.PkgPath != "" {
			continue FIELD_LOOP
		}

		inFieldName := field.Name

		pbTags := strings.Split(field.Tag.Get("pb"), ",")
		for _, v := range pbTags {
			if v == "ignore=true" {
				continue FIELD_LOOP
			}

			if v != "" {
				inFieldName = strings.TrimSpace(v)
			}
		}

		inField := in.Field(i)
		outField := out.Elem().FieldByNameFunc(func(n string) bool {
			return strings.EqualFold(strings.ToLower(n), strings.ToLower(inFieldName))
		})

		if !outField.IsValid() {
			return errors.New("No field found for " + inFieldName)
		}

		err := setValue(inField, outField)
		if err != nil {
			return err
		}
	}

	return nil
}

func setValue(inField, outField reflect.Value) error {
	k := inField.Type().Kind()
	if k == reflect.Ptr {
		if inField.IsZero() {
			return nil
		}

		k = inField.Elem().Type().Kind()
		inField = inField.Elem()
	}

	switch k {
	case reflect.Array, reflect.Slice:
		// Same types, so we can directly copy
		if inField.Type() == outField.Type() {
			outField.Set(inField)
			return nil
		}
		outField.Set(reflect.MakeSlice(outField.Type(), inField.Len(), inField.Len()))
		for i := 0; i < inField.Len(); i++ {
			err := setValue(inField.Index(i), outField.Index(i))
			if err != nil {
				return err
			}
		}
	case reflect.Struct:
		err := toProtoStruct(inField, outField)
		if err != nil {
			return err
		}
	case reflect.Map:
		outField.Set(reflect.MakeMap(outField.Type()))
		for _, key := range inField.MapKeys() {
			v := reflect.New(outField.Type().Elem()).Elem()
			err := setValue(inField.MapIndex(key), v)
			if err != nil {
				return err
			}

			outField.SetMapIndex(key, v)
		}
	case reflect.Chan, reflect.Func, reflect.Interface:
		return errors.New("input type not supported: " + inField.Type().Kind().String())
	default:
		err := setScalar(inField, outField)
		if err != nil {
			return err
		}
	}

	return nil

}

func setScalar(inF, outF reflect.Value) error {
	// We're trying to output to an ENUM
	if outF.MethodByName("Enum").IsValid() {
		// Deal with int enum and string
		if inF.Kind() == reflect.String {
			if inF.String() == "" {
				return nil
			}

			d := protoimpl.X.EnumDescriptorOf(outF.Interface())
			dv := d.Values().ByName(protoreflect.Name(inF.String()))
			outF.SetInt(int64(dv.Number()))
		} else {
			outF.SetInt(int64(inF.Int()))
		}

		return nil
	}

	if outF.Type().String() == "*durationpb.Duration" {
		pbd := ptypes.DurationProto(time.Duration(inF.Int()))
		outF.Set(reflect.ValueOf(pbd))

		return nil
	}

	outF.Set(inF)
	return nil
}
