# bashutils-go

A Go-based reimplementation of bash coreutils with bash-like path globbing support.

## Features

- **Bash-like Path Globbing**: All file-operating commands support bash-style glob patterns (`*`, `?`, `[...]`)
- **Core Utilities**: Implements common Unix utilities like `cat`, `head`, `tail`, `wc`, `grep`, `sort`, `uniq`, `cut`, `paste`, `split`, `tr`, and `echo`

## Globbing Support

All commands that accept file arguments now support bash-like glob patterns:

- `*` - matches any sequence of characters
- `?` - matches any single character  
- `[...]` - matches any character within the brackets

### Examples

```bash
# Process all text files
./bashutils cat "test_files/*.txt"

# Process files with specific patterns
./bashutils head "test_files/file?.txt"

# Multiple glob patterns
./bashutils wc "test_files/*.txt" "*.go"

# Process files in subdirectories
./bashutils grep "pattern" "**/*.txt"
```

## Commands

### cat
Concatenate and display files with glob support.

```bash
./bashutils cat "*.txt"
```

### head
Output the first part of files with glob support.

```bash
./bashutils head -n 5 "*.txt"
```

### tail
Output the last part of files with glob support.

```bash
./bashutils tail -n 10 "*.txt"
```

### wc
Print newline, word, and byte counts for files with glob support.

```bash
./bashutils wc "*.txt"
```

### grep
Print lines matching a pattern with glob support.

```bash
./bashutils grep "pattern" "*.txt"
```

### sort
Sort lines of text files with glob support.

```bash
./bashutils sort "*.txt"
```

### uniq
Filter out repeated lines with glob support.

```bash
./bashutils uniq "*.txt"
```

### cut
Extract specific columns or byte ranges from lines with glob support.

```bash
./bashutils cut -f 1,3 "*.csv"
```

### paste
Merge lines from files with glob support.

```bash
./bashutils paste "file1.txt" "file2.txt"
```

### split
Split files into pieces with glob support.

```bash
./bashutils split -l 1000 "large_file.txt"
```

### tr
Translate or delete characters (works with stdin).

```bash
echo "hello" | ./bashutils tr "a-z" "A-Z"
```

### echo
Echo arguments to standard output.

```bash
./bashutils echo "Hello, World!"
```

## Installation

```bash
go build -o bashutils
```

## Usage

```bash
./bashutils [command] [options] [arguments]
```

## Globbing Behavior

- **No matches**: If a glob pattern doesn't match any files, the original pattern is preserved (bash behavior)
- **Multiple matches**: Commands process all matching files
- **File validation**: Commands validate that files exist and are readable before processing
- **Sorted output**: Glob matches are sorted for consistent output
