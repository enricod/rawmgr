nothing, yet

Lettura RAF

main 



    if (filters || colors == 1)

        imposta raw_image

        entra in if (raw_image) #13111

        fuji_rotate()
        convert_to_rgb()
        stretch()
        

    status = identify()
        calcola thumb_offset
        calcole thaumb_position
        parse_tiff (data_offset = get4());
        parse_tiff (thumb_offset+12);
        apply_tiff();
        parse_fuji( 4 bytest)
            cerca num entries
            per ogni entry
                tag
                len
                tag == 0x131
                    trova filters
                    popola xtrans_abs

                tag == 0x4949
                    assegna order

            
            

