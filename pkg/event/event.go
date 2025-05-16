package event

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"sync"
	"unicode"
	"unique"

	//nolint:exptostd // constraint.Integer, constraint.Float do not exist in cmp.
	"golang.org/x/exp/constraints"
)

type ustring = unique.Handle[string]

type Record struct {
	// eventName is a name for a wide event this record represents.
	// Child records have eventName = "".
	eventName ustring

	mu     sync.Mutex
	values map[ustring]any
}

func (r *Record) LogValue() slog.Value {
	r.mu.Lock()
	defer r.mu.Unlock()

	values := make([]slog.Attr, 0, len(r.values)+1)
	if r.eventName != (ustring{}) {
		values = append(values, slog.String("event", r.eventName.Value()))
	}
	for name, v := range r.values {
		reco, ok := v.(Recorder)
		if ok {
			values = append(values, slog.Any(name.Value(), reco.EventRecord()))
		} else {
			values = append(values, slog.Any(name.Value(), v))
		}
	}

	return slog.GroupValue(values...)
}

func wrongKey(key string) bool {
	return key == "" ||
		strings.ContainsFunc(
			key,
			func(r rune) bool {
				return r != '/' && r != '_' && r != '@' && (unicode.IsPunct(r) || unicode.IsSpace(r))
			},
		)
}

type Recorder interface {
	EventRecord() *Record
}

// Add add a new value to the Record.
// If the value already exists, Add tries to sum the existing value and a new value.
// The existing value and the new value must be of the same type.
//
// The list of supported types:
// - All integer types.
// - All float types.
// - Error types.
func (r *Record) Add(keyValuePairs ...any) {
	r.putValues(true, keyValuePairs)
}

// Set sets a value in the Record.
// If the value already exists, it is overwritten.
func (r *Record) Set(keyValuePairs ...any) {
	r.putValues(false, keyValuePairs)
}

func (r *Record) putValues(add bool, keyValuePairs []any) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(keyValuePairs)%2 != 0 {
		panic("keys and values should be in pairs")
	}

	for i := 0; i < len(keyValuePairs); i += 2 {
		key, ok := keyValuePairs[i].(string)
		if !ok {
			panic(
				"keys should be non-empty strings, should not contain punctuation, spaces and '$', except for '_', '/' and'@'",
			)
		}

		if wrongKey(key) {
			panic(
				"keys should be non-empty strings, should not contain punctuation, spaces and '$', except for '_', '/' and'@'",
			)
		}
		if add {
			r.values[unique.Make(key)] = addValues(r.values[unique.Make(key)], keyValuePairs[i+1])
		} else {
			r.values[unique.Make(key)] = keyValuePairs[i+1]
		}
	}
}

func addValues(to, v any) any {
	if to == nil {
		return v
	}

	if addNumbers[int8](&to, v) ||
		addNumbers[int16](&to, v) ||
		addNumbers[int32](&to, v) ||
		addNumbers[int64](&to, v) ||
		addNumbers[int](&to, v) ||
		addNumbers[uint8](&to, v) ||
		addNumbers[uint16](&to, v) ||
		addNumbers[uint32](&to, v) ||
		addNumbers[uint64](&to, v) ||
		addErrors(&to, v) {
		return to
	}

	panic(fmt.Sprintf("types %s and %s cannot be added", reflect.TypeOf(to).Name(), reflect.TypeOf(v).Name()))
}

func addErrors(to *any, val any) bool {
	s, ok1 := (*to).(error)
	v, ok2 := val.(error)

	// s already wraps v
	if errors.Is(s, v) {
		return true
	}

	if errors.Is(v, s) {
		*to = v
		return true
	}

	if ok1 && ok2 {
		*to = errors.Join(s, v)
		return true
	}
	return false
}

func addNumbers[T constraints.Integer | constraints.Float](to *any, val any) bool {
	s, ok1 := (*to).(T)
	v, ok2 := val.(T)
	if ok1 && ok2 {
		*to = s + v
		return true
	}
	return false
}

func (r *Record) Sub(name string) *Record {
	if name == "" || wrongKey(name) {
		panic(
			"names should be non-empty strings, should not contain punctuation, spaces and '$', except for '_', '/' and'@'",
		)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if sub := r.values[unique.Make(name)]; sub != nil {
		sub, ok := sub.(*Record)
		if !ok {
			panic("value with this name already exists")
		}

		return sub
	}

	sub := newRecord()
	r.values[unique.Make(name)] = sub

	return sub
}

func (r *Record) EventName() string {
	if r.eventName == (ustring{}) {
		return ""
	}
	return r.eventName.Value()
}

func (r *Record) AllValues() map[string]any {
	r.mu.Lock()
	defer r.mu.Unlock()

	vals := make(map[string]any)
	if r.eventName != (ustring{}) {
		vals["$event"] = r.eventName.Value()
	}
	for n, val := range r.values {
		switch v := val.(type) {
		case *Record:
			prefix := n.Value() + "."
			for n, val := range v.AllValues() {
				vals[prefix+n] = val
			}
		case Recorder:
			prefix := n.Value() + "."
			for n, val := range v.EventRecord().AllValues() {
				vals[prefix+n] = val
			}
		default:
			vals[n.Value()] = val
		}
	}

	return vals
}

func (r *Record) valueHolder(key string) (re *Record, subKey string) {
	idx := strings.IndexByte(key, '.')
	if idx == -1 {
		return r, key
	}

	// It's a key, value for which should be in some Record, that is a value of the current record.
	subKey = key[:idx]
	val, ok := r.values[unique.Make(subKey)]
	if !ok {
		return nil, ""
	}

	switch v := val.(type) {
	case *Record:
		return v.valueHolder(key[idx+1:])
	default:
		return nil, ""
	}
}

func (r *Record) Value(key string) any {
	if key == "" {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	vHolder, subKey := r.valueHolder(key)
	if vHolder == nil {
		return nil
	}

	if vHolder == r {
		return r.values[unique.Make(subKey)]
	}

	return vHolder.Value(subKey)
}

const defaultMapSize = 30

var mapPool = sync.Pool{
	New: func() any {
		return make(map[ustring]any, defaultMapSize)
	},
}

// Finish should be called when this Record will not be used again, to free the resources.
func (r *Record) Finish() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, v := range r.values {
		if re, ok := v.(*Record); ok {
			re.Finish()
		}
	}

	clear(r.values)
	mapPool.Put(r.values)
	r.values = nil
}

func (r *Record) Wrap(parent context.Context) context.Context {
	return context.WithValue(parent, eventCtxKey, r)
}

func newRecord() *Record {
	m, _ := mapPool.Get().(map[ustring]any)
	return &Record{
		values: m,
	}
}

func NewRecord(parent context.Context, eventName string) (context.Context, *Record) {
	rec := newRecord()
	newRecord().eventName = unique.Make(eventName)
	parent = context.WithValue(parent, rootCtxKey, rec)
	return context.WithValue(parent, eventCtxKey, rec), rec
}

func Group(keyValuePairs ...any) *Record {
	nr := newRecord()
	nr.Set(keyValuePairs...)
	return nr
}

type ctxKey int

const (
	eventCtxKey ctxKey = iota + 1
	rootCtxKey
)

func Get(from context.Context) *Record {
	ev := from.Value(eventCtxKey)
	if ev == nil {
		panic("context does not have any events associated with it")
	}

	r, _ := ev.(*Record)

	return r
}

func Root(from context.Context) *Record {
	ev := from.Value(rootCtxKey)
	if ev == nil {
		panic("context does not have any root events associated with it")
	}

	r, _ := ev.(*Record)

	return r
}
