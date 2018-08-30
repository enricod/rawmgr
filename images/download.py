
import os.path

baseUrl = "https://raw.pixls.us/getfile.php"
cameras = ["/799/nice/Canon%20-%20EOS%206D%20-%20RAW.CR2", "/1680/nice/Fujifilm%20-%20X-Pro2%20-%2014bit%20uncompressed%20\(3:2\).RAF", "/2156/nice/Sony%20-%20DSC-RX100M3%20-%2012bit%20compressed%20\(3:2\).ARW"]
for c in cameras:
    f = c.split("/")[-1]
    bashCommand = "wget " + baseUrl + c
    os.system(bashCommand)
