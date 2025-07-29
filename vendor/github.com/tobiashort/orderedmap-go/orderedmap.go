package orderedmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"iter"
	"reflect"
	"runtime/debug"
	"strings"
)

type OrderedMap[K comparable, V any] struct {
	keys      []K
	keyValues map[K]V
}

func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		keys:      make([]K, 0),
		keyValues: make(map[K]V),
	}
}

func (m *OrderedMap[K, V]) Len() int {
	return len(m.keys)
}

func (m *OrderedMap[K, V]) Keys() []K {
	clone := make([]K, m.Len())
	copy(clone, m.keys)
	return clone
}

func (m *OrderedMap[K, V]) Values() []V {
	values := make([]V, m.Len())
	for idx, key := range m.keys {
		values[idx] = m.keyValues[key]
	}
	return values
}

func (m *OrderedMap[K, V]) Has(key K) bool {
	_, ok := m.keyValues[key]
	return ok
}

func (m *OrderedMap[K, V]) Put(key K, value V) {
	hasKey := m.Has(key)
	m.keyValues[key] = value
	if !hasKey {
		m.keys = append(m.keys, key)
	}
}

func (m *OrderedMap[K, V]) Get(key K) (V, bool) {
	value, ok := m.keyValues[key]
	return value, ok
}

func (m *OrderedMap[K, V]) Del(key K) {
	if m.Has(key) {
		delete(m.keyValues, key)
		newKeys := make([]K, 0)
		for _, keyInKeys := range m.keys {
			if keyInKeys != key {
				newKeys = append(newKeys, keyInKeys)
			}
		}
		m.keys = newKeys
	}
}

func (m *OrderedMap[K, V]) Iterate() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for _, key := range m.keys {
			if !yield(key, m.keyValues[key]) {
				return
			}
		}
	}
}

func (m *OrderedMap[K, V]) UnmarshalJSON(data []byte) (err error) {
	if reflect.TypeOf(m) != reflect.TypeOf(NewOrderedMap[string, V]()) {
		return fmt.Errorf("UnmarshalJSON is not implemented for type %s. Keys must be of type string.", reflect.TypeOf(m))
	}

	defer func() {
		if msg := recover(); msg != nil {
			err = fmt.Errorf("%v\n%s", msg, debug.Stack())
		}
	}()

	assertNil := func(v any) {
		if v != nil {
			panic(v)
		}
	}

	m.keys = make([]K, 0)
	m.keyValues = make(map[K]V)

	decoder := json.NewDecoder(bytes.NewReader(data))

	// parse {
	token, err := decoder.Token()
	assertNil(err)
	if token != json.Delim('{') {
		return fmt.Errorf("at %d expected '{', got '%s'", decoder.InputOffset(), token)
	}

next:
	// parse }
	token, err = decoder.Token()
	assertNil(err)
	if token == json.Delim('}') {
		return nil
	}

	// if not } we parsed a key
	key := token.(K)

	// parse value
	var valueRaw json.RawMessage
	err = decoder.Decode(&valueRaw)
	assertNil(err)

	var valueCompacted bytes.Buffer
	err = json.Compact(&valueCompacted, valueRaw)
	assertNil(err)

	valueStr := valueCompacted.String()

	var value any

	// if current OrderedMap is of type [string, any],
	// subsequent json objects shall also be unmarshaled into
	// OrderedMap[string, any]
	isStringToAny := reflect.TypeOf(m) == reflect.TypeOf(NewOrderedMap[string, any]())
	if isStringToAny && strings.HasPrefix(valueStr, "{") {
		var omap OrderedMap[K, V]
		err = json.Unmarshal([]byte(valueStr), &omap)
		assertNil(err)
		value = omap
	} else if isStringToAny && strings.HasPrefix(valueStr, "[{") {
		var omaps []OrderedMap[K, V]
		err = json.Unmarshal([]byte(valueStr), &omaps)
		assertNil(err)
		value = omaps
	} else if strings.HasPrefix(valueStr, "{") {
		// if current OrderedMap is of type [string, V],
		// subsequent json object shall be unmarshaled to
		// the concrete type V
		var tmp V
		err = json.Unmarshal([]byte(valueStr), &tmp)
		assertNil(err)
		value = tmp
	} else if strings.HasPrefix(valueStr, "[{") {
		var tmp []V
		err = json.Unmarshal([]byte(valueStr), &tmp)
		assertNil(err)
		value = tmp
	} else {
		// in any other case the valueStr shall be unmarshaled
		// to whatever type it may be (int, float, bool, str, etc.)
		err = json.Unmarshal([]byte(valueStr), &value)
		assertNil(err)
	}

	// add key and value
	m.keys = append(m.keys, key)
	m.keyValues[key] = value.(V)

	// continue parsing
	goto next
}

func (m OrderedMap[K, V]) MarshalJSON() ([]byte, error) {
	if reflect.TypeOf(m) != reflect.TypeOf(*NewOrderedMap[string, V]()) {
		return nil, fmt.Errorf("MarshalJSON is not implemented for type %s. Keys must be of type string.", reflect.TypeOf(m))
	}

	builder := strings.Builder{}
	encoder := json.NewEncoder(&builder)
	builder.WriteString("{")
	for idx, key := range m.keys {
		val := m.keyValues[key]
		builder.WriteString(fmt.Sprintf(`"%v":`, key))
		err := encoder.Encode(val)
		if err != nil {
			return nil, err
		}
		if idx < m.Len()-1 {
			builder.WriteString(",")
		}
	}
	builder.WriteString("}")
	return []byte(builder.String()), nil
}
