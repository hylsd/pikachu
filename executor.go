package pikachu

import (
	"fmt"
	"strings"
)

func AddTestCase(ts string, testcase ITestCase) {
	testsuite, exist := regression[ts]
	if !exist {
		return
	}

	testsuite.AddTestCase(testcase)
}

func Run(testsuites string) {
	testReport := NewTestReport()
	defer testReport.Done()

	defer func() {
		if r := recover(); r != nil {
			testReport.Interrupt("encounter panic")
		}
		fmt.Println(testReport.Export())
	}()

	slices := strings.Split(testsuites, ".")
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