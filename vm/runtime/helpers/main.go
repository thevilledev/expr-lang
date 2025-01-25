package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"strings"
	"text/template"
)

func main() {
	var b bytes.Buffer
	err := template.Must(
		template.New("helpers").
			Funcs(template.FuncMap{
				"cases":          func(op string) string { return cases(op, uints, ints, floats) },
				"cases_int_only": func(op string) string { return cases(op, uints, ints) },
				"cases_with_duration": func(op string) string {
					return cases(op, uints, ints, floats, []string{"time.Duration"})
				},
				"array_equal_cases": func() string { return arrayEqualCases([]string{"string"}, uints, ints, floats) },
			}).
			Parse(helpers),
	).Execute(&b, nil)
	if err != nil {
		panic(err)
	}

	formatted, err := format.Source(b.Bytes())
	if err != nil {
		panic(err)
	}
	fmt.Print(string(formatted))
}

var ints = []string{
	"int",
	"int8",
	"int16",
	"int32",
	"int64",
}

var uints = []string{
	"uint",
	"uint8",
	"uint16",
	"uint32",
	"uint64",
}

var floats = []string{
	"float32",
	"float64",
}

func cases(op string, xs ...[]string) string {
	var types []string
	for _, x := range xs {
		types = append(types, x...)
	}

	_, _ = fmt.Fprintf(os.Stderr, "Generating %s cases for %v\n", op, types)

	var out string
	echo := func(s string, xs ...any) {
		out += fmt.Sprintf(s, xs...) + "\n"
	}

	// Only generate pointer cases for equality operations
	if op == "==" {
		for _, a := range types {
			// Handle *T cases
			echo(`case *%v:`, a)
			echo(`switch y := b.(type) {`)
			echo(`case %v:`, a)
			echo(`return *x == y`)
			echo(`case *%v:`, a)
			echo(`return *x == *y`)
			echo(`}`)
		}
	}

	// Generate regular cases
	for _, a := range types {
		echo(`case %v:`, a)
		echo(`switch y := b.(type) {`)

		// Add pointer case for equality
		if op == "==" {
			echo(`case *%v:`, a)
			echo(`return x == *y`)
		}

		// Add regular type cases
		for _, b := range types {
			t := "int"
			if isDuration(a) || isDuration(b) {
				t = "time.Duration"
			}
			if isFloat(a) || isFloat(b) {
				t = "float64"
			}
			echo(`case %v:`, b)
			if op == "/" {
				echo(`return float64(x) / float64(y)`)
			} else {
				echo(`return %v(x) %v %v(y)`, t, op, t)
			}
		}
		echo(`}`)
	}
	return strings.TrimRight(out, "\n")
}

func arrayEqualCases(xs ...[]string) string {
	var types []string
	for _, x := range xs {
		types = append(types, x...)
	}

	_, _ = fmt.Fprintf(os.Stderr, "Generating array equal cases for %v\n", types)

	var out string
	echo := func(s string, xs ...any) {
		out += fmt.Sprintf(s, xs...) + "\n"
	}
	echo(`case []any:`)
	echo(`switch y := b.(type) {`)
	for _, a := range append(types, "any") {
		echo(`case []%v:`, a)
		echo(`if len(x) != len(y) { return false }`)
		echo(`for i := range x {`)
		echo(`if !Equal(x[i], y[i]) { return false }`)
		echo(`}`)
		echo("return true")
	}
	echo(`}`)
	for _, a := range types {
		echo(`case []%v:`, a)
		echo(`switch y := b.(type) {`)
		echo(`case []any:`)
		echo(`return Equal(y, x)`)
		echo(`case []%v:`, a)
		echo(`if len(x) != len(y) { return false }`)
		echo(`for i := range x {`)
		echo(`if x[i] != y[i] { return false }`)
		echo(`}`)
		echo("return true")
		echo(`}`)
	}
	return strings.TrimRight(out, "\n")
}

func isFloat(t string) bool {
	return strings.HasPrefix(t, "float")
}

func isDuration(t string) bool {
	return t == "time.Duration"
}

