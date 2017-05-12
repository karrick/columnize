# columnize

Like `column`, but right justifies columns that are all numbers.

## Usage

```
columnize < input
```

### Left Justify

To force all columns to be left-justified, add the `-l` argument.

```
columnize -l input.txt
```

### Right Justify

To force all columns to be right-justified, add the `-r` argument.

```
columnize -r input.txt
```

## Specifying a Delimiter

By default this program splits each line by whitespace. When the `-d
S` command line option is given, it uses the provided string as the
field delimiter. `S` may be a string of multiple characters.

```
columnize -d : input.txt
```

## Installation

If you don't have the Go programming language installed, then you'll
need to install a copy from http://golang.org.

Once you have Go installed:

```
go get github.com/karrick/columnize
```
