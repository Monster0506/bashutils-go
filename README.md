# bashutils

[![Go Report Card](https://goreportcard.com/badge/github.com/monster0506/bashutils-go)](https://goreportcard.com/report/monster0506/bashutils-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go-based reimplementation of bash coreutils with bash-like path globbing support.

## Bringing Unix Comfort to Windows

If you've spent any significant time in a Linux or macOS terminal, then tried to do something meaningful in the Windows Command Prompt or PowerShell, you know the feeling. The common, indispensable utilities like `grep`, `cut`,
`sort`, `uniq`, `cat`, `head`, `tail`, `echo`, `tr`, `paste`, `split`, and `wc` are just different or entirely missing.

Sure, WSL (Windows Subsystem for Linux) is fantastic, and Git Bash provides a
reasonable shim, but sometimes you just want native, lightweight tools that
*feel* right, without the overhead or compatibility layers. As a personal fun
project, and out of frustration, I decided to to build my own versions of these bash utilities in Go.

This project, `bashutils`, is my attempt to bring that familiar, powerful
Unix-like command-line experience directly to Windows (and anywhere else Go
compiles!), because, why not? It's a fun way to explore Go, and maybe make command-line life on Windows a tiny bit more enjoyable for fellow developers.

## Features

`bashutils` is a single executable that provides a suite of common command-line
tools. Each command aims to replicate the core functionality of its Unix
counterpart, with a focus on simplicity and portability.

*   **Bash-like Path Globbing**: All file-operating commands support bash-style
    glob patterns (`*`, `?`, `[...]`).
*   **Core Utilities**: Implements common Unix utilities like `cat`, `head`,
    `tail`, `wc`, `grep`, `sort`, `uniq`, `cut`, `paste`, `split`, `tr`, `echo`, and `xargs`.

Currently supported commands:

*   **`cal`**: Display a calendar. `-m` Monday first, `-y` whole year, `-3` three months.
*   **`cat`**: Concatenate files. `-n -b -s -E -T`.
*   **`cut`**: Remove sections from each line. `-f -d -c -b -s --complement --output-delimiter`.
*   **`echo`**: Display a line of text. `-n -e -E`, `--expand-env`.
*   **`grep`**: Print lines matching a pattern. `-e -f -E -F -i -v -w -x -n -c -l -L -q -s -m -H -h -o -A -B -C -r -R --include --exclude --exclude-dir`.
*   **`head`**: Output the first part of files. `-n NUM` (or `-NUM`, negative NUM), `-c`, `-q`, `-v`.
*   **`paste`**: Merge lines of files. `-d -s`.
*   **`sort`**: Sort lines. `-k KEYDEF -t -n -r -u -f -b -s -c -o`.
*   **`split`**: Split a file into pieces. `-l -b -d -a --additional-suffix`.
*   **`tail`**: Output the last part of files. `-n NUM` (or `-NUM`, `+NUM`), `-c`, `-q`, `-v`.
*   **`tr`**: Translate or delete characters. `-d -c -s -t`, classes like `[:alpha:]`, escapes like `\n`.
*   **`uniq`**: Report or omit repeated lines. `-c -d -D -u -i -f -s -w`.
*   **`wc`**: Print line, word, character, byte and line-length counts. `-l -w -m -c -L`.
*   **`xargs`**: Build and execute command lines from standard input. `-n -L -I -0 -d -a -E -P -r -t`.

`-` reads standard input wherever a file is accepted. Read errors set a non-zero
exit status without stopping the remaining inputs. Lines keep their CRLF endings
and a missing final newline is preserved by `cat`, `head`, `tail` and `split`.
`xargs` runs the commands above in-process before searching `PATH`.

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

*   Go 1.22+ installed ([Download Go](https://golang.org/dl/))
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

### Running commands without the `bashutils` prefix

The executable picks its command from its own file name (`grep.exe` runs
`bashutils grep`). Create one executable per command in a directory and add
that directory to your `PATH`:

```bash
bashutils install "$HOME/bashutils-bin"
```

Hardlinks are used (a copy if linking fails). Rerun with `--force` after
rebuilding to refresh them. On Windows, PowerShell aliases (`cat`, `echo`,
`sort`) take precedence over executables on `PATH`; remove them with
`Remove-Item Alias:cat,Alias:echo,Alias:sort` in your `$PROFILE`.

## Contributing

This project is a personal endeavor born out of a desire to learn and fill a
gap. Contributions, bug reports, and feature requests are welcome! Feel free to
open an issue or submit a pull request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE)
file for details.
