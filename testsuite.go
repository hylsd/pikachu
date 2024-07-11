package pikachu

import (
	"time"
)

const (
	TC_SYNTAX_RUN_ALL  = "*"
	REG_SYNTAX_RUN_ALL = "all"
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

func (inst *TestSuite) AddTestCase(testcase ITestCase) {
	tcName := testcase.Name()
	inst.testcases[tcName] = testcase
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
	start := time.Now()

	// 1. prepare
	req, assert := constructor()
	err := tc.Prepare()
	if err != nil {
		result.RecordAsError(TEST_PHASE_PREPARE, "failed")
		return
	}
	result.Record(TEST_PHASE_PREPARE, "finished")

	// 4. cleanup
	defer func() {
		result.Record(TEST_PHASE_CLEANUP, "started")
		tc.Cleanup()
		result.Record(TEST_PHASE_CLEANUP, "finished")
	}()

	// 2. execute
	result.Record(TEST_PHASE_EXECUTE, "started")
	resp, err := tc.Execute(req)
	result.Record(TEST_PHASE_EXECUTE, "finished")
	if err != nil {
		result.RecordAsError(TEST_PHASE_EXECUTE, "error response %s", err)
		return
	}

	// 3. assert
	// check if the request is timeout
	end := time.Now()
	if err = assert.IsTimeout(start, end); err != nil {
		result.Record(TEST_PHASE_ASSERT, "request timeout %s", err)
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
