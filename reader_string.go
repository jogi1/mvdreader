package mvdreader
import(
    "bytes"
    "strings"
    "fmt"
)

type ReaderString struct {
    String string `json:"string"`
    Byte []byte `json:"byte"`
}

func (ps *ReaderString) Equal(c interface{}) bool {
    switch v:= c.(type) {
    case string:
        return ps.String == v
    case []byte:
        return bytes.Equal(ps.Byte, v)
    case ReaderString:
        return bytes.Equal(ps.Byte, v.Byte)
    case *ReaderString:
        if v == nil {
            return false
        }
        return bytes.Equal(ps.Byte, v.Byte)
    }
    return false
}

func ReaderStringNew(b []byte, ascii_table []rune) *ReaderString {
	s := new(ReaderString)
    s.Byte = b
	var sb strings.Builder
	for _, bi := range b {
		fmt.Fprintf(&sb, "%c", ascii_table[uint(bi)])
	}
    s.String = sb.String()
	return s
}



func AsciiTableInit(ascii_table_in *string) []rune {
    ascii_table := "________________[]0123456789____ !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_'abcdefghijklmnopqrstuvwxyz{|}~_________________[]0123456789____ !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_'abcdefghijklmnopqrstuvwxyz{|}~_"
    if ascii_table_in != nil {
        ascii_table = *ascii_table_in
    }
	return []rune(ascii_table)
}

