# columnize

Like `column -t`, but right justifies columns that are all numbers.

## Usage

This program reads from the files specified on the command line after
all the flags have been processed, or will read from standard input
when no files are specified. The following two invocations will have
the same effect:

    $ columnize < sample.txt
    $ columnize sample.txt

However, more than a single file can be provided on the command line:

    $ columnize benchmarks-a.out benchmarks-b.out

By default this program inspects fields on every line to determine max
field width and whether the field is numerical. By default numerical
fields are right justified and non-numerical fields are left
justified.

### Skip Header

When the `-s` command line option is provided this program still uses
the first line of input for column width determination, but ignores
the fields in the first line when determining whether the column is
numerical.

### Left Justify

When the `-l` command line option is provided, all columns will be
left justified.

```
columnize -l input.txt
```

### Right Justify

When the `-r` command line option is provided, all columns will be
right justified.

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
need to install a copy from
[https://golang.org/dl](https://golang.org/dl).

Once you have Go installed:

```
go get github.com/karrick/columnize
```
