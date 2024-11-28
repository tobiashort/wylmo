package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const hardTimeoutTest = "Hard timeout"
const inactivityTimeoutTest = "Inactivity timeout"

const colorMagenta = "\033[1;35m"
const colorRed = "\033[1;31m"
const colorReset = "\033[0;0m"

var noColorsFlag bool
var intervalFlag time.Duration

var referenceResponse string
var referenceResponseVec map[string]float64

var startTime = time.Now()

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
	if noColorsFlag {
		return text
	}
	return colorMagenta + text + colorReset
}

func red(text string) string {
	if noColorsFlag {
		return text
	}
	return colorRed + text + colorReset
}

func beginColor(color string) {
	if noColorsFlag {
		return
	}
	fmt.Print(color)
}

func endColor() {
	if noColorsFlag {
		return
	}
	fmt.Print(colorReset)
}

func interpretColorHints(text string) string {
	reds := regexp.MustCompile("#r\\{([^\\}]*)\\}")
	blues := regexp.MustCompile("#m\\{([^\\}]*)\\}")
	if noColorsFlag {
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
		printf("#m{[%d] %s}\n", index, choice)
	}
	printf("Please enter your choice (0-%d): ", len(choices)-1)
	beginColor(colorMagenta)
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
	beginColor(colorMagenta)
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

func formatTime(t time.Time) string {
	elapsed := t.Sub(startTime)
	return fmt.Sprintf("%s +%s", t.Format("2006-01-02 15-04-05"), elapsed)
}

func requestCurlCommand() string {
	println("Please enter the curl command and accept with Ctrl-D.")
	beginColor(colorMagenta)
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
	printf("Please hit enter to review the curl command's output before continuing.")
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
		referenceResponse = string(outBytes)
		referenceResponseVec = text2vec(referenceResponse)
		return curlCommand
	}
	return requestCurlCommand()
}

func text2vec(text string) map[string]float64 {
	vec := make(map[string]float64)
	for _, field := range strings.Fields(text) {
		count := vec[field]
		count++
		vec[field] = count
	}
	return vec
}

func cosineSimilarity(vec1, vec2 map[string]float64) float64 {
	biggerVec := vec1
	if len(vec2) > len(vec1) {
		biggerVec = vec2
	}
	var divident float64
	for key := range biggerVec {
		divident += vec1[key] * vec2[key]
	}
	var divisorPart1 float64
	for _, value := range vec1 {
		divisorPart1 += value * value
	}
	divisorPart1 = math.Sqrt(divisorPart1)
	var divisorPart2 float64
	for _, value := range vec2 {
		divisorPart2 += value * value
	}
	divisorPart2 = math.Sqrt(divisorPart2)
	divisor := divisorPart1 * divisorPart2
	return divident / divisor
}

func performHardTimeoutTest(curlCommand string) {
	printf("Performing #m{'%s'} test...\n", hardTimeoutTest)
	if _, err := os.Stat("hard_timeout"); err == nil {
		if yesno("Remove previous test results?") {
			must(os.RemoveAll("hard_timeout"))
		} else {
			printf("Abort.\n")
			os.Exit(1)
		}
	}
	must(os.Mkdir("hard_timeout", 0755))
	must(os.WriteFile("hard_timeout/curl_command", []byte(curlCommand), 0644))
	logFile := must2(os.OpenFile("hard_timeout/log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644))
	defer logFile.Close()
	interval := intervalFlag
	if interval == 0 {
		interval = 5 * time.Minute
	}
	printf("Interval is set to #m{'%v'}\n", interval)
	startTime = time.Now()
	for {
		cmd := exec.Command("bash", "-c", curlCommand)
		bytesOut, err := cmd.CombinedOutput()
		output := string(bytesOut)
		if err != nil {
			output = err.Error() + "\n" + output
		}
		now := time.Now()
		curlLogFile := fmt.Sprintf("hard_timeout/%v", formatTime(now))
		must(os.WriteFile(curlLogFile, []byte(output), 0644))
		if err != nil {
			printf("%v #r{%s}\n", formatTime(now), output)
			must2(logFile.WriteString(fmt.Sprintf("%v %s\n", formatTime(now), output)))
		} else {
			similarity := cosineSimilarity(referenceResponseVec, text2vec(output))
			printf("%v #m{%f} similarity\n", formatTime(now), similarity)
			must2(logFile.WriteString(fmt.Sprintf("%v %f similarity\n", formatTime(now), similarity)))
		}
		time.Sleep(interval)
	}
}

func performInactivityTimeoutTest(curlCommand string) {
	printf("Performing #m{'%s'} test...\n", inactivityTimeoutTest)
	if _, err := os.Stat("inactivity_timeout"); err == nil {
		if yesno("Remove previous test results?") {
			must(os.RemoveAll("inactivity_timeout"))
		} else {
			printf("Abort.\n")
			os.Exit(1)
		}
	}
	must(os.Mkdir("inactivity_timeout", 0755))
	must(os.WriteFile("inactivity_timeout/curl_command", []byte(curlCommand), 0644))
	logFile := must2(os.OpenFile("inactivity_timeout/log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644))
	defer logFile.Close()
	interval := 0 * time.Minute
	startTime = time.Now()
	for {
		printf("Waiting for #m{'%v'}\n", interval)
		time.Sleep(interval)
		cmd := exec.Command("bash", "-c", curlCommand)
		bytesOut, err := cmd.CombinedOutput()
		output := string(bytesOut)
		if err != nil {
			output = err.Error() + "\n" + output
		}
		now := time.Now()
		curlLogFile := fmt.Sprintf("inactivity_timeout/%v", formatTime(now))
		must(os.WriteFile(curlLogFile, []byte(output), 0644))
		if err != nil {
			printf("%v #r{%s}\n", formatTime(now), output)
			must2(logFile.WriteString(fmt.Sprintf("%v %s\n", formatTime(now), output)))
		} else {
			similarity := cosineSimilarity(referenceResponseVec, text2vec(output))
			printf("%v #m{%f} similarity\n", formatTime(now), similarity)
			must2(logFile.WriteString(fmt.Sprintf("%v %f similarity\n", formatTime(now), similarity)))
		}
		if intervalFlag == 0 {
			interval += 15 * time.Minute
		} else {
			interval += intervalFlag
		}
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
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		printf(colorReset)
		printf("\nAbort.\n")
		os.Exit(1)
	}()
	println("Welcome to #m{wylmo}!")
	flag.BoolVar(&noColorsFlag, "nocolors", false, "Disable colored output")
	flag.DurationVar(&intervalFlag, "interval", 0, "Timeout interval in minutes (default: 5min for hard timeout, 15min for inactivity timeout).")
	flag.Parse()
	typeOfTest := choose("Please choose the type of test to perform", []string{
		hardTimeoutTest,
		inactivityTimeoutTest,
	})
	printf("Thank you for choosing #m{'%s'}\n", typeOfTest)
	curlCommand := requestCurlCommand()
	performTest(typeOfTest, curlCommand)
}
