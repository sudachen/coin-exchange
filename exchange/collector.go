package exchange

const MsgQueueLength = 61

type collector struct{ Messages chan interface{} }

var Collector = &collector{
	make(chan interface{}, MsgQueueLength),
}

func (m *collector) Next() interface{} {
	return <-m.Messages
}
