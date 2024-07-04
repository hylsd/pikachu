package pikachu

import "fmt"

type TestCaseTracer struct {
	//tcName string
}

func (inst *TestCaseTracer) Enter() {
	fmt.Println("1")
}

func (inst *TestCaseTracer) Exception(exception string) {
	fmt.Println("1")
}

func (inst *TestCaseTracer) Leave() {
	fmt.Println("2")
}

func (inst *TestCaseTracer) SendMessage(req, resp IProtoMessage, err error) {
	fmt.Println("2")
}
