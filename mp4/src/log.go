package main

import (
    "fmt"
    "os"
    "os/exec"
	"bufio"
	"strings"
	"strconv"
	"hash/crc32"
	"math/rand"
	"time"
)

func main() {
    if len(os.Args) != 8 {
        fmt.Println("Usage: <op1_exe> <op2_exe> <hydfs_src_file> <hydfs_dest_filename> <num_tasks> <X> <stateful>")
        return
    }

    op1Exe := os.Args[1]
    op2Exe := os.Args[2]
    hydfsSrcFile := os.Args[3]
    hydfsDestFilename := os.Args[4]
    numTasks := os.Args[5]
    X := os.Args[6]
    stateful := os.Args[7]

	totalNum, err := strconv.Atoi(numTasks)

	emptyFilename := fmt.Sprintf("%s/empty", os.Getenv("HOME"))

	localfilename := fmt.Sprintf("%s/localfile", os.Getenv("HOME"))

	emptyFile, _ := os.Create(emptyFilename)

	cmd := exec.Command("go", "run", "mp3_client.go", "create", "--localfilename", emptyFilename, "--HyDFSfilename", hydfsDestFilename)

	cmd = exec.Command("go", "run", "mp3_client.go", "get", "--localfilename", localfilename, "--HyDFSfilename", hydfsSrcFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error executing get: %v\n", err)
		fmt.Printf("Command output: %s\n", string(output))
		return
	}

	// read the local file
	file, err := os.Open(localfilename)
	if err != nil {
		fmt.Printf("Error opening local file: %v\n", err)
		return
	}
	// read the file line by line
	scanner := bufio.NewScanner(file)

	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	totalLines := len(lines)
    partitionSize := totalLines / totalNum
    if err != nil {
        fmt.Printf("Error converting taskNo to integer: %v\n", err)
        return
    }

    stage2 := make(map[int][]string)
	stage3 := make(map[int][]string)

    // Partition lines into totalNum parts
    for i := 0; i < totalNum; i++ {
        start := i * partitionSize
        end := start + partitionSize
        if i == totalNum-1 {
            end = totalLines // Ensure the last partition includes any remaining lines
        }
        stage2[i] = lines[start:end]
    }
	
	for i := 0; i < totalNum; i++ {
		stage2Filename := fmt.Sprintf("/tmp/2_%d_PROC", i)
		file, err := os.Create(stage2Filename)

		stage2ACKFilename := fmt.Sprintf("/tmp/2_%d_ACKED", i)
		ackFile, err := os.Create(stage2ACKFilename)

		for _, line := range lines {
			cmd = exec.Command("../ops/" + op1Exe, X)
			cmd.Stdin = strings.NewReader(line)
			output, err = cmd.Output()
			if err != nil {
				fmt.Printf("Error running external program: %v\n", err)
				continue
			}

			if strings.TrimSpace(string(output)) != "1" {
				continue
			}

			cmd = exec.Command("../ops/" + op2Exe)
			cmd.Stdin = strings.NewReader(line)
			output, err = cmd.Output()
			if err != nil {
				fmt.Printf("Error running external program: %v\n", err)
				continue
			}
			outputStr := string(output)

			file.WriteString(fmt.Sprintf("%s@%s^%s\n", line, line, outputStr))
			ackFile.WriteString(fmt.Sprintf("%s^%s\n", line, outputStr))

			// compute hash value
			h := crc32.ChecksumIEEE([]byte(outputStr))
			hash := int(h % uint32(totalNum))
			stage3[hash] = append(stage3[hash], outputStr)
		}

		exec.Command("go", "run", "mp3_client.go", "create", "--localfilename", stage2Filename, "--HyDFSfilename", fmt.Sprintf("2_%d_PROC", i))
		exec.Command("go", "run", "mp3_client.go", "create", "--localfilename", stage2ACKFilename, "--HyDFSfilename", fmt.Sprintf("2_%d_ACK", i))
	}

	rand.Seed(time.Now().UnixNano())
    for i := 0; i < totalNum; i++ {
        rand.Shuffle(len(stage3[i]), func(j, k int) {
            stage3[i][j], stage3[i][k] = stage3[i][k], stage3[i][j]
        })
    }

	for i := 0; i < totalNum; i++ {
		state := make(map[string]int)

		file, _ := os.Open(fmt.Sprintf("/tmp/3_%d_PROC", i))	
		outputfile, _ := os.Create(fmt.Sprintf("/tmp/3_%d_OUT", i))
		stateFile, _ := os.Open(fmt.Sprintf("/tmp/3_%d_STATE", i))
		for _, line := range stage3[i] {
			if stateful == "stateful" {
				state[line]++
			} else {
				outputfile.WriteString(fmt.Sprintf("%s\n", line))
				emptyFile.WriteString(fmt.Sprintf("%s\n", line))
			}
			file.WriteString(fmt.Sprintf("%s\n", line))
		}
		if stateful == "stateful" {
			for k, v := range state {
				stateFile.WriteString(fmt.Sprintf("%s %d\n", k, v))
				emptyFile.WriteString(fmt.Sprintf("%s %d\n", k, v))
			}
		} 
	}
	cmd = exec.Command("go", "run", "mp3_client.go", "create", "--localfilename", emptyFilename, "--HyDFSfilename", hydfsDestFilename)
}