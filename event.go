package mvdreader

//go:generate
//stringer -type=PE_TYPE
type PE_TYPE int

const (
	PE_MOVEMENT    PE_TYPE = 1 << 1
	PE_STATS       PE_TYPE = 1 << 2
	PE_ANIMATION   PE_TYPE = 1 << 3
	PE_NETWORKINFO PE_TYPE = 1 << 4
	PE_USERINFO    PE_TYPE = 1 << 5
)
