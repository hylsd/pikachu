package pikachu

type ITestCase interface {
	Name() string
	FlowMap() map[string]func() (IProtoMessage, *TestAssert)

	Prepare() error
	Execute(IProtoMessage) (IProtoMessage, error)
	Cleanup() error
}

type TestCase struct {
	Request  IProtoMessage
	Response IProtoMessage
	Error    error

	name    string
	flowmap map[string]func() (IProtoMessage, *TestAssert)
}

func (inst *TestCase) Name() string {
	return inst.name
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