const helpers = `// Code generated by vm/runtime/helpers/main.go. DO NOT EDIT.

package runtime

import (
	"fmt"
	"reflect"
	"time"
)

func Equal(a, b interface{}) bool {
	switch x := a.(type) {
	{{ cases "==" }}
	{{ array_equal_cases }}
	case string:
		switch y := b.(type) {
		case string:
			return x == y
		}
	case time.Time:
		switch y := b.(type) {
		case time.Time:
			return x.Equal(y)
		}
	case time.Duration:
		switch y := b.(type) {
		case time.Duration:
			return x == y
		}
	case bool:
		switch y := b.(type) {
		case bool:
			return x == y
		}
	}

	// Handle unknown pointer types and custom types
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	// Handle pointers to unknown types
	if va.Kind() == reflect.Ptr {
		if !va.IsValid() || va.IsNil() {
			return vb.Kind() == reflect.Ptr && (!vb.IsValid() || vb.IsNil())
		}
		va = va.Elem()
	}
	if vb.Kind() == reflect.Ptr {
		if !vb.IsValid() || vb.IsNil() {
			return va.Kind() == reflect.Ptr && (!va.IsValid() || va.IsNil())
		}
		vb = vb.Elem()
	}

	// Handle custom integer types by converting to int64
	if va.IsValid() && vb.IsValid() {
		ka := va.Kind()
		kb := vb.Kind()
		if (ka == reflect.Int || ka == reflect.Int8 || ka == reflect.Int16 || ka == reflect.Int32 || ka == reflect.Int64) &&
		   (kb == reflect.Int || kb == reflect.Int8 || kb == reflect.Int16 || kb == reflect.Int32 || kb == reflect.Int64) {
			return va.Int() == vb.Int()
		}
	}

	if IsNil(a) && IsNil(b) {
		return true
	}
	return reflect.DeepEqual(a, b)
}

func Less(a, b interface{}) bool {
	switch x := a.(type) {
	{{ cases "<" }}
	case string:
		switch y := b.(type) {
		case string:
			return x < y
		}
	case time.Time:
		switch y := b.(type) {
		case time.Time:
			return x.Before(y)
		}
	case time.Duration:
		switch y := b.(type) {
		case time.Duration:
			return x < y
		}
	}
	panic(fmt.Sprintf("invalid operation: %T < %T", a, b))
}

func More(a, b interface{}) bool {
	switch x := a.(type) {
	{{ cases ">" }}
	case string:
		switch y := b.(type) {
		case string:
			return x > y
		}
	case time.Time:
		switch y := b.(type) {
		case time.Time:
			return x.After(y)
		}
	case time.Duration:
		switch y := b.(type) {
		case time.Duration:
			return x > y
		}
	}
	panic(fmt.Sprintf("invalid operation: %T > %T", a, b))
}

func LessOrEqual(a, b interface{}) bool {
	switch x := a.(type) {
	{{ cases "<=" }}
	case string:
		switch y := b.(type) {
		case string:
			return x <= y
		}
	case time.Time:
		switch y := b.(type) {
		case time.Time:
			return x.Before(y) || x.Equal(y)
		}
	case time.Duration:
		switch y := b.(type) {
		case time.Duration:
			return x <= y
		}
	}
	panic(fmt.Sprintf("invalid operation: %T <= %T", a, b))
}

func MoreOrEqual(a, b interface{}) bool {
	switch x := a.(type) {
	{{ cases ">=" }}
	case string:
		switch y := b.(type) {
		case string:
			return x >= y
		}
	case time.Time:
		switch y := b.(type) {
		case time.Time:
			return x.After(y) || x.Equal(y)
		}
	case time.Duration:
		switch y := b.(type) {
		case time.Duration:
			return x >= y
		}
	}
	panic(fmt.Sprintf("invalid operation: %T >= %T", a, b))
}

func Add(a, b interface{}) interface{} {
	switch x := a.(type) {
	{{ cases "+" }}
	case string:
		switch y := b.(type) {
		case string:
			return x + y
		}
	case time.Time:
		switch y := b.(type) {
		case time.Duration:
			return x.Add(y)
		}
	case time.Duration:
		switch y := b.(type) {
		case time.Time:
			return y.Add(x)
		case time.Duration:
			return x + y
		}
	}
	panic(fmt.Sprintf("invalid operation: %T + %T", a, b))
}

func Subtract(a, b interface{}) interface{} {
	switch x := a.(type) {
	{{ cases "-" }}
	case time.Time:
		switch y := b.(type) {
		case time.Time:
			return x.Sub(y)
		case time.Duration:
			return x.Add(-y)
		}
	case time.Duration:
		switch y := b.(type) {
		case time.Duration:
			return x - y
		}
	}
	panic(fmt.Sprintf("invalid operation: %T - %T", a, b))
}

func Multiply(a, b interface{}) interface{} {
	switch x := a.(type) {
	{{ cases_with_duration "*" }}
	}
	panic(fmt.Sprintf("invalid operation: %T * %T", a, b))
}

func Divide(a, b interface{}) float64 {
	switch x := a.(type) {
	{{ cases "/" }}
	}
	panic(fmt.Sprintf("invalid operation: %T / %T", a, b))
}

func Modulo(a, b interface{}) int {
	switch x := a.(type) {
	{{ cases_int_only "%" }}
	}
	panic(fmt.Sprintf("invalid operation: %T %% %T", a, b))
}
`
