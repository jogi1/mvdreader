package mvdreader

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

func (mvd *Mvd) demotimeReadahead() (error, float64) {
	err, b := mvd.readByteAhead()
	if err != nil {
		return err, 0
	}
	return nil, float64(b)
}

func (mvd *Mvd) demotime() error {
	mvd.traceAddReadTrace("demotime")
	err, b := mvd.readByte()
	if err != nil {
		return err
	}

	mvd.traceReadTraceAdditionalInfo("demotime_change", float64(b)*0.001)
	mvd.demo.time += float64(b) * 0.001
	if mvd.debug != nil {
		mvd.debug.Printf("time (%v)", mvd.demo.time)
	}
	return nil
}

func (mvd *Mvd) readBytes(count uint) (error, *bytes.Buffer) {
	if mvd.debug != nil {
		mvd.debug.Println("------------- READBYTES: ", mvd.getInfo(count), count)
	}
	if mvd.file_offset+count > mvd.file_length {
		return errors.New("readBytes: trying to read beyond"), nil
	}

	mvd.traceStartReadTrace("readBytes", &mvd.file_offset, nil, nil)
	b := bytes.NewBuffer(mvd.file[mvd.file_offset : mvd.file_offset+count])
	mvd.file_offset += count
	mvd.traceStartReadTrace("readBytes", nil, &mvd.file_offset, b)
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
	mvd.traceStartReadTrace("readByte", &mvd.file_offset, nil, nil)
	if mvd.file_offset+1 > mvd.file_length {
		return errors.New("readByte: trying to read beyond"), byte(0)
	}
	b := mvd.file[mvd.file_offset]
	mvd.file_offset += 1
	mvd.traceStartReadTrace("readByte", nil, &mvd.file_offset, b)
	return nil, b
}

func (mvd *Mvd) readInt() (error, int32) {
	var i int32
	mvd.traceStartReadTrace("readInt", &mvd.file_offset, nil, nil)
	err, b := mvd.readBytes(4)
	if err != nil {
		return err, 0
	}

	err = binary.Read(b, binary.LittleEndian, &i)
	if err != nil {
		return err, 0
	}

	mvd.traceStartReadTrace("readInt", nil, &mvd.file_offset, i)
	return nil, i
}

func (mvd *Mvd) readUint() (error, uint32) {
	var i uint32
	mvd.traceStartReadTrace("readUint", &mvd.file_offset, nil, nil)
	err, b := mvd.readBytes(4)
	if err != nil {
		return err, 0
	}
	err = binary.Read(b, binary.LittleEndian, &i)
	if err != nil {
		return err, 0
	}

	mvd.traceStartReadTrace("readUint", nil, &mvd.file_offset, i)
	return nil, i
}

func (mvd *Mvd) readIt(cmd DEM_TYPE) (error, bool) {
	mvd.traceStartReadItTrace()
	err, i := mvd.readUint()
	if err != nil {
		return err, false
	}

	current_size := int(i)
	mvd.traceReadItTraceCurrentSize(current_size)

	if current_size == 0 {
		if mvd.debug != nil {
			mvd.debug.Println("ReadIt: current size 0 go to next Frame! <----------")
		}
		return nil, false
	}
	old_offset := mvd.file_offset
	mvd.file_offset += uint(current_size)
	if mvd.debug != nil {
		mvd.debug.Printf("------------- moving ahead %v from (%v) to (%v) filesize: %v", current_size, old_offset, mvd.file_offset, len(mvd.file))
	}
	if mvd.file_offset > mvd.file_length {
		return fmt.Errorf("offset (%d) larger than filesize (%d)\n", mvd.file_offset, mvd.file_length), false
	}
	if mvd.demo.last_type == dem_multiple && mvd.demo.last_to == 0 {
		if mvd.debug != nil {
			mvd.debug.Println("ReadIt: we skip this message?")
		}
		return err, false
	}
	message := Message{
		size:        uint(current_size),
		data:        mvd.file[old_offset:mvd.file_offset],
		OffsetStart: old_offset}
	err, fullRead := mvd.messageParse(message)
	if err != nil {
		return err, false
	}
	if fullRead {
		return nil, false
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
