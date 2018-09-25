
dcraw sources by Dave Coffin, a real genius

https://www.cybercom.net/~dcoffin/dcraw/



gcc -o dcraw -O4 dcraw.c -lm -DNODEPS

per debug

```
gcc -g -o dcraw  dcraw.c -lm -DNODEPS
```


# DEBUG

gdb --args dcraw ../images/Canon/Canon_001.CR2

b <riga>
run
n  next
print <variabile>

