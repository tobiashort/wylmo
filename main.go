package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const hardTimeoutTest = "Hard timeout"
const inactivityTimeoutTest = "Inactivity timeout"

const colorBlue = "\033[38;5;45m"
const colorRed = "\033[38;5;198m"
const colorReset = "\033[0;0m"

var noColors bool

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func must2[T any](v T, err error) T {
	must(err)
	return v
}

func blue(text string) string {
	if noColors {
		return text
	}
	return colorBlue + text + colorReset
}

func red(text string) string {
	if noColors {
		return text
	}
	return colorRed + text + colorReset
}

func beginColor(color string) {
	if noColors {
		return
	}
	fmt.Print(color)
}

func endColor() {
	if noColors {
		return
	}
	fmt.Print(colorReset)
}

func interpretColorHints(text string) string {
	reds := regexp.MustCompile("#r\\{([^\\}]*)\\}")
	blues := regexp.MustCompile("#b\\{([^\\}]*)\\}")
	if noColors {
		text = reds.ReplaceAllString(text, "${1}")
		text = blues.ReplaceAllString(text, "${1}")
	} else {
		text = reds.ReplaceAllString(text, red("${1}"))
		text = blues.ReplaceAllString(text, blue("${1}"))
	}
	return text
}

func println(text string) {
	fmt.Println(interpretColorHints(text))
}

func printf(text string, a ...any) {
	fmt.Printf(interpretColorHints(text), a...)
}

func readLine() string {
	reader := bufio.NewReader(os.Stdin)
	bytesRead := must2(reader.ReadString('\n'))
	return strings.TrimSpace(string(bytesRead))
}

func readMultiLine() string {
	bytesRead := must2(io.ReadAll(os.Stdin))
	return strings.TrimSpace(string(bytesRead))
}

func choose(text string, choices []string) string {
	println(text)
	for index, choice := range choices {
		printf("#b{[%d] %s}\n", index, choice)
	}
	printf("Please enter your choice (0-%d): ", len(choices)-1)
	beginColor(colorBlue)
	answer := readLine()
	endColor()
	choosen, err := strconv.Atoi(answer)
	if err != nil {
		println("#r{Invalid input. Please try again.}")
		return choose(text, choices)
	}
	if choosen < 0 || choosen >= len(choices) {
		println("#r{Invalid input. Please try again.}")
		return choose(text, choices)
	}
	return choices[choosen]
}

func yesno(question string) bool {
	printf(question + " (y/n) ")
	beginColor(colorBlue)
	answer := readLine()
	endColor()
	answer = strings.ToLower(answer)
	switch answer {
	case "y":
		return true
	case "n":
		return false
	}
	return yesno(question)
}

func requestCurlCommand() string {
	println("Please enter the curl command and accept with Ctrl-D.")
	beginColor(colorBlue)
	curlCommand := readMultiLine()
	endColor()
	if !strings.HasPrefix(curlCommand, "curl ") {
		printf("#r{Not a curl command: %v\n}", curlCommand)
		return requestCurlCommand()
	}
	println("Testing curl command...")
	cmd := exec.Command("bash", "-c", curlCommand)
	outBytes, err := cmd.CombinedOutput()
	output := string(outBytes)
	output = strings.TrimSpace(output)
	if err != nil {
		println("#r{Curl command failed}")
		println(red(err.Error()))
		if output != "" {
			println(red(output))
		}
		return requestCurlCommand()
	}
	println("Curl command was successful.")
	printf("Please hit enter to review the curl command's output before continuing")
	readLine()
	cmd = exec.Command("more")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	writer := must2(cmd.StdinPipe())
	must(cmd.Start())
	must2(writer.Write(outBytes))
	must(writer.Close())
	must(cmd.Wait())
	if ok := yesno("Is the curl command's output ok?"); ok {
		return curlCommand
	}
	return requestCurlCommand()
}

func performHardTimeoutTest(curlCommand string) {
	printf("Performing #b{'%s'} test...\n", hardTimeoutTest)
	must(os.Mkdir("hard_timeout", 0755))
	must(os.WriteFile("hard_timeout/curl_command.txt", []byte(curlCommand), 0644))
	interval := 5 * time.Minute
	printf("Interval is set to #b{'%v'}\n", interval)
	for {
		cmd := exec.Command("bash", "-c", curlCommand)
		bytesOut, err := cmd.CombinedOutput()
		output := string(bytesOut)
		if err != nil {
			output = err.Error() + "\n" + output
		}
		now := time.Now()
		logFile := strings.ReplaceAll(fmt.Sprintf("hard_timeout/%v.txt", now), " ", "_")
		must(os.WriteFile(logFile, []byte(output), 0644))
		firstLine := must2(bufio.NewReader(strings.NewReader(output)).ReadString('\n'))
		if err != nil {
			printf("%v #r{%s}", now, firstLine)
		} else {
			printf("%v #b{%s}", now, firstLine)
		}
		time.Sleep(interval)
	}
}

func performInactivityTimeoutTest(curlCommand string) {
	printf("Performing #b{'%s'} test...\n", inactivityTimeoutTest)
	must(os.Mkdir("inactivity_timeout", 0755))
	must(os.WriteFile("inactivity_timeout/curl_command.txt", []byte(curlCommand), 0644))
	interval := 0 * time.Minute
	for {
		printf("Waiting for #b{'%v'}\n", interval)
		time.Sleep(interval)
		cmd := exec.Command("bash", "-c", curlCommand)
		bytesOut, err := cmd.CombinedOutput()
		output := string(bytesOut)
		if err != nil {
			output = err.Error() + "\n" + output
		}
		now := time.Now()
		logFile := strings.ReplaceAll(fmt.Sprintf("inactivity_timeout/%v.txt", now), " ", "_")
		must(os.WriteFile(logFile, []byte(output), 0644))
		firstLine := must2(bufio.NewReader(strings.NewReader(output)).ReadString('\n'))
		if err != nil {
			printf("%v #r{%s}", now, firstLine)
		} else {
			printf("%v #b{%s}", now, firstLine)
		}
		interval += 15 * time.Minute
	}
}

func performTest(typeOfTest string, curlCommand string) {
	switch typeOfTest {
	case hardTimeoutTest:
		performHardTimeoutTest(curlCommand)
	case inactivityTimeoutTest:
		performInactivityTimeoutTest(curlCommand)
	default:
		panic("Unknown test to perform: " + typeOfTest)
	}
}

func main() {
	println("Welcome to #b{wylmo}!")
	flag.BoolVar(&noColors, "nocolors", false, "Disable colored output")
	flag.Parse()
	typeOfTest := choose("Please choose the type of test to perform", []string{
		hardTimeoutTest,
		inactivityTimeoutTest,
	})
	printf("Thank you for choosing #b{'%s'}\n", typeOfTest)
	curlCommand := requestCurlCommand()
	performTest(typeOfTest, curlCommand)
}
