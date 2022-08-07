# All-in-one VOL file manipulator Starsiege and Starsiege: Tribes

A tool to inspect, create, and extract .vol files for Starsiege and Starsiege: Tribes. Run without arguments for usage.
Should work on Windows, Mac, and *nix (only tested on Windows so far).

## Examples

### Info (list files in vol)
```
$ vol.exe info my.vol
my.vol contains 3 files:
file1.txt:      14 bytes        (compression: None)
file2.txt:      14 bytes      (compression: None)
dir\file3.txt:  14 bytes       (compression: None)
```

### Dump (quick look at files in vol without extracting)
```
$ vol.exe dump my.vol
*** file1.txt ***
hello world1
*** file2.txt ***
hello world2
*** file3.txt ***
hello world3

$ vol.exe dump my.vol file1.txt
*** file1.txt ***
hello world1
```

### Unpack (extract files from vol)
```
$ vol.exe unpack my.vol
unpacked file1.txt
unpacked file2.txt
unpacked dir\file3.txt

$ vol.exe unpack my.vol outdir
unpacked file1.txt to outdir\file1.txt
unpacked file2.txt to outdir\file2.txt
unpacked dir\file3.txt to outdir\dir\file3.txt

$ vol.exe unpack my.vol --strip-paths
unpacked file1.txt
unpacked file2.txt
unpacked dir\file3.txt to file3.txt

$ vol.exe unpack my.vol . file1.txt file2.txt
unpacked file1.txt
unpacked file2.txt

$ vol.exe unpack my.vol . d*\file*.txt
unpacked dir\file3.txt
```

### Pack (create or add files to vol)
```
$ vol.exe pack new.vol fileA.txt fileB.txt dir\fileC.txt
packed fileA.txt
packed fileB.txt
packed dir\fileC.txt

$ vol.exe pack new.vol fileD.txt
packed fileD.txt

$ vol.exe pack new.vol fileD.txt
Error: file fileD.txt already exists in vol file new.vol; use --overwrite to overwrite
...

$ vol.exe pack new.vol fileD.txt --overwrite
packed fileD.txt

$ vol.exe pack new.vol dir\fileC.txt --strip-paths
packed dir\fileC.txt (as fileC.txt)
```

## Building
```
go build -o vol.exe github.com/iambob314/vol/cmd
```

## Limitations
* Cannot read content of LZH- or RLE-compressed files in vol
  * _Can_ inspect such files with `vol.exe info`
  * _Can_ add new files to an existing vol containing such files with `vol.exe pack` 