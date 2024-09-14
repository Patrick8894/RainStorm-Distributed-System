# cs425G66

## MP1

### Prerequisites

- Install the Go compiler. You can download and install it from the [official Go website](https://golang.org/dl/).

### Setup Instructions

1. **Clone the Repository**:
    ```sh
    git clone <repository-url>
    cd cs425G66
    ```

2. **Sync All Machines with the GitLab Repository**:
    - Run the `pull_repo.sh` script to sync all machines with the GitLab repository:
    ```sh
    ./pull_repo.sh
    ```

3. **Start the Receiver Server**:
    - Navigate to the `mp1` directory and run the `setup.sh` script to start the receiver server:
    ```sh
    cd mp1
    ./setup.sh
    ```

4. **Run the Sender**:
    - Navigate to the `src` directory and run the `sender` with any supported options of the `grep` command to match a pattern. Noted that you don't need to provide filename like in `grep`. You will see the output from all machine:
    ```sh
    cd src
    ./sender <grep-options> <grep-pattern>
    ```

    - Example:
    ```sh
    ./sender -cH "00"
    ```