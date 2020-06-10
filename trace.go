package mvdreader

type traceInfo struct {
	enabled                  bool
	traces                   []*TraceParseFrame
	current                  *TraceParseFrame
	currentReadTraceReciever TraceReadAdd
	currentReadTrace         *TraceRead
}

func (mvd *Mvd) traceParseFrameStart() {
	if mvd.trace.enabled == false {
		return
	}

	traceParseFrame := new(TraceParseFrame)
	mvd.trace.current = traceParseFrame
	mvd.trace.traces = append(mvd.trace.traces, traceParseFrame)
	traceParseFrame.FrameStart = mvd.Frame
}

func (mvd *Mvd) traceParseFrameStop() {
	if mvd.trace.enabled == false {
		return
	}
	mvd.trace.current.LastFrame = mvd.done
	mvd.trace.current.FrameStop = mvd.Frame
}

func (mvd *Mvd) traceReadFrameStart() {
	if mvd.trace.enabled == false {
		return
	}
	cTrace := new(TraceReadFrame)
	mvd.trace.current.currentReadFrameTrace = cTrace
	mvd.trace.current.TraceReadFrame = append(mvd.trace.current.TraceReadFrame, cTrace)

	mvd.trace.currentReadTraceReciever = cTrace

	cTrace.FileOffsetStart = mvd.file_offset
}

func (mvd *Mvd) traceReadFrameStop() {
	if mvd.trace.enabled == false {
		return
	}
	cTrace := mvd.trace.current.currentReadFrameTrace
	cTrace.FileOffsetStop = mvd.file_offset
}

func (mvd *Mvd) traceAddReadTrace(identifier string) *TraceRead {
	if mvd.trace.enabled == false {
		return nil
	}
	c := mvd.trace.currentReadTraceReciever.addReadTrace(identifier)
	mvd.trace.currentReadTrace = c
	return c
}

func (mvd *Mvd) traceGetCurrentReadTrace() *TraceRead {
	if mvd.trace.enabled == false {
		return nil
	}
	return mvd.trace.currentReadTrace
}

type TraceReadAdd interface {
	addReadTrace(string) *TraceRead
}

type AdditionalInfo struct {
	Info  string
	Value interface{}
}

type TraceRead struct {
	Info                    string
	Identifier              string
	Value                   interface{}
	OffsetStart, OffsetStop uint
	AdditionalInfo          []AdditionalInfo
}

type TraceMessageTrace struct {
	Message     Message
	Type        SVC_TYPE
	Reads       []*TraceRead
	MessageData []byte
	currentRead *TraceRead
}

func (mt *TraceMessageTrace) addReadTrace(info string) *TraceRead {
	tr := new(TraceRead)
	mt.Reads = append(mt.Reads, tr)
	mt.currentRead = tr
	tr.Info = info
	return tr
}

type TraceReadIt struct {
	CurrentSize         int
	Reads               []*TraceRead
	MessageTrace        []*TraceMessageTrace
	currentMessageTrace *TraceMessageTrace
}

func (mvd *Mvd) getReadItTrace() *TraceReadIt {
	if mvd.trace.enabled == false {
		return nil
	}
	return mvd.trace.current.currentReadFrameTrace.currentReadIt
}

func (mvd *Mvd) getCurrentMessageTrace() *TraceMessageTrace {
	if mvd.trace.enabled == false {
		return nil
	}

	return mvd.trace.current.currentReadFrameTrace.currentReadIt.currentMessageTrace
}
func (mvd *Mvd) getCurrentMessageTraceReadTrace() *TraceRead {
	if mvd.trace.enabled == false {
		return nil
	}
	c := mvd.trace.current
	// this is stupid
	if c != nil {
		//return mvd.trace.current.currentReadFrameTrace.currentReadIt.currentMessageTrace.currentRead
		cc := c.currentReadFrameTrace
		if cc != nil {
			ccc := cc.currentReadIt
			if ccc != nil {
				cccc := ccc.currentMessageTrace
				if cccc != nil {
					if cccc.currentRead != nil {
						return cccc.currentRead
					}
				}
			}
		}
	}
	return nil
}

func (tri *TraceReadIt) addMessageTrace(message Message) *TraceMessageTrace {
	mt := new(TraceMessageTrace)
	mt.Message = message
	mt.MessageData = message.data
	tri.currentMessageTrace = mt
	tri.MessageTrace = append(tri.MessageTrace, mt)
	return mt
}

func (tri *TraceReadIt) addReadTrace(info string) *TraceRead {
	rt := new(TraceRead)
	rt.Info = info
	tri.Reads = append(tri.Reads, rt)
	return rt
}

func (tri *TraceReadIt) addMessageReadTrace(info string) *TraceRead {
	rt := new(TraceRead)
	rt.Info = info
	tri.Reads = append(tri.Reads, rt)
	return rt
}

type TraceReadFrame struct {
	DemoTime                        float64
	FileOffsetStart, FileOffsetStop uint
	ReadItTraces                    []*TraceReadIt
	currentReadIt                   *TraceReadIt
	Reads                           []*TraceRead
}

func (mvd *Mvd) startReadItTrace() *TraceReadIt {
	tri := new(TraceReadIt)
	trf := mvd.trace.current.currentReadFrameTrace
	trf.currentReadIt = tri
	trf.ReadItTraces = append(trf.ReadItTraces, tri)
	mvd.trace.currentReadTraceReciever = tri
	return tri
}

func (trace *TraceReadFrame) addReadTrace(info string) *TraceRead {
	ctrace := new(TraceRead)
	ctrace.Info = info
	trace.Reads = append(trace.Reads, ctrace)
	return ctrace
}

func (read *TraceRead) addAdditionalInfo(info string, value interface{}) {
	ai := AdditionalInfo{info, value}
	read.AdditionalInfo = append(read.AdditionalInfo, ai)
}

// mvd.ParseFrame trace
type TraceParseFrame struct {
	FrameStart, FrameStop uint
	LastFrame             bool
	State
	TraceReadFrame        []*TraceReadFrame
	currentReadFrameTrace *TraceReadFrame
	currentReadTrace      *TraceRead
}
