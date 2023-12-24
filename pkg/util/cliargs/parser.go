package cliargs

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
)

func ParseStruct[T any](
	structPtr *T,
	flagSetName string,
	flagSetErrorHandling flag.ErrorHandling,
	args []string,
) error {
	sVal := reflect.ValueOf(structPtr).Elem()
	if sVal.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %v", sVal.Kind())
	}

	flagSet := flag.NewFlagSet(flagSetName, flagSetErrorHandling)
	afterParseCallbacks := make(map[string]func(bool))
	sValType := sVal.Type()

	var flagArgsField reflect.Value
	for i := 0; i < sVal.NumField(); i++ {
		// parse `flag`, `usage` tags of field
		tags := sValType.Field(i).Tag
		flagName := tags.Get("flag")
		flagArgs := tags.Get("flagArgs")
		isFlagArgs := false
		if flagArgs != "" {
			var err error
			isFlagArgs, err = strconv.ParseBool(flagArgs)
			if err != nil {
				return fmt.Errorf("field %v has invalid 'flagArgs' tag value: %w", sValType.Field(i).Name, err)
			}
		}
		if flagName == "" && !isFlagArgs {
			continue
		}
		if isFlagArgs {
			flagArgsField = sVal.Field(i)
			if flagArgsField.Kind() != reflect.Slice && flagArgsField.Type().Elem().Kind() != reflect.String {
				return fmt.Errorf("field %v tagged as 'flagArgs' must be slice of strings", sValType.Field(i).Name)
			}
		} else {
			usage := tags.Get("usage")
			afterParseClb, err := createFlag(flagSet, sVal.Field(i), flagName, usage)
			if err != nil {
				return fmt.Errorf("field %v: %w", sValType.Field(i).Name, err)
			}
			afterParseCallbacks[flagName] = afterParseClb
		}
	}
	if err := flagSet.Parse(args); err != nil {
		return err
	}
	presentFlags := make(map[string]struct{})
	flagSet.Visit(func(f *flag.Flag) {
		presentFlags[f.Name] = struct{}{}
	})
	for name, clb := range afterParseCallbacks {
		_, isPresent := presentFlags[name]
		clb(isPresent)
	}
	if flagArgsField.IsValid() {
		flagArgsField.Set(reflect.ValueOf(flagSet.Args()))
	}
	return nil
}

func createFlag(
	flagSet *flag.FlagSet,
	value reflect.Value,
	name string,
	usage string,
) (afterParseClb func(isPresent bool), err error) {
	valType := value.Type()

	if value.Kind() == reflect.Ptr {
		value.Set(reflect.New(valType.Elem()))
		if err := createBuiltInTypeFlag(flagSet, value, name, usage); err != nil {
			return nil, err
		}

		return func(isPresent bool) {
			if !isPresent {
				value.Set(reflect.Zero(valType))
			}
		}, nil

	}
	newValue := reflect.New(valType)
	if err := createBuiltInTypeFlag(flagSet, newValue, name, usage); err != nil {
		return nil, err
	}
	return func(isPresent bool) {
		if isPresent {
			value.Set(newValue.Elem())
		}
	}, nil
}

// createFlag calls flagSet.IntVar() or similar for the given value depending on its type
func createBuiltInTypeFlag(
	flagSet *flag.FlagSet,
	valuePtr reflect.Value,
	name string,
	usage string,
) error {
	valueType := valuePtr.Type().Elem()
	switch valueType.Kind() {
	case reflect.Int:
		flagSet.IntVar((*int)(valuePtr.UnsafePointer()), name, 0, usage)
	case reflect.Uint:
		flagSet.UintVar((*uint)(valuePtr.UnsafePointer()), name, 0, usage)
	case reflect.Int64:
		flagSet.Int64Var((*int64)(valuePtr.UnsafePointer()), name, 0, usage)
	case reflect.Uint64:
		flagSet.Uint64Var((*uint64)(valuePtr.UnsafePointer()), name, 0, usage)
	case reflect.Float64:
		flagSet.Float64Var((*float64)(valuePtr.UnsafePointer()), name, 0, usage)
	case reflect.String:
		flagSet.StringVar((*string)(valuePtr.UnsafePointer()), name, "", usage)
	case reflect.Bool:
		flagSet.BoolVar((*bool)(valuePtr.UnsafePointer()), name, false, usage)
	default:
		return fmt.Errorf("unsupported type %v", valueType.Kind())
	}
	return nil
}
