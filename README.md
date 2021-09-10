# gounzip
Unzip to match gozip

## Arguments

* `gounzip -h` - print usage information
* `gounzip -l <zipfile>` - list contents of zipfile
* `gounzip <zipfile>` - unzip zipfile
* `gounzip -u <zipfile>` - unzip zipfile adding new file but only update newer versions
* `gounzip -f <zipfile>` - unzip zipfile but only update newer versions of existing
* `gounzip -d <zipfile>` - unzip zipfile to destination as root
* `gozip -f <zipfile> <file>...` - only update newer files already in archive

## Usage

To build

`go build .`

To run tests

`go test -v .`

## Notes

Here is a sample list

```
  compressed uncompressed      date       time        name
---------------------------------------------------------------------------
        3495        7210   2021-09-08  19:02:28  sample/1.txt
        2330        4621   2021-09-08  19:02:28  sample/2.txt
        1174        2178   2021-09-08  19:02:28  sample/3.txt
        1021        1827   2021-09-08  19:02:28  sample/4.txt
         497         918   2021-09-08  19:02:28  sample/5.txt
        3495        7210   2021-09-07  23:26:00  sample/orig/1.txt
        2330        4621   2021-09-07  23:26:00  sample/orig/2.txt
        1174        2178   2021-09-07  23:26:00  sample/orig/3.txt
        1021        1827   2021-09-07  23:26:00  sample/orig/4.txt
         497         918   2021-09-07  23:26:00  sample/orig/5.txt
---------------------------------------------------------------------------
       17034       33508                         10
```
