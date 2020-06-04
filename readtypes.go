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
	mvd.Demo.Time += float64(b) * 0.0001
	if mvd.Debug != nil {
		mvd.Debug.Printf("time (%v)", mvd.Demo.Time)
	}
	return nil
}

func (mvd *Mvd) readBytes(count uint) (error, *bytes.Buffer) {
	if mvd.Debug != nil {
		mvd.Debug.Println("------------- READBYTES: ", mvd.getInfo(count), count)
	}
	if mvd.file_offset+count > mvd.file_length {
		return errors.New("readBytes: trying to read beyond"), nil
	}
	b := bytes.NewBuffer(mvd.file[mvd.file_offset : mvd.file_offset+count])
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
	if mvd.Debug != nil {
		mvd.Debug.Println("------------- READBYTE: ", mvd.getInfo(1))
	}
	if mvd.file_offset+1 > mvd.file_length {
		return errors.New("readByte: trying to read beyond"), byte(0)
	}
	b := mvd.file[mvd.file_offset]
	mvd.file_offset += 1
	return nil, b
}

func (mvd *Mvd) readInt() (error, int32) {
	var i int32
	err, b := mvd.readBytes(4)
	if err != nil {
		return err, 0
	}

	err = binary.Read(b, binary.LittleEndian, &i)
	if err != nil {
		return err, 0
	}
	return nil, i
}

func (mvd *Mvd) readUint() (error, uint32) {
	var i uint32
	err, b := mvd.readBytes(4)
	if err != nil {
		return err, 0
	}
	err = binary.Read(b, binary.LittleEndian, &i)
	if err != nil {
		return err, 0
	}
	return nil, i
}

func (mvd *Mvd) readIt(cmd DEM_TYPE) (error, bool) {
	err, i := mvd.readUint()
	if err != nil {
		return err, false
	}

	current_size := int(i)
	if current_size == 0 {
		if mvd.Debug != nil {
			mvd.Debug.Println("ReadIt: current size 0 go to next Frame! <----------")
		}
		return nil, false
	}
	old_offset := mvd.file_offset
	mvd.file_offset += uint(current_size)
	if mvd.Debug != nil {
		mvd.Debug.Printf("------------- moving ahead %v from (%v) to (%v) filesize: %v", current_size, old_offset, mvd.file_offset, len(mvd.file))
	}
	err = mvd.messageParse(Message{size: uint(current_size), data: mvd.file[old_offset:mvd.file_offset]})
	if err != nil {
		return err, false
	}
	if mvd.Demo.last_type == dem_multiple {
		if mvd.Debug != nil {
			mvd.Debug.Println("looping")
		}
		return nil, true
	}
	if mvd.Debug != nil {
		mvd.Debug.Println("ReadIt: go to next Frame! <----------")
	}
	return nil, false
}
