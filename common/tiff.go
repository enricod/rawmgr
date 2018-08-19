package common

import (
	"os"
)

/*
	int CLASS parse_tiff (int base)
	{
		int doff;

		fseek (ifp, base, SEEK_SET);
		order = get2();
		if (order != 0x4949 && order != 0x4d4d) return 0;
		get2();
		while ((doff = get4()))
		{
			fseek (ifp, doff+base, SEEK_SET);
			if (parse_tiff_ifd (base)) break;
		}
		return 1;
	}
*/

type TiffInfo struct {
	Tag  uint16
	Type uint16
	Len  uint32
	Save int64
}

/*
    *tag  = get2();
    *type = get2();
	*len  = get4();
    *save = ftell(ifp) + 4;s
    if (*len * ("11124811248484"[*type < 14 ? *type:0]-'0') > 4)
		fseek (ifp, get4()+base, SEEK_SET);

*/

// GetTiff  legge da file info su tiff
func GetTiff(f *os.File, base int64) (TiffInfo, int64) {

	var start, nextPos int64
	var tag, typ uint16
	var len uint32
	tag, start = GetUint16(f, base)
	typ, start = GetUint16(f, start)
	len, start = GetUint32(f, start)

	var idx uint16
	if typ < 14 {
		idx = typ
	} else {
		idx = 0
	}
	c := int("11124811248484"[idx]) - int('0')

	if len*uint32(c) > 4 {
		var v uint32
		v, start = GetUint32(f, start)
		nextPos = base + int64(v)
	} else {
		nextPos = base
	}

	return TiffInfo{Tag: tag, Type: typ, Save: base + 4}, nextPos
}

// ParseTIFF elaborazione TIFF?
func ParseTIFF(f *os.File, base int64) {
	/*
		int CLASS parse_tiff (int base) {
		    int doff;

		    fseek (ifp, base, SEEK_SET);
		    order = get2();
		    if (order != 0x4949 && order != 0x4d4d) return 0;
		    get2();
		    while ((doff = get4()))
		    {
		        fseek (ifp, doff+base, SEEK_SET);
		        if (parse_tiff_ifd (base)) break;
		    }
		    return 1;
		}
	*/
	var order uint16
	var doff uint32
	var start int64

	order, start = GetUint16(f, base)

	if order == 0x4949 || order == 0x4d4d {
		_, start = GetUint16(f, start)
		doff, start = GetUint32(f, start)
		for doff > 0 {
			start = start + int64(doff)
			if parseTiffIfd(base) {
				return
			}

			doff, start = GetUint32(f, start)
		}

	}

}

// parse_tiff_ifd
func parseTiffIfd(base int64) bool {
	return false
}
