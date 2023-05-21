package tcp

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func Unmarshall(s Stream, v any) (time.Time, error) {
	if n, t, err := unmarshall(s, reflect.ValueOf(v)); err != nil {
		return t, err
	} else {
		return t, s.Consume(n)
	}
}

func unmarshall(s PeekStream, v reflect.Value) (int, time.Time, error) {
	pack := make([]byte, s.Length())
	n, t, _, err := s.Peek(pack)
	if err != nil && err != io.EOF {
		return n, t, err
	} else if n != len(pack) {
		pack = pack[:n]
	}
	n, err = unpack(pack, v)
	return n, t, err
}

func unpack(pack []byte, v reflect.Value) (n int, err error) {
	switch k := v.Kind(); k {
	case reflect.Pointer, reflect.Interface:
		return unpack(pack, v.Elem())
	case reflect.Uint8:
		if len(pack) < n+1 {
			return n, io.EOF
		}
		v.SetUint(uint64(pack[n]))
		n += 1
	case reflect.Uint16:
		if len(pack) < n+2 {
			return n, io.EOF
		}
		v.SetUint(uint64(binary.BigEndian.Uint16(pack[n : n+2])))
		n += 2
	case reflect.Uint32:
		if len(pack) < n+4 {
			return n, io.EOF
		}
		v.SetUint(uint64(binary.BigEndian.Uint32(pack[n : n+4])))
		n += 4
	case reflect.Int32:
		if len(pack) < n+4 {
			return n, io.EOF
		}
		v.SetInt(int64(int32(binary.BigEndian.Uint32(pack[n : n+4]))))
		n += 4
	case reflect.Array:
		for j := 0; j < v.Type().Len(); j++ {
			if len(pack) <= n {
				return n, io.EOF
			}
			var m int
			m, err = unpack(pack[n:], v.Index(j))
			n += m
			if err != nil {
				break
			}
		}
	case reflect.Struct:
		return unpackStructure(pack, v)
	default:
		err = fmt.Errorf(`cannot unmarshal kind %s`, k.String())
	}
	return n, err
}

func unpackStructure(pack []byte, v reflect.Value) (n int, err error) {
	sizes := make(map[string]reflect.Value)
	for i, t := 0, v.Type(); i < t.NumField(); i++ {
		f := v.Field(i)
		// Handle any tags
		if tag, ok := t.Field(i).Tag.Lookup(`tcp`); ok {
			parts := strings.Split(tag, `,`)
			// Handle second positions (either a length value or reference)
			if len(parts) >= 2 {
				ref := strings.TrimSpace(parts[1])
				// Require the reference to be set (avoids ",,..." from evaluating as 0)
				if len(ref) > 0 {
					if f.CanUint() {
						// If the field is an uint, consider it a size reference
						sizes[ref] = f
					} else if size, err := strconv.ParseUint(ref, 0, 0); err == nil {
						// Otherwise consider it a size value
						sizes[t.Field(i).Name] = reflect.ValueOf(size)
					}
				}
			}
		}
		// Assign the value
		switch f.Kind() {
		case reflect.String:
			size := len(pack) - n
			if ref, ok := sizes[t.Field(i).Name]; ok && ref.CanUint() {
				size = int(ref.Uint())
			} else if ok && ref.CanInt() {
				size = int(ref.Int())
			}
			if len(pack) < n+size {
				return n, io.EOF
			}
			f.SetString(string(pack[n : n+size]))
			n += size
		case reflect.Slice:
			size := f.Len()
			ref, ok := sizes[t.Field(i).Name]
			if ok && ref.CanUint() {
				size = int(ref.Uint())
			} else if ok && ref.CanInt() {
				size = int(ref.Int())
			}

			j := 0
			for ; len(pack) > n && j < size; j++ {
				e := reflect.New(f.Type().Elem()).Elem()
				m, err := unpack(pack[n:], e)
				n += m
				if err != nil {
					return n, err
				}
				f.Set(reflect.Append(f, e))
			}

			if ok && j < size {
				return n, io.EOF
			}

		case reflect.Struct:
			m, err := unpackStructure(pack[n:], f)
			n += m
			if err != nil {
				return n, err
			}
		default:
			var m int
			m, err = unpack(pack[n:], f)
			n += m
		}

		if err != nil {
			break
		}
	}
	return n, err
}

func UnmarshallPeek(s PeekStream, v any) (time.Time, error) {
	_, t, err := unmarshall(s, reflect.ValueOf(v))
	return t, err
}
