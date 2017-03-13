package graphql

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

type Marshaler interface {
	MarshalGraphQL(w *Writer, name string) error
}

type HideStruct struct{}

func (HideStruct) MarshalGraphQL(w *Writer, name string) error {
	w.Println(name)
	return nil
}

func UnmarshalJSONString(data []byte, v interface{}) error {
	s := string(data)
	n := len(s)
	if n < 2 || s[0] != '"' || s[0] != s[n-1] {
		return fmt.Errorf("expected quoted string")
	}
	s = s[1 : n-1]
	s = strings.Replace(s, `\"`, `"`, -1)

	return json.Unmarshal([]byte(s), v)
}

type PageInfo struct {
	HasNextPage     bool
	HasPreviousPage bool
	StartCursor     string
	EndCursor       string
}

type Connection struct {
	TotalCount int
	PageInfo   PageInfo
}

//func (c Connection) MarshalGraphQL() ([]byte, error) {
//	return nil, nil
//}

type ID int

func (i *ID) UnmarshalJSON(data []byte) error {
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("expected string")
	}

	data = data[1 : len(data)-1]
	data, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return err
	}

	splits := strings.Split(string(data), ":")
	if len(splits) != 2 {
		return fmt.Errorf("unexpected decoded id: %s (%+v)", string(data), splits)
	}

	id, err := strconv.Atoi(splits[1])
	if err != nil {
		return fmt.Errorf(`invalid id in "%s": %s`, string(data), err)
	}
	*i = ID(id)

	return nil
}

func getName(field reflect.StructField) string {
	name := strings.Replace(field.Name, "ID", "Id", -1)
	name = strings.Replace(name, "URL", "Url", -1)
	t := []rune(name)
	t[0] = unicode.ToLower(t[0])
	return string(t)
}

func getNameAndArguments(field reflect.StructField) string {
	name := getName(field)
	if value, ok := field.Tag.Lookup("graphql"); ok {
		name = fmt.Sprintf("%s(%s)", name, value)
	}
	return name
}

type queryVariable struct {
	name     string
	typeName string
}

func (q *queryVariable) String() string {
	return fmt.Sprintf("$%s: %s", q.name, q.typeName)
}

type Query struct {
	name      string
	object    interface{}
	variables []queryVariable
}

func NewQuery(name string, object interface{}) *Query {
	return &Query{
		name:   name,
		object: object,
	}
}

func (q *Query) signature() string {
	if len(q.name) == 0 {
		return ""
	}
	s := fmt.Sprintf("query %s", q.name)
	if len(q.variables) > 0 {
		vs := make([]string, 0, len(q.variables))
		for _, v := range q.variables {
			vs = append(vs, v.String())
		}
		s = fmt.Sprintf("%s(%s)", s, strings.Join(vs, ", "))
	}
	return s
}

func query(w *Writer, name string, ty reflect.Type) {

	queryStruct := func(structType reflect.Type) {
		for i := 0; i < structType.NumField(); i++ {
			field := structType.Field(i)
			name := getNameAndArguments(field)
			query(w, name, field.Type)
		}
	}

	switch ty.Kind() {
	case reflect.Struct:
		if ty == reflect.TypeOf((*Connection)(nil)).Elem() {
			queryStruct(ty)
			return
		}

		marshalerType := reflect.TypeOf((*Marshaler)(nil)).Elem()
		if ty.Implements(marshalerType) {
			err := reflect.New(ty).Interface().(Marshaler).MarshalGraphQL(w, name)
			if err != nil {
				return // TODO error
			}
			return
		}

		w.Scope(name, func() {
			queryStruct(ty)
		})
	case reflect.Slice:
		elem := ty.Elem()
		query(w, name, elem)
	default:
		w.Println(name)
	}
}

func (q *Query) Marshal() ([]byte, error) {
	qType := reflect.TypeOf(q.object)
	if qType.Kind() != reflect.Struct || qType.NumField() == 0 {
		return nil, fmt.Errorf("query object must be a non-empty struct")
	}

	w := new(Writer)

	w.Scope(q.signature(), func() {
		for i := 0; i < qType.NumField(); i++ {
			field := qType.Field(i)
			name := getNameAndArguments(field)
			query(w, name, field.Type)
		}
	})

	return w.Bytes(), nil
}

func (q *Query) MarshalString() (string, error) {
	bytes, err := q.Marshal()
	return string(bytes), err
}

func (q *Query) DefineVariable(name, typeName string) *Query {
	q.variables = append(q.variables, queryVariable{name, typeName})
	return q
}

func Unmarshal(data []byte, obj interface{}) error {
	test := struct {
		Data map[string]json.RawMessage
	}{}
	if err := json.Unmarshal(data, &test); err != nil {
		return err
	}

	value := reflect.Indirect(reflect.ValueOf(obj))
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := value.Type().Field(i)

		name := getName(fieldType)

		raw, ok := test.Data[name]
		if !ok {
			continue
		}

		p := reflect.New(field.Type())
		if err := json.Unmarshal(raw, p.Interface()); err != nil {
			return err
		}
		field.Set(reflect.Indirect(p))
	}

	return nil
}
