package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/t-hg/wylmo/assert"
)

const colorBlue = "\033[38;5;45m"
const colorRed = "\033[38;5;198m"
const colorReset = "\033[0;0m"

func blue(text string) string {
	return colorBlue + text + colorReset
}

func red(text string) string {
	return colorRed + text + colorReset
}

func readLine() string {
	reader := bufio.NewReader(os.Stdin)
	bytesRead, err := reader.ReadString('\n')
	assert.NoErr(err)
	return strings.TrimSpace(string(bytesRead))
}

func readMultiLine() string {
	bytesRead, err := io.ReadAll(os.Stdin)
	assert.NoErr(err)
	return strings.TrimSpace(string(bytesRead))
}

func choose(text string, choices []string) string {
	fmt.Println(text)
	for index, choice := range choices {
		fmt.Printf(blue("[%d] %s\n"), index, choice)
	}
	fmt.Printf("Please enter your choice (0-%d): ", len(choices)-1)
	fmt.Printf(colorBlue)
	answer := readLine()
	fmt.Printf(colorReset)
	choosen, err := strconv.Atoi(answer)
	if err != nil {
		fmt.Println(red("Invalid input. Please try again."))
		return choose(text, choices)
	}
	if choosen < 0 || choosen >= len(choices) {
		fmt.Println(red("Invalid input. Please try again."))
		return choose(text, choices)
	}
	return choices[choosen]
}

func yesno(question string) bool {
	fmt.Print(question + " (y/n) ")
	fmt.Print(colorBlue)
	answer := readLine()
	answer = strings.ToLower(answer)
	fmt.Print(colorReset)
	switch answer {
	case "y":
		return true
	case "n":
		return false
	}
	return yesno(question)
}

func requestCurlCommand() string {
	fmt.Println("Please enter the curl command and accept with Ctrl-D.")
	fmt.Print(colorBlue)
	curlCommand := readMultiLine()
	fmt.Print(colorReset)
	if !strings.HasPrefix(curlCommand, "curl ") {
		fmt.Printf(red("Not a curl command: %v\n"), curlCommand)
		return requestCurlCommand()
	}
	fmt.Println("Testing curl command...")
	cmd := exec.Command("bash", "-c", curlCommand)
	outBytes, err := cmd.CombinedOutput()
	output := string(outBytes)
	output = strings.TrimSpace(output)
	if err != nil {
		fmt.Println(red("Curl command failed"))
		fmt.Println(red(err.Error()))
		if output != "" {
			fmt.Println(red(output))
		}
		return requestCurlCommand()
	}
	fmt.Println("Curl command was successful.")
	fmt.Print("Please hit enter to review the curl command's output before continuing")
	readLine()
	assert.NoErr(err)
	cmd = exec.Command("more")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	writer, err := cmd.StdinPipe()
	assert.NoErr(err)
	err = cmd.Start()
	assert.NoErr(err)
	writer.Write(outBytes)
	writer.Close()
	err = cmd.Wait()
	assert.NoErr(err)
	ok := yesno("Is the curl command's output ok?")
	if ok {
		return curlCommand
	}
	return requestCurlCommand()
}

func performTest(typeOfTest string, curlCommand string) {
	assert.True(typeOfTest == "Hard timeout" || typeOfTest == "Inactivity timeout", "unknown type of test: "+typeOfTest)
	fmt.Println("Performing '" + blue(typeOfTest) + "' test...")
	start := time.Now()
	if typeOfTest == "Hard timeout" {
		err := os.Mkdir("hard_timeout", 0755)
		if err != nil && !os.IsExist(err) {
			panic(err)
		}
		for {
			now := time.Now()
			fmt.Printf("It is now "+blue("'%v'")+"\n", now)
			cmd := exec.Command("bash", "-c", curlCommand)
			bytesOut, _ := cmd.CombinedOutput()
			logFile := fmt.Sprintf("hard_timeout/%v.log", now)
			os.WriteFile(logFile, bytesOut, 0644)
			response, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(bytesOut)), nil)
			if err != nil || response.StatusCode > 299 || response.StatusCode < 200 {
				fmt.Println(red(string(bytesOut)))
				fmt.Printf(red("Time elapsed: %v\n"), time.Now().Sub(start))
				break
			}
			time.Sleep(5 * time.Minute)
		}
	} else {
		err := os.Mkdir("inactivity_timeout", 0755)
		if err != nil && !os.IsExist(err) {
			panic(err)
		}
		var duration time.Duration
		for {
			fmt.Printf("Waiting for "+blue("'%v'")+"\n", duration)
			time.Sleep(duration)
			now := time.Now()
			fmt.Printf("It is now "+blue("'%v'")+"\n", now)
			cmd := exec.Command("bash", "-c", curlCommand)
			bytesOut, _ := cmd.CombinedOutput()
			logFile := fmt.Sprintf("inactivity_timeout/%v.log", now)
			os.WriteFile(logFile, bytesOut, 0644)
			response, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(bytesOut)), nil)
			if err != nil || response.StatusCode > 299 || response.StatusCode < 200 {
				fmt.Println(red(string(bytesOut)))
				fmt.Printf(red("Time elapsed: %v\n"), time.Now().Sub(start))
				break
			}
			duration += 15 * time.Minute
		}
	}
}

func main() {
	fmt.Println("Welcome to wylmo!")
	typeOfTest := choose("Please choose the type of test to perform", []string{
		"Hard timeout",
		"Inactivity timeout",
	})
	fmt.Printf("Thank you for choosing "+blue("'%s'")+"\n", typeOfTest)
	curlCommand := requestCurlCommand()
	performTest(typeOfTest, curlCommand)
}
