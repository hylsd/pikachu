package pikachu

import (
	"fmt"
	"reflect"
	"strings"
)

type IProtoMessage interface {
	String() string
}

type ProtoMessageElem struct {
	name  string
	kind  reflect.Kind
	value reflect.Value
}

func (inst *ProtoMessageElem) SetValue(value reflect.Value) {
	inst.value = value
}

func (inst *ProtoMessageElem) Kind() reflect.Kind {
	return inst.kind
}

func (inst *ProtoMessageElem) IsString() bool {
	return inst.kind == reflect.String
}
func (inst *ProtoMessageElem) IsInteger() bool {
	switch inst.kind {
	case reflect.Int:
	case reflect.Int8:
	case reflect.Int16:
	case reflect.Int32:
	case reflect.Int64:
		return true
	default:
	}

	return false
}

func (inst *ProtoMessageElem) HasLength() bool {
	switch inst.kind {
	case reflect.Array:
	case reflect.Chan:
	case reflect.Slice:
	case reflect.Map:
	case reflect.String:
		return true
	default:
	}

	return false
}

func (inst *ProtoMessageElem) ToString() string {
	return fmt.Sprintf("Name: %s, Kind: %s", inst.name, inst.kind)
}

func NewProtoMessageElem(name string, kind reflect.Kind) *ProtoMessageElem {
	return &ProtoMessageElem{
		name: name,
		kind: kind,
	}
}

func parseAllMessageElements(messaga interface{}) map[string]*ProtoMessageElem {
	if reflect.ValueOf(messaga).Kind() == reflect.Ptr {
		msg := reflect.ValueOf(messaga).Elem()
		return parseMessageElements("", msg)
	} else {
		msg := reflect.ValueOf(messaga)
		return parseMessageElements("", msg)
	}
}

func parseMessageElements(parent string, msg reflect.Value) map[string]*ProtoMessageElem {
	elements := make(map[string]*ProtoMessageElem, 0)
	msgType := msg.Type()
	for i := 0; i < msg.NumField(); i++ {
		var curName string
		if len(parent) < 1 {
			curName = msgType.Field(i).Name
		} else {
			curName = fmt.Sprintf("%s.%s", parent, msgType.Field(i).Name)
		}
		curName = strings.ToLower(curName)

		switch msg.Field(i).Kind() {
		case reflect.Ptr:
			submsg := msg.Field(i).Elem()
			subElements := parseMessageElements(curName, submsg)
			for k, v := range subElements {
				elements[k] = v
			}
		default:
			elem := NewProtoMessageElem(curName, msg.Field(i).Kind())
			elem.SetValue(msg.Field(i))
			elements[curName] = elem
		}
	}

	return elements
}
