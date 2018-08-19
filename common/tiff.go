package common

import (
	"log"
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
    *save = ftell(ifp) + 4;
    if (*len * ("11124811248484"[*type < 14 ? *type:0]-'0') > 4)
		fseek (ifp, get4()+base, SEEK_SET);

*/

// GetTiff  legge da file info su tiff
func GetTiff(f *os.File, order uint16, base int64) (TiffInfo, int64) {

	var start, nextPos, save int64
	var tag, typ uint16
	var len uint32
	tag, start = GetUint16WithOrder(f, order, base)
	typ, start = GetUint16WithOrder(f, order, start)
	len, start = GetUint32WithOrder(f, order, start)
	save = start + 4

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

	return TiffInfo{Tag: tag, Type: typ, Save: save}, nextPos
}

// ParseTiff elaborazione TIFF?
func ParseTiff(f *os.File, base int64) {
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
		doff, start = GetUint32WithOrder(f, order, start)
		for doff > 0 {
			if parseTiffIfd(f, order, base+int64(doff), base) {
				return
			}

			doff, start = GetUint32(f, start)
		}

	}

}

// parseTiffIfd
func parseTiffIfd(f *os.File, order uint16, filePos int64, base int64) bool {

	var start int64
	var entries uint16
	var tiffInfo TiffInfo

	entries, start = GetUint16WithOrder(f, order, filePos)

	if entries > 512 {
		return true
	}

	for i := 0; i < int(entries); i++ {
		tiffInfo, start = GetTiff(f, order, start)
		switch tiffInfo.Tag {
		case 61440:
			// FUJI HS10 table
			parseTiffIfd(f, order, filePos, base)
			/*
			   /* Fuji HS10 table
			   fseek (ifp, get4()+base, SEEK_SET);
			   parse_tiff_ifd (base);
			   break;
			*/
		default:
			log.Printf("TIFF_PARSE_IFD  tag=%d", tiffInfo.Tag)
		}
	}

	return false
}
