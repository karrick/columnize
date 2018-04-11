# columnize

Like `column -t`, but right justifies numerical fields, and can be
configured to ignore header and footer lines for the purpose of
determining field widths.

## Usage

This program reads from the files specified on the command line after
all the flags have been processed, or will read from standard input
when no files are specified. The following two invocations will have
the same effect:

    $ columnize < sample.txt
    $ columnize sample.txt

However, more than a single file can be provided on the command line:

    $ columnize benchmarks-a.out benchmarks-b.out

### Header and Footer

By default this program inspects fields on every line to determine max
field width. When the `--header N` flag is provided, this program
instead blindly copies the first N lines of input directly to its
standard output without checking column widths.

Similarly, when the `--footer N` flag is provided, this program
blindly copies the final N lines of input directly to its standard
output without checking column widths.

Compare the output of the following two commands.

    $ columnize testdata/bench.out
    $ columnize --header 3 --footer 2 testdata/bench.out

### Skip Header (deprecated: see --header flag)

The `--skip-header, -s` flag behaves exactly as if the `--header 1`
flag were provided.

### Left Justify

When the `-l` command line option is provided, all columns will be
left justified.

    $ columnize -l input.txt

### Right Justify

When the `-r` command line option is provided, all columns will be
right justified.

    $ columnize -r input.txt

## Output Formating Delimiter

By default this program uses a minimum of two space characters between
the columns in the output. The `--delimiter, -d` flag may be used to
specify a different joining string between output columns. This string
may be one or more characters.

    $ columnize -d " | " input.txt

## Installation

If you don't have the Go programming language installed, then you'll
need to install a copy from
[https://golang.org/dl](https://golang.org/dl).

Once you have Go installed:

    $ go get github.com/karrick/columnize
