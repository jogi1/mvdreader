// Code generated by "stringer -type=DEM_TYPE"; DO NOT EDIT.

package mvdreader

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[dem_cmd-0]
	_ = x[dem_read-1]
	_ = x[dem_set-2]
	_ = x[dem_multiple-3]
	_ = x[dem_single-4]
	_ = x[dem_stats-5]
	_ = x[dem_all-6]
}

const _DEM_TYPE_name = "dem_cmddem_readdem_setdem_multipledem_singledem_statsdem_all"

var _DEM_TYPE_index = [...]uint8{0, 7, 15, 22, 34, 44, 53, 60}

func (i DEM_TYPE) String() string {
	if i >= DEM_TYPE(len(_DEM_TYPE_index)-1) {
		return "DEM_TYPE(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _DEM_TYPE_name[_DEM_TYPE_index[i]:_DEM_TYPE_index[i+1]]
}
