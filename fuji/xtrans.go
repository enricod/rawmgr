package fuji

import (
	"log"
	"os"

	"github.com/enricod/rawmgr/common"
)

func XtransInterpolate(passes int) {

}

func ParseFuji(f *os.File, offset int64) {

	var tag, len uint16
	var start, posizione int64
	var rawHeight, rawWidth uint16
	var height, fujiLayout uint16
	//var width uint32
	//var fujiWidth uint32
	var filters int
	var xtransAbs [6][6]uint16

	start = offset
	entries, start := common.GetUint32(f, start)

	log.Printf("entries %d \n", entries)
	if entries < 255 {

		for i := 0; i < int(entries); i++ {
			/*
				tag = get2();
				len = get2();
				save = ftell(ifp);
			*/
			tag, start = common.GetUint16(f, start)
			len, start = common.GetUint16(f, start)
			posizione = start
			log.Printf("posizione=%d, tag=%d, len=%d", posizione, tag, len)
			switch tag {
			case 0x100:
				rawHeight, start = common.GetUint16(f, start)
				rawWidth, start = common.GetUint16(f, start)
				log.Printf("raw_width=%d, raw_height=%d", rawHeight, rawWidth)

			case 0x121:
				height, start = common.GetUint16(f, start)
				log.Printf("height=%d", height)

			case 0x130:
				fujiLayout, start = common.GetUint16(f, start)
				fujiLayout = fujiLayout >> 7
				log.Printf("fujiLayout=%d", fujiLayout)
			//fujiWidth = !(fgetc(ifp) & 8)
			case 0x131:
				filters = 9
				var val uint16
				for r := 5; r >= 0; r-- {
					for c := 5; c >= 0; c-- {
						val, start = common.Get1Byte(f, start)
						xtransAbs[r][c] = val & 3
					}
				}
				log.Printf("filters=%d, xtransAbs=%v", filters, xtransAbs)
				/*
					filters = 9;
					FORC(36) xtrans_abs[0][35-c] = fgetc(ifp) & 3;
				*/

			case 0x2ff0:
				log.Printf("WARN tag non ancora elaborato=%d", tag)
			case 0xc000:
				if len > 20000 {
					/*
						c = order;
						order = 0x4949;
						while ((tag = get4()) > raw_width);
						width = tag;
						height = get4();
						order = c;
					*/

				}
			default:

			}

			start = posizione + int64(len)
		}

	}

}
