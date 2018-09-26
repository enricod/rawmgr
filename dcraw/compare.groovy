
def good = new File("./OUT")
def maybe = new File("../OUT")


LineNumberReader readerGood = good.newReader()
LineNumberReader readerMaybe = maybe.newReader()

while ((l1 = readerGood.readLine()) != null) {
    l2 = readerMaybe.readLine() 

    if (l2 != null) {
        def startLine = l2.split()[0]
        if (!l1.startsWith(startLine)) {
            println "ERRORE in ${l2} vs ${l1}"
            break
        }
    }
    
}
