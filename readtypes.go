package mvdreader

import (
	"bytes"
	"encoding/binary"
	"errors"
)

func (mvd *Mvd) demotimeReadahead() (error, float64) {
	err, b := mvd.readByteAhead()
	if err != nil {
		return err, 0
	}
	return nil, float64(b)
}

func (mvd *Mvd) demotime() error {
	err, b := mvd.readByte()
	if err != nil {
		return err
	}

	rt := mvd.traceAddReadTrace("demotime")
	if rt != nil {
		rt.addAdditionalInfo("demotime_change", float64(b)*0.0001)
	}
	mvd.demo.time += float64(b) * 0.0001
	if mvd.debug != nil {
		mvd.debug.Printf("time (%v)", mvd.demo.time)
	}
	return nil
}

func (mvd *Mvd) readBytes(count uint) (error, *bytes.Buffer) {
	first := true
	if mvd.debug != nil {
		mvd.debug.Println("------------- READBYTES: ", mvd.getInfo(count), count)
	}
	if mvd.file_offset+count > mvd.file_length {
		return errors.New("readBytes: trying to read beyond"), nil
	}

	rt := mvd.traceGetCurrentReadTrace()
	if rt != nil {
		// this is called by other read functions too
		if len(rt.Identifier) == 0 {
			rt.Identifier = "readBytes"
			first = false
		} else {
			rt.OffsetStart = mvd.file_offset
		}
	}
	b := bytes.NewBuffer(mvd.file[mvd.file_offset : mvd.file_offset+count])

	if rt != nil {
		if first {
			rt.Value = b
			rt.OffsetStart = mvd.file_offset
			rt.OffsetStop = mvd.file_offset + 4
		}
	}
	mvd.file_offset += count
	return nil, b
}

func (mvd *Mvd) getInfo(a ...interface{}) string {
	return ""
}

func (mvd *Mvd) readByteAhead() (error, byte) {
	if mvd.file_offset+1 > mvd.file_length {
		return errors.New("readByteAhead: trying to read beyond"), byte(0)
	}
	b := mvd.file[mvd.file_offset]
	return nil, b
}

func (mvd *Mvd) readByte() (error, byte) {
	if mvd.debug != nil {
		mvd.debug.Println("------------- READBYTE: ", mvd.getInfo(1))
	}
	rt := mvd.traceGetCurrentReadTrace()
	if rt != nil {
		rt.Identifier = "readByte"
	}
	if mvd.file_offset+1 > mvd.file_length {
		return errors.New("readByte: trying to read beyond"), byte(0)
	}
	b := mvd.file[mvd.file_offset]
	if rt != nil {
		rt.Value = b
		rt.OffsetStart = mvd.file_offset
	}
	mvd.file_offset += 1
	if rt != nil {
		rt.OffsetStop = mvd.file_offset
	}
	return nil, b
}

func (mvd *Mvd) readInt() (error, int32) {
	var i int32
	rt := mvd.traceGetCurrentReadTrace()
	if rt != nil {
		rt.Identifier = "readInt"
		rt.OffsetStart = mvd.file_offset
	}
	err, b := mvd.readBytes(4)
	if err != nil {
		return err, 0
	}

	err = binary.Read(b, binary.LittleEndian, &i)
	if err != nil {
		return err, 0
	}

	if rt != nil {
		rt.Value = i
		rt.OffsetStop = mvd.file_offset
	}
	return nil, i
}

func (mvd *Mvd) readUint() (error, uint32) {
	var i uint32
	rt := mvd.traceGetCurrentReadTrace()
	if rt != nil {
		rt.Identifier = "readUint"
		rt.OffsetStart = mvd.file_offset
	}
	err, b := mvd.readBytes(4)
	if err != nil {
		return err, 0
	}
	err = binary.Read(b, binary.LittleEndian, &i)
	if err != nil {
		return err, 0
	}

	if rt != nil {
		rt.Value = i
		rt.OffsetStop = mvd.file_offset
	}
	return nil, i
}

func (mvd *Mvd) readIt(cmd DEM_TYPE) (error, bool) {
	rit := mvd.startReadItTrace()
	mvd.traceAddReadTrace("current_size")
	err, i := mvd.readUint()
	if err != nil {
		return err, false
	}

	current_size := int(i)
	if rit != nil {
		rit.CurrentSize = current_size
	}

	if current_size == 0 {
		if mvd.debug != nil {
			mvd.debug.Println("ReadIt: current size 0 go to next Frame! <----------")
		}
		if rit != nil {
		}
		return nil, false
	}
	old_offset := mvd.file_offset
	mvd.file_offset += uint(current_size)
	if mvd.debug != nil {
		mvd.debug.Printf("------------- moving ahead %v from (%v) to (%v) filesize: %v", current_size, old_offset, mvd.file_offset, len(mvd.file))
	}
	err = mvd.messageParse(Message{size: uint(current_size), data: mvd.file[old_offset:mvd.file_offset]})
	if err != nil {
		return err, false
	}
	if mvd.demo.last_type == dem_multiple {
		if mvd.debug != nil {
			mvd.debug.Println("looping")
		}
		return nil, true
	}
	if mvd.debug != nil {
		mvd.debug.Println("ReadIt: go to next Frame! <----------")
	}
	return nil, false
}
