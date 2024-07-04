package pikachu

import (
	"fmt"
	"strings"
)

const (
	TC_SYNTAX_RUN_ALL  = "*"
	TEST_PHASE_PREPARE = "Prepare"
	TEST_PHASE_EXECUTE = "Execute"
	TEST_PHASE_ASSERT  = "Assert"
	TEST_PHASE_CLEANUP = "Cleanup"
)

var regression map[string]*TestSuite

func init() {
	regression = make(map[string]*TestSuite, 0)
}

type TestSuite struct {
	name       string
	testcases  map[string]ITestCase
	testReport *TestReport
}

func (inst *TestSuite) AddTestCase(tc string, fc ITestCase) {
	inst.testcases[tc] = fc
}

func (inst *TestSuite) Run(args ...string) {
	if args[0] == TC_SYNTAX_RUN_ALL {
		for tcName, testcase := range inst.testcases {
			for flowName, constructor := range testcase.FlowMap() {
				testResult := inst.testReport.Append(inst.name, tcName, flowName)
				inst.runTestCase(testcase, constructor, testResult)
			}
		}
	} else {
		tcName := args[0]
		testcase, exist := inst.testcases[tcName]
		if !exist {
			inst.testReport.Interrupt("test case not exist ")
			return
		}

		flowmap := testcase.FlowMap()
		if len(args) > 1 && args[1] != TC_SYNTAX_RUN_ALL {
			flowName := args[1]

			constructor, exist := flowmap[flowName]
			if !exist {
				inst.testReport.Interrupt("test flow not exist ")
				return
			}

			testResult := inst.testReport.Append(inst.name, tcName, flowName)
			inst.runTestCase(testcase, constructor, testResult)
			return
		}

		for flowName, constructor := range flowmap {
			testResult := inst.testReport.Append(inst.name, tcName, flowName)
			inst.runTestCase(testcase, constructor, testResult)
		}
	}
}

func (inst *TestSuite) runTestCase(tc ITestCase, constructor func() (IProtoMessage, *TestAssert), result *TestResult) {
	result.Record(TEST_PHASE_PREPARE, "started")
	req, assert := constructor()

	err := tc.Prepare()
	if err != nil {
		result.RecordAsError(TEST_PHASE_PREPARE, "failed")
		return
	}
	result.Record(TEST_PHASE_PREPARE, "finished")

	defer func() {
		result.Record(TEST_PHASE_CLEANUP, "started")
		tc.Cleanup()
		result.Record(TEST_PHASE_CLEANUP, "finished")
	}()

	result.Record(TEST_PHASE_EXECUTE, "started")
	resp, err := tc.Execute(req)
	result.Record(TEST_PHASE_EXECUTE, "finished")

	if err != nil {
		result.RecordAsError(TEST_PHASE_EXECUTE, "error response %s", err)
		return
	}

	assert.Check(resp, result)
}

func (inst *TestSuite) SetReport(report *TestReport) {
	inst.testReport = report
}

func NewTestSuite(name string) *TestSuite {
	ts := &TestSuite{
		name:       name,
		testcases:  make(map[string]ITestCase, 0),
		testReport: nil,
	}

	regression[name] = ts
	return ts
}

func Run(testcase string) {
	testReport := NewTestReport()
	defer testReport.Done()

	defer func() {
		if r := recover(); r != nil {
			testReport.Interrupt("encounter panic")
		}
		fmt.Println(testReport.Export())
	}()

	slices := strings.Split(testcase, ".")
	if len(slices[0]) < 1 {
		testReport.Interrupt("unspecified test suite")
		return
	}

	testsuiteName := slices[0]
	testsuite, exist := regression[testsuiteName]
	if !exist {
		testReport.Interrupt("test suite not exist")
		return
	}

	testsuite.SetReport(testReport)
	cnt := len(slices)
	switch cnt {
	case 1:
		testsuite.Run(TC_SYNTAX_RUN_ALL)
	case 2:
		testsuite.Run(slices[1], TC_SYNTAX_RUN_ALL)
	default:
		testsuite.Run(slices[1], slices[2])
	}
}
