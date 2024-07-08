package pikachu

type ITestCase interface {
	Prepare() error
	Execute(IProtoMessage) (IProtoMessage, error)
	Cleanup() error
	FlowMap() map[string]func() (IProtoMessage, *TestAssert)
}

type TestCase struct {
	Request  IProtoMessage
	Response IProtoMessage
	Error    error

	tracer  *TestCaseTracer
	flowmap map[string]func() (IProtoMessage, *TestAssert)
}

func (inst *TestCase) Tracer() *TestCaseTracer {
	return inst.tracer
}

func (inst *TestCase) FlowMap() map[string]func() (IProtoMessage, *TestAssert) {
	if inst.flowmap == nil {
		return make(map[string]func() (IProtoMessage, *TestAssert), 0)
	}

	return inst.flowmap
}

func (inst *TestCase) Register(name string, fc func() (IProtoMessage, *TestAssert)) {
	if inst.flowmap == nil {
		inst.flowmap = make(map[string]func() (IProtoMessage, *TestAssert), 0)
	}
	inst.flowmap[name] = fc
}
