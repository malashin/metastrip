// Code generated by "stringer -type=Marker"; DO NOT EDIT.

package jpg

import "strconv"

const _Marker_name = "SOF0SOF1SOF2SOF3DHTSOF5SOF6SOF7JPGSOF9SOF10SOF11DACSOF13SOF14SOF15RST0RST1RST2RST3RST4RST5RST6RST7SOIEOISOSDQTDNLDRIDHPEXPAPP0APP1APP2APP3APP4APP5APP6APP7APP8APP9APP10APP11APP12APP13APP14APP15JPG0JPG1JPG2JPG3JPG4JPG5JPG6JPG7JPG8JPG9JPG10JPG11JPG12JPG13COM"

var _Marker_index = [...]uint8{0, 4, 8, 12, 16, 19, 23, 27, 31, 34, 38, 43, 48, 51, 56, 61, 66, 70, 74, 78, 82, 86, 90, 94, 98, 101, 104, 107, 110, 113, 116, 119, 122, 126, 130, 134, 138, 142, 146, 150, 154, 158, 162, 167, 172, 177, 182, 187, 192, 196, 200, 204, 208, 212, 216, 220, 224, 228, 232, 237, 242, 247, 252, 255}

func (i Marker) String() string {
	i -= 192
	if i >= Marker(len(_Marker_index)-1) {
		return "Marker(" + strconv.FormatInt(int64(i+192), 10) + ")"
	}
	return _Marker_name[_Marker_index[i]:_Marker_index[i+1]]
}
