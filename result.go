package pikachu

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const TIME_LAYOUT = "2006-01-02 15:04:05"
const TEST_REPORT_TEMPLATE = `
#######################################################################################
##
## Test Report: %s
## ------------------------------------------------------------------------------------
##
## Start Time: %s		End Time: %s
## %s
##
## Notice: %s
##
## ------------------------------------------------------------------------------------
## Test Result:
## %s 
##
#######################################################################################
##
## Details: 
%s

`

const (
	TEST_RESULT_PASS    = "PASS"
	TEST_RESULT_FAILED  = "FAILED"
	TEST_RESULT_UNKOWN  = "UNKOWN"
	TEST_RESULT_TIMEOUT = "TIMEOUT"

	TEST_REG_STATE_IDLE     = "IDLE"
	TEST_REG_STATE_FAILED   = "FAILED"
	TEST_REG_STATE_RUNNING  = "RUNNING"
	TEST_REG_STATE_FINISHED = "FINISHED"
)

type TestReport struct {
	uniqueNo  string
	outPath   string
	startTime time.Time
	endTime   time.Time
	state     string
	lastWord  string
	results   []*TestResult
}

func (inst *TestReport) Done() {
	if inst.state == TEST_REG_STATE_IDLE || inst.state == TEST_REG_STATE_RUNNING {
		inst.state = TEST_REG_STATE_FINISHED
	}
	inst.endTime = time.Now()
}

func (inst *TestReport) Interrupt(lastword string) {
	inst.state = TEST_REG_STATE_FAILED
	inst.lastWord = lastword
}

func (inst *TestReport) Append(ts, tc, flow string) *TestResult {
	result := NewTestResult(ts, tc, flow)
	inst.results = append(inst.results, result)
	return result
}

func (inst *TestReport) Export() string {
	testSummary := NewTestSummary()
	for _, result := range inst.results {
		testSummary.Calculate(result)
	}

	testReults := make([]string, len(inst.results))
	for idx, result := range inst.results {
		testReults[idx] = result.ResultToString()
	}
	content := strings.Join(testReults, "\n")

	startTime := inst.startTime.Format(TIME_LAYOUT)
	endTime := inst.endTime.Format(TIME_LAYOUT)

	details := make([]string, 0)
	for _, result := range inst.results {
		title := "## " + result.Title()
		details = append(details, title)

		resultDetails := result.Details()
		details = append(details, resultDetails...)
	}

	detailContents := strings.Join(details, "\n")

	return fmt.Sprintf(TEST_REPORT_TEMPLATE,
		inst.uniqueNo,
		startTime,
		endTime,
		testSummary.ToString(),
		inst.lastWord,
		content,
		detailContents,
	)
}

func NewTestReport() *TestReport {
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}

	currentTime := time.Now()
	formattedTime := currentTime.Format(TIME_LAYOUT)
	uniqueNo := fmt.Sprintf("test_report_%s_%d", formattedTime, rand.Intn(9000)+1000)
	outpath := filepath.Join(cwd, "regression", uniqueNo+".txt")

	return &TestReport{
		uniqueNo:  uniqueNo,
		outPath:   outpath,
		startTime: currentTime,
		endTime:   currentTime,
		state:     TEST_REG_STATE_IDLE,
		results:   make([]*TestResult, 0),
	}
}

const TEST_RESULT_TEMPLATE = "| Start Time: %s | End Time %s | Result %8s | %s.%s.%s"

type TestResult struct {
	testsuite string
	testcase  string
	testflow  string
	startTime time.Time
	endTime   time.Time
	result    string
	details   map[string]string
	incr      int32
}

func (inst *TestResult) SetResult(result string) {
	inst.result = result
}

func (inst *TestResult) Record(phase, format string, values ...any) {
	inst.incr += 1

	ts := time.Now().Format(TIME_LAYOUT)
	content := fmt.Sprintf(format, values...)
	content = fmt.Sprintf("[%7s] %s", phase, content)
	prefix := fmt.Sprintf("%03d.%s", inst.incr, ts)
	inst.details[prefix] = content
}

func (inst *TestResult) RecordAsError(phase, format string, values ...any) {
	inst.Record(phase, format, values...)
	inst.result = TEST_RESULT_FAILED
}

func (inst *TestResult) Result() (string, string, string, string) {
	return inst.testsuite, inst.testcase, inst.testflow, inst.result
}

func (inst *TestResult) Title() string {
	return fmt.Sprintf("%s:%s:%s", inst.testsuite, inst.testcase, inst.testflow)
}

func (inst *TestResult) Details() []string {
	keys := make([]string, len(inst.details))
	for k, _ := range inst.details {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	slices := make([]string, 0, len(inst.details))
	for _, orderedKey := range keys {
		value, exist := inst.details[orderedKey]
		if !exist {
			continue
		}

		content := fmt.Sprintf("##     [%s] %s", orderedKey, value)
		slices = append(slices, content)
	}

	return slices
}

func (inst *TestResult) ResultToString() string {
	startTime := inst.startTime.Format(TIME_LAYOUT)
	endTime := inst.endTime.Format(TIME_LAYOUT)

	return fmt.Sprintf(TEST_RESULT_TEMPLATE,
		startTime,
		endTime,
		inst.result,
		inst.testsuite,
		inst.testcase,
		inst.testflow,
	)
}

func NewTestResult(ts, tc, flow string) *TestResult {
	currentTime := time.Now()
	return &TestResult{
		testsuite: ts,
		testcase:  tc,
		testflow:  flow,
		startTime: currentTime,
		endTime:   currentTime,
		result:    TEST_RESULT_UNKOWN,
		incr:      0,
		details:   make(map[string]string, 0),
	}
}

const TEST_SUMMARY_TEMPLATE = `
## Test Summary:
##   Total Suites: %3d  Total Cases: %3d  Total Flows: %3d
##	 Passed: %3d     Failed: %3d     Timeout: %3d      Other: %3d`

type TestSummary struct {
	testsuite map[string]string
	testcase  map[string]string
	testflow  map[string]string
	result    map[string]int
}

func (inst *TestSummary) Calculate(result *TestResult) {
	ts, tc, flow, ret := result.Result()
	inst.testsuite[ts] = ts
	inst.testcase[tc] = tc
	inst.testflow[flow] = flow
	_, exist := inst.result[ret]
	if !exist {
		inst.result[ret] = 1
	} else {
		inst.result[ret] = inst.result[ret] + 1
	}
}

func (inst *TestSummary) ToString() string {
	suiteCount := len(inst.testsuite)
	caseCount := len(inst.testcase)
	flowCount := len(inst.testflow)

	var pass, failed, timeout, other int
	cnt, exist := inst.result[TEST_RESULT_PASS]
	if !exist {
		pass = 0
	} else {
		pass = cnt
	}

	cnt, exist = inst.result[TEST_RESULT_FAILED]
	if !exist {
		failed = 0
	} else {
		failed = cnt
	}

	cnt, exist = inst.result[TEST_RESULT_UNKOWN]
	if !exist {
		other = 0
	} else {
		other = cnt
	}

	cnt, exist = inst.result[TEST_RESULT_TIMEOUT]
	if !exist {
		timeout = 0
	} else {
		timeout = cnt
	}

	return fmt.Sprintf(TEST_SUMMARY_TEMPLATE,
		suiteCount, caseCount, flowCount, pass, failed, timeout, other)
}

func NewTestSummary() *TestSummary {
	return &TestSummary{
		testsuite: make(map[string]string, 0),
		testcase:  make(map[string]string, 0),
		testflow:  make(map[string]string, 0),
		result:    make(map[string]int, 0),
	}
}
