package common

import (
	"log"
	"os"
)

// TiffInfo info about tiff image
type TiffInfo struct {
	Tag  uint16
	Typ  uint16
	Len  uint32
	Save int64
}

type TiffIfd struct {
	Width      int64
	Height     int64
	Bps        int64 // bites per second?
	Comp       int
	Phint      int
	Offset     int
	Flip       int
	Samples    int
	Bytes      uint32
	TileWidth  int
	TileLength int
	Shutter    float32
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
		v, _ = GetUint32(f, start)
		nextPos = base + int64(v)
	} else {
		nextPos = start
	}

	return TiffInfo{Tag: tag, Typ: typ, Len: len, Save: save}, nextPos
}

// ParseTiff elaborazione TIFF?
func ParseTiff(f *os.File, base int64, tiffIfdArray []TiffIfd) []TiffIfd {
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

	ret2 := make([]TiffIfd, len(tiffIfdArray))
	copy(ret2, tiffIfdArray)

	var ret bool
	order, start = GetUint16(f, base)

	if order == 0x4949 || order == 0x4d4d {
		_, start = GetUint16(f, start)
		doff, start = GetUint32WithOrder(f, order, start)
		for doff > 0 {
			ret, ret2 = parseTiffIfd(f, order, base+int64(doff), base, ret2)
			if ret == true {
				return ret2
			}
			doff, start = GetUint32(f, start)
		}
	}
	return ret2
}

// MinInt64 min between two numbers
func MinInt64(a int64, b int64) int64 {
	if a < b {
		return a
	} else {
		return b
	}
}

// MaxInt64 max between two numbers
func MaxInt64(a int64, b int64) int64 {
	return MinInt64(b, a)
}

func parseMakernote(f *os.File, base int64) {

}
func parseExif(f *os.File, order uint16, filePos int64, base int64) {

	var start int64
	var entries uint16
	var tiffInfo TiffInfo

	entries, start = GetUint16WithOrder(f, order, filePos)
	log.Printf("TIFF_PARSE_IFD  entries=%d, start=%d", entries, start)

	for i := 0; i < int(entries); i++ {

		tiffInfo, start = GetTiff(f, order, start)
		switch tiffInfo.Tag {
		case 37500:
			parseMakernote(f, base)
		}
	}
}

// parseTiffIfd
func parseTiffIfd(f *os.File, order uint16, filePos int64, base int64, tiffIfdArray []TiffIfd) (bool, []TiffIfd) {

	var start int64
	var entries uint16
	var tiffInfo TiffInfo
	var tiffBps int64

	log.Printf("TIFF_PARSE_IFD  filePos=%d, base=%d", filePos, base)
	entries, start = GetUint16WithOrder(f, order, filePos)
	log.Printf("TIFF_PARSE_IFD  entries=%d, start=%d", entries, start)

	var tiffIfdNew = TiffIfd{}
	//tiffIfdArray2 := append(tiffIfdArray, tiffIfdNew)

	if entries > 512 {
		return false, tiffIfdArray
	}

	for i := 0; i < int(entries); i++ {
		tiffInfo, start = GetTiff(f, order, start)
		switch tiffInfo.Tag {
		case 34665: /* EXIF tag */
			var v uint32
			v, start = GetUint32WithOrder(f, order, start)
			parseExif(f, order, int64(v)+base, base)
			// fseek (ifp, get4()+base, SEEK_SET);
			// parse_exif (base);
			log.Printf("TIFF_PARSE_IFD  tag=%d", tiffInfo.Tag)

		case 61440: // Fuji HS10 table
			var v uint32
			v, start = GetUint32WithOrder(f, order, start)
			parseTiffIfd(f, order, int64(v)+base, base, append(tiffIfdArray, tiffIfdNew))
			/*
			   fseek (ifp, get4()+base, SEEK_SET);
			   parse_tiff_ifd (base);
			   break;
			*/
		case 61441: // image width
			/*
				case 61441:	/* ImageWidth
				tiff_ifd[ifd].width = getint(type);
				break;
			*/

			// FIXME eliminare indice 1 hardcoded

			tiffIfdNew.Width, start = GetInt(f, order, start, tiffInfo.Typ)

		case 61442: // image height
			tiffIfdNew.Height, start = GetInt(f, order, start, tiffInfo.Typ)
		case 258:
			fallthrough
		case 61443:
			/*
				tiff_ifd[ifd].samples = len & 7;
				if ((tiff_ifd[ifd].bps = getint(type)) > 32)
					tiff_ifd[ifd].bps = 8;
				if (tiff_bps < tiff_ifd[ifd].bps)
					tiff_bps = tiff_ifd[ifd].bps;
			*/
			tiffIfdNew.Samples = 1 & 7
			tiffIfdNew.Bps, start = GetInt(f, order, start, tiffInfo.Typ)
			if tiffIfdNew.Bps > 32 {
				tiffIfdNew.Bps = 8
			}
			tiffBps = MaxInt64(tiffIfdNew.Bps, tiffBps)

		case 61448:
			//tiff_ifd[ifd].bytes = get4();
			tiffIfdNew.Bytes, start = GetUint32WithOrder(f, order, start)

		default:
			// log.Printf("TIFF_PARSE_IFD  tag=%d", tiffInfo.Tag)
		}
		start = tiffInfo.Save
	}

	return false, append(tiffIfdArray, tiffIfdNew)
}
