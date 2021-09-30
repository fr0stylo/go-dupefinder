# go-dupefinder

Application finds duplicate files for the specified directory by computing hash of files. Results are displayed in console grouped by hash

```
dupefinder - cli tool that finds duplicate files on you file system

Usage:
  dupefinder [params] <root path>
Example:
  dupefinder -e node_modules ./project


Params:
  -e value
        List of excluded folders
  -h    see full help
  -p int
        sets paralelization level for hashing (default 10)
  -st int
        sets size threshold in kb
  -v    version
```

## Dependencies

project do not have any external dependency

## Build

if using windows:
`go build -o dupefinder.exe .`

if using linux:
`go build -o dupefinder .`


## Running

To run application find executable and execute following command in terminal:

`./dupefinder <root path for recursive search>`

