# Trash CLI

**Trash CLI** is a command-line utility for managing files and directories in the trash, compliant with the  [XDg Trash specification 1.0](https://specifications.freedesktop.org/trash-spec/1.0/).

## Features

- **Trashing**: Moves files and directories to the trash.
- **Restoring**: Restores files and directories from the trash to their original location.
- **Listing**: Displays the files currently in the trash.
- **Purging**: Completely empties the trash.

## Requirements

- **Go**: Version 1.24.2 or higher.
- Unix/Linux-based operating system.

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/The-EpaG/trash-cli.git
   cd trash-cli
   ```
2. Build the project:
    ```shell
    go build -o trash-cli
    ```
3. (Optional) Install the binary:
    ```shell
    sudo mv trash-cli /usr/local/bin/
    ```

## Usage
Main Commands

- **Trash files or directories:**
    ```shell
    trash-cli trash [file...]
    ```
    Aliases: `rm`, `remove`, `delete`, `del`.

- **Restore a file:**
    ```shell
    trash-cli restore [file]
    ```

- **List files in the trash:**
    ```shell
    trash-cli list
    ```
    Option: `-l` o `--details` to show additional details.

- **Empty the trash:**
    ```shell
    trash-cli purge
    ```

### Examples
- **Move a file to the trash:**
    ```shell
    trash-cli trash example.txt
    ```
- **Restore a file from the trash:**
    ```shell
    trash-cli restore example.txt
    ```
- **List files in the trash with details:**
    ```shell
    trash-cli list --details
    ```
- **Empty the trash:**
    ```shell
    trash-cli purge
    ```

## Project Structure
- `main.go`: Main entry point of the application.
- `cmd`: Contains the main commands (`trash`, `restore`, `list`, `purge`).
- `internal`: Contains the internal logic for managing the trash.
- `directive.rst`: XDG Trash specification for reference.

## Compliance with the XDG Trash Specification
This project implements the XDG Trash specification v1.0. For more details, see the `directive.rst` file.

## Contributing
1. Fork the repository.
2. Create a branch for your feature or fix:
3. Commit your changes:
4. Push the branch:
5. Open a pull request.

## License
See the `LICENSE` file for licensing details.

## Authors
- **Mikhail Ramendik**, **David Faure**, **Alexander Larsson**, **Ryan Lortie** - XDG Trash Specification.
- **The-EpaG** - CLI Implementation.

## Contact
- For questions or bug reports, open an issue on [GitHub](https://github.com/The-EpaG/trash-cli/issues).

