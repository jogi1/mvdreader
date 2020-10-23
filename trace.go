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
	Message     *Message
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

func (mvd *Mvd) traceGetCurrentMessageTrace() *TraceMessageTrace {
	if mvd.trace.enabled == false {
		return nil
	}
	crt := mvd.getReadItTrace()
	if crt == nil {
		return nil
	}
	return crt.currentMessageTrace
}

func (mvd *Mvd) traceGetCurrentMessageTraceReadTrace() *TraceRead {
	if mvd.trace.enabled == false {
		return nil
	}
	cmt := mvd.traceGetCurrentMessageTrace()
	if cmt != nil {
		return cmt.currentRead
	}
	return nil
}

func (message *Message) traceAddMessageReadTrace(info string) {
	if message.mvd.trace.enabled == false {
		return
	}

	cmt := message.mvd.traceGetCurrentMessageTrace()
	cmt.addReadTrace(info)
}

func (message *Message) traceStartMessageReadTrace(identifier string, offsetStart, offsetStop *uint, value interface{}) {
	if message.mvd.trace.enabled == false {
		return
	}
	rt := message.mvd.traceGetCurrentMessageTraceReadTrace()
	if rt == nil {
		return
	}
	if len(rt.Identifier) == 0 {
		rt.Identifier = identifier
	} else {
		if rt.Identifier != identifier {
			return
		}
	}
	if offsetStart != nil {
		rt.OffsetStart = *offsetStart
	}

	if offsetStop != nil {
		rt.OffsetStop = *offsetStop
	}

	if value != nil {
		rt.Value = value
	}
}

func (tri *TraceReadIt) addMessageTrace(message *Message) *TraceMessageTrace {
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

func (mvd *Mvd) traceStartReadItTrace() {
	if mvd.trace.enabled == false {
		return
	}
	tri := new(TraceReadIt)
	trf := mvd.trace.current.currentReadFrameTrace
	trf.currentReadIt = tri
	trf.ReadItTraces = append(trf.ReadItTraces, tri)
	mvd.trace.currentReadTraceReciever = tri
}

func (mvd *Mvd) traceReadItTraceCurrentSize(current_size int) {
	if mvd.trace.enabled == false {
		return
	}
	mvd.trace.current.currentReadFrameTrace.currentReadIt.CurrentSize = current_size
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

func (mvd *Mvd) traceAddReadTrace(identifier string) {
	if mvd.trace.enabled == false {
		return
	}
	mvd.trace.currentReadTrace = mvd.trace.currentReadTraceReciever.addReadTrace(identifier)
}

func (mvd *Mvd) traceReadTraceAdditionalInfo(info string, value interface{}) {
	if mvd.trace.enabled == false {
		return
	}
	ai := AdditionalInfo{info, value}
	mvd.trace.currentReadTrace.AdditionalInfo = append(mvd.trace.currentReadTrace.AdditionalInfo, ai)
}

func (mvd *Mvd) traceStartReadTrace(identifier string, offsetStart, offsetStop *uint, value interface{}) {
	if mvd.trace.enabled == false {
		return
	}
	rt := mvd.traceGetCurrentReadTrace()
	if rt == nil {
		return
	}
	if len(rt.Identifier) == 0 {
		rt.Identifier = identifier
	} else {
		if rt.Identifier != identifier {
			return
		}
	}
	if offsetStart != nil {
		rt.OffsetStart = *offsetStart
	}

	if offsetStop != nil {
		rt.OffsetStop = *offsetStop
	}

	if value != nil {
		rt.Value = value
	}
}

func (mvd *Mvd) traceStartMessageTrace(message *Message) {
	if mvd.trace.enabled == false {
		return
	}
	rit := mvd.getReadItTrace()
	rit.currentMessageTrace = rit.addMessageTrace(message)
}

func (mvd *Mvd) traceMessageInfo(msgType SVC_TYPE) {
	if mvd.trace.enabled == false {
		return
	}
	rit := mvd.getReadItTrace()
	cmt := rit.currentMessageTrace
	cmt.Type = msgType
}

func (mvd *Mvd) traceStartMessageTraceReadTrace(info string) {
	if mvd.trace.enabled == false {
		return
	}
	rit := mvd.getReadItTrace()
	rit.currentMessageTrace.addReadTrace(info)
}

func (message *Message) traceMessageReadAdditionlInfo(info string, value interface{}) {
	if message.mvd.trace.enabled == false {
		return
	}
	rit := message.mvd.traceGetCurrentMessageTraceReadTrace()
	if rit == nil {
		return
	}
	rit.addAdditionalInfo(info, value)
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
