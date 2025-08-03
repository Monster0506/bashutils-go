# bashutils

[![Go Report Card](https://goreportcard.com/badge/github.com/monster0506/bashutils-go)](https://goreportcard.com/report/monster0506/bashutils-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go-based reimplementation of bash coreutils with bash-like path globbing support.

## The Frustration is Real: Bringing Unix Comfort to Windows

Let's be honest. If you've spent any significant time in a Linux or macOS terminal,
then tried to do something meaningful in the Windows Command Prompt or PowerShell,
you know the feeling. The common, indispensable utilities like `grep`, `cut`,
`sort`, `uniq`, `cat`, `head`, `tail`, `echo`, `tr`, `paste`, `split`, and `wc`,
they're just... different, or entirely missing.

Sure, WSL (Windows Subsystem for Linux) is fantastic, and Git Bash provides a
reasonable shim, but sometimes you just want native, lightweight tools that
*feel* right, without the overhead or compatibility layers. As a personal fun
project, and out of sheer frustration with the impedance mismatch, I decided to
embark on a journey: **to build my own versions of these essential bash
utilities in Go.**

This project, `bashutils`, is my attempt to bring that familiar, powerful
Unix-like command-line experience directly to Windows (and anywhere else Go
compiles!), because, why not? It's a fun way to explore Go, and maybe, just
maybe, make command-line life on Windows a tiny bit more enjoyable for fellow
frustrated developers.

## Features

`bashutils` is a single executable that provides a suite of common command-line
tools. Each command aims to replicate the core functionality of its Unix
counterpart, with a focus on simplicity and portability.

*   **Bash-like Path Globbing**: All file-operating commands support bash-style
    glob patterns (`*`, `?`, `[...]`).
*   **Core Utilities**: Implements common Unix utilities like `cat`, `head`,
    `tail`, `wc`, `grep`, `sort`, `uniq`, `cut`, `paste`, `split`, `tr`, `echo`, and `xargs`.

Currently supported commands:

*   **`cat`**: Concatenate files and print on the standard output.
*   **`cut`**: Remove sections from each line of files.
*   **`echo`**: Display a line of text.
*   **`grep`**: Print lines matching a pattern.
*   **`head`**: Output the first part of files.
*   **`paste`**: Merge lines of files.
*   **`sort`**: Sort lines of text files.
*   **`split`**: Split a file into pieces.
*   **`tail`**: Output the last part of files.
*   **`tr`**: Translate or delete characters.
*   **`uniq`**: Report or omit repeated lines.
*   **`wc`**: Print newline, word, and byte counts for each file.
*   **`xargs`**: Build and execute command lines from standard input.

## Globbing Support

All commands that accept file arguments now support bash-like glob patterns:

*   `*` - matches any sequence of characters
*   `?` - matches any single character
*   `[...]` - matches any character within the brackets
*   `**` - matches any sequence of characters across directory boundaries
    (recursive globbing)

### Globbing Behavior

*   **No matches**: If a glob pattern doesn't match any files, the original
    pattern is preserved (bash behavior).
*   **Multiple matches**: Commands process all matching files.
*   **File validation**: Commands validate that files exist and are readable
    before processing.
*   **Sorted output**: Glob matches are sorted for consistent output.

## Installation

### Prerequisites

*   Go 1.16+ installed ([Download Go](https://golang.org/dl/))
*   `GOPATH` configured and added to your system's `PATH` environment variable.

### Building and Installing

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/monster0506/bashutils-go.git
    cd bashutils-go
    ```

2.  **Download dependencies:**
    ```bash
    go mod tidy
    ```

3.  **Build the executable:**
    We'll explicitly name the executable `bashutils`. This will create
    `bashutils.exe` (on Windows) or `bashutils` (on Linux/macOS) in your current
    directory.
    ```bash
    go build -o bashutils .
    ```

4.  **Install the executable (move to your `PATH`):**
    To make `bashutils` accessible from any directory, move the compiled
    executable to a location already in your system's `PATH`, for example, your
    `GOPATH/bin` directory.

    **On Windows (Command Prompt):**
    ```cmd
    move bashutils.exe "%GOPATH%\bin\bashutils.exe"
    ```
    **On Windows (PowerShell):**
    ```powershell
    Move-Item -Path ".\bashutils.exe" -Destination "$env:GOPATH\bin\bashutils.exe"
    ```
    **On Linux/macOS:**
    ```bash
    mv bashutils "$GOPATH/bin/"
    ```

5.  **Verify installation:**
    Open a **new** terminal window (or restart your current one) to ensure your
    `PATH` is refreshed. Then, try:
    ```bash
    bashutils --help
    ```
    You should see the help message for the `bashutils` command.

## Usage and Commands

Once installed, you can use `bashutils` by specifying the command name as a
subcommand:

```bash
bashutils <command> [arguments...]
```

For detailed usage and flags for each command, use the `--help` flag:

```bash
bashutils <command> --help
```

Below are examples for each supported command, including how globbing can be
used.

### `cat`

Concatenate files and print on the standard output.

```bash
# Display content of a single file
bashutils cat myfile.txt

# Concatenate multiple files using globbing
bashutils cat "test_files/*.txt"

# Process files across subdirectories
bashutils cat "**/*.log"
```

### `cut`

Remove sections from each line of files. Reads from standard input if no file is
provided.

```bash
# Extract the 1st and 3rd comma-separated fields from a file
bashutils cut -d ',' -f 1,3 data.csv

# Extract a character range from multiple CSV files using globbing
bashutils cut -c 1-5,10- data_*.csv
```

### `echo`

Display a line of text.

```bash
bashutils echo "Hello from bashutils!"
bashutils echo "This supports multiple" "arguments."
```

### `grep`

Print lines matching a pattern. Reads from standard input if no file is
provided.

```bash
# Search for "error" (case-insensitive) in a log file
bashutils grep -i "error" logfile.txt

# Search for a pattern in all text files using globbing
bashutils grep "pattern" "test_files/*.txt"

# Search recursively for a pattern in all Python files
bashutils grep "import" "**/*.py"
```

### `head`

Output the first part of files. Reads from standard input if no file is
provided.

```bash
# Display the first 5 lines of a file
bashutils head -n 5 anotherfile.log

# Display the first 10 lines of all markdown files using globbing
bashutils head -n 10 "*.md"
```

### `paste`

Merge lines of files.

```bash
# Merge lines from two files, separated by a tab (default)
bashutils paste names.txt ages.txt

# Merge lines from multiple files using a colon as a delimiter
bashutils paste -d ':' file1.txt file2.txt file3.txt
```

### `sort`

Sort lines of text files. Reads from standard input if no file is provided.

```bash
# Sort a file in ascending order (default)
bashutils sort numbers.txt

# Sort in reverse order
bashutils sort -r mylist.txt

# Sort all CSV files by the second column, numeric sort
bashutils sort -k 2 -n "*.csv"
```

### `split`

Split a file into pieces.

```bash
# Split a large CSV file into chunks of 1000 lines, prefixing output with "chunk_"
bashutils split -l 1000 large_data.csv chunk_

# Split a file into chunks of 1MB (bytes)
bashutils split -b 1M big_file.bin binary_chunk_

# Split all large text files in a directory
bashutils split -l 500 "large_text_files/*.txt" part_
```

### `tail`

Output the last part of files. Reads from standard input if no file is
provided.

```bash
# Display the last 10 lines of a server log
bashutils tail -n 10 server.log

# Display the last 20 lines of all log files in subdirectories
bashutils tail -n 20 "logs/**/*.log"
```

### `tr`

Translate or delete characters. Always reads from standard input and writes to
standard output.

```bash
# Translate lowercase to uppercase
echo "hello world" | bashutils tr 'a-z' 'A-Z'

# Delete all digits from input
echo "My phone is 123-456-7890" | bashutils tr -d '0-9'
```

### `uniq`

Report or omit repeated lines. Often used with `sort`. Reads from standard
input if no file is provided.

```bash
# Find unique lines in a sorted file
bashutils sort mylist.txt | bashutils uniq

# Count occurrences of unique lines
bashutils sort mylist.txt | bashutils uniq -c

# Show only lines that appear exactly once in sorted files
bashutils uniq -u "sorted_data/*.txt"
```

### `wc`

Print newline, word, and byte counts for each file. Reads from standard input
if no file is provided.

```bash
# Count lines in a file
bashutils wc -l somefile.txt

# Count words and characters for multiple files using globbing
bashutils wc -wc "reports/*.txt"

# Get all counts for all files in the current directory
bashutils wc "*"
```

### `xargs`

Build and execute command lines from standard input. Useful for processing lists
of files or arguments from other commands.

```bash
# Count lines in all Python files
git ls-files | bashutils xargs bashutils wc -l

# Find all text files and count their words
find . -name "*.txt" | bashutils xargs bashutils wc -w

# Process files in batches of 10
echo "file1.txt file2.txt file3.txt" | bashutils xargs -n 10 bashutils cat

# Use a custom delimiter (comma-separated values)
echo "file1.txt,file2.txt,file3.txt" | bashutils xargs -d ',' bashutils wc -l

# Replace placeholder in command
echo "file1.txt file2.txt" | bashutils xargs -I {} bashutils echo "Processing: {}"
```

## Contributing

This project is a personal endeavor born out of a desire to learn and fill a
gap. Contributions, bug reports, and feature requests are welcome! Feel free to
open an issue or submit a pull request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE)
file for details.
