// Code generated by "stringer -type=CompressionType"; DO NOT EDIT.

package vol

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[None-0]
	_ = x[RLE-1]
	_ = x[LZ-2]
	_ = x[LZH-3]
}

const _CompressionType_name = "NoneRLELZLZH"

var _CompressionType_index = [...]uint8{0, 4, 7, 9, 12}

func (i CompressionType) String() string {
	if i >= CompressionType(len(_CompressionType_index)-1) {
		return "CompressionType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _CompressionType_name[_CompressionType_index[i]:_CompressionType_index[i+1]]
}
