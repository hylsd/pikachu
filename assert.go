package pikachu

import (
	"errors"
	"strconv"
	"strings"
)

const (
	TEST_ASSERT_OP_IDLE    = ""
	TEST_ASSERT_OP_EQUAL   = "=="
	TEST_ASSERT_OP_LARGER  = ">"
	TEST_ASSERT_OP_SMALLER = "<"
)

const (
	TEST_ASSERT_METHOD_VALUE  = "{{value}}"
	TEST_ASSERT_METHOD_LEN    = "{{len}}"
	TEST_ASSERT_METHOD_IGNORE = "{{ignore}}"
)

type TestAssertRule struct {
	sut      string
	method   string
	op       string
	expected interface{}
}

func (inst *TestAssertRule) IsIgnore() bool {
	return inst.method == TEST_ASSERT_METHOD_IGNORE
}

func (inst *TestAssertRule) Assert(elem *ProtoMessageElem) error {
	switch inst.method {
	case TEST_ASSERT_METHOD_VALUE:
		return inst.AssertValue(elem)
	case TEST_ASSERT_METHOD_LEN:
		return inst.AssertLength(elem)
	}

	return nil
}

func (inst *TestAssertRule) AssertValue(elem *ProtoMessageElem) error {
	if elem.IsString() {
		actual := elem.value.String()
		expected, ok := inst.expected.(string)
		if !ok {
			return errors.New("couldn't get sut value with right type")
		}

		if !inst.assertString(inst.op, actual, expected) {
			return errors.New("unmatched string value")
		}
	}

	if elem.IsInteger() {
		var expected int
		actual := int(elem.value.Int())
		expected, ok := inst.expected.(int)
		if !ok {
			expectedStr, ok := inst.expected.(string)
			if !ok {
				return errors.New("couldn't get sut value with right type")
			}

			var err error
			expected, err = strconv.Atoi(expectedStr)
			if err != nil {
				return errors.New("couldn't convert sut value to int type")
			}
		}

		if !inst.assertInt(inst.op, actual, expected) {
			return errors.New("unmatched int value")
		}
	}

	return nil
}

func (inst *TestAssertRule) AssertLength(elem *ProtoMessageElem) error {
	if !elem.HasLength() {
		return errors.New("couldn't get length")
	}

	length := elem.value.Len()
	expected, ok := inst.expected.(int)
	if !ok {
		return errors.New("couldn't get sut length")
	}

	if !inst.assertInt(inst.op, length, expected) {
		return errors.New("unmatched length value")
	}

	return nil
}

func (inst *TestAssertRule) assertInt(op string, value1, value2 int) bool {
	switch op {
	case TEST_ASSERT_OP_EQUAL:
		return (value1 == value2)
	case TEST_ASSERT_OP_LARGER:
		return (value1 > value2)
	case TEST_ASSERT_OP_SMALLER:
		return (value1 < value2)
	}
	return true
}

func (inst *TestAssertRule) assertString(op string, value1, value2 string) bool {
	switch op {
	case TEST_ASSERT_OP_EQUAL:
		return (value1 == value2)
	case TEST_ASSERT_OP_LARGER:
	case TEST_ASSERT_OP_SMALLER:
		return false
	}
	return false
}

func NewAssertRule(sut, method, op string, expected interface{}) *TestAssertRule {
	return &TestAssertRule{
		sut:      strings.ToLower(sut),
		method:   method,
		op:       op,
		expected: expected,
	}
}

type TestAssert struct {
	rules map[string]*TestAssertRule
}

func (inst *TestAssert) AddRule(sut, method, op string, value interface{}) {
	inst.rules[sut] = NewAssertRule(sut, method, op, value)
}

func (inst *TestAssert) AddIgnoreRule(sut string) {
	inst.rules[sut] = NewAssertRule(sut, TEST_ASSERT_METHOD_IGNORE, TEST_ASSERT_OP_IDLE, nil)
}

func (inst *TestAssert) Check(resp IProtoMessage, result *TestResult) {
	if resp == nil {
		result.RecordAsError(TEST_PHASE_ASSERT, "empty response")
		return
	}

	// parse response message
	elements := parseAllMessageElements(resp)

	// check items one by on
	for rulename, rule := range inst.rules {
		if rule.IsIgnore() {
			result.Record(TEST_PHASE_ASSERT, "Item '%s' is ignore", rulename)
			continue
		}

		match := strings.ToLower(rulename)
		elem, exist := elements[match]
		if !exist {
			result.RecordAsError(TEST_PHASE_ASSERT, "Item '%s' doesn't exist in response", rulename)
			return
		}

		err := rule.Assert(elem)
		if err != nil {
			result.RecordAsError(TEST_PHASE_ASSERT, "Item '%s' check failed, err = '%s'", rulename, err)
			return
		}

		result.Record(TEST_PHASE_ASSERT, "compare '%s' passed", rulename)
	}

	result.SetResult(TEST_RESULT_PASS)
}

func NewTestAssert() *TestAssert {
	return &TestAssert{
		rules: make(map[string]*TestAssertRule, 0),
	}
}
