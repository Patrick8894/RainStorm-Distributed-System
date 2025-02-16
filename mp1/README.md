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
    cd src/client
    ./sender <number-of-machine-to-connect> <grep-options> <grep-pattern>
    ```

    - Example:
    ```sh
    ./sender 4 -cH "00"
    ```


###  Testing
1. Generate test log file
10000 is the argument for number of random lines
    ```
    python3 generator.py 10000
    ```
2. Navigate to sender directory
    ```sh
    cd mp1/src/client
    ```
3. run go test and watch detail
    ```sh
    go test -v
    ```

