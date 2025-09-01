package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/tobiashort/ansi-go"
	"github.com/tobiashort/cfmt-go"
	"github.com/tobiashort/choose-go"
	"github.com/tobiashort/clap-go"

	. "github.com/tobiashort/cosine-similarity-go"
	. "github.com/tobiashort/utils-go/must"
)

const (
	hardTimeoutTest       = "Hard timeout"
	inactivityTimeoutTest = "Inactivity timeout"
)

type Args struct {
	Interval time.Duration `clap:"description='Timeout interval in minutes (default: 5min for hard timeout, 15min for inactivity timeout).'"`
}

var (
	intervalArg       time.Duration
	referenceResponse string
	startTime         = time.Now()
)

func readLine() string {
	reader := bufio.NewReader(os.Stdin)
	bytesRead := Must2(reader.ReadString('\n'))
	return strings.TrimSpace(string(bytesRead))
}

func readMultiLine() string {
	bytesRead := Must2(io.ReadAll(os.Stdin))
	return strings.TrimSpace(string(bytesRead))
}

func formatTime(t time.Time) string {
	elapsed := t.Sub(startTime)
	return fmt.Sprintf("%s +%s", t.Format("2006-01-02 15-04-05"), elapsed)
}

func requestCurlCommand() string {
	fmt.Println("Please enter the curl command and accept with Ctrl-D.")
	cfmt.Begin(ansi.DecorPurple)
	curlCommand := readMultiLine()
	cfmt.End()
	if !strings.HasPrefix(curlCommand, "curl ") {
		cfmt.Printf("#r{Not a curl command: %v\n}", curlCommand)
		return requestCurlCommand()
	}
	fmt.Println("Testing curl command...")
	cmd := exec.Command("bash", "-c", curlCommand)
	outBytes, err := cmd.CombinedOutput()
	output := string(outBytes)
	output = strings.TrimSpace(output)
	if err != nil {
		cfmt.Println("#r{Curl command failed}")
		cfmt.CPrintln(ansi.DecorRed, err.Error())
		if output != "" {
			cfmt.CPrintln(ansi.DecorRed, (output))
		}
		return requestCurlCommand()
	}
	fmt.Println("Curl command was successful.")
	fmt.Printf("Please hit enter to review the curl command's output before continuing.")
	readLine()
	cmd = exec.Command("more")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	writer := Must2(cmd.StdinPipe())
	Must(cmd.Start())
	Must2(writer.Write(outBytes))
	Must(writer.Close())
	Must(cmd.Wait())
	if choose.YesNo("Is the curl command's output ok?", choose.DEFAULT_NONE) {
		referenceResponse = string(outBytes)
		return curlCommand
	}
	return requestCurlCommand()
}

func performHardTimeoutTest(curlCommand string) {
	cfmt.Printf("Performing #yB{'%s'} test...\n", hardTimeoutTest)
	if _, err := os.Stat("hard_timeout"); err == nil {
		if choose.YesNo("Remove previous test results?", choose.DEFAULT_NONE) {
			Must(os.RemoveAll("hard_timeout"))
		} else {
			fmt.Printf("Abort.\n")
			os.Exit(1)
		}
	}
	Must(os.Mkdir("hard_timeout", 0755))
	Must(os.WriteFile("hard_timeout/curl_command", []byte(curlCommand), 0644))
	logFile := Must2(os.OpenFile("hard_timeout/log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644))
	defer logFile.Close()
	interval := intervalArg
	if interval == 0 {
		interval = 5 * time.Minute
	}
	cfmt.Printf("Interval is set to #yB{'%v'}\n", interval)
	startTime = time.Now()
	for {
		cmd := exec.Command("bash", "-c", curlCommand)
		bytesOut, err := cmd.CombinedOutput()
		output := string(bytesOut)
		if err != nil {
			output = err.Error() + "\n" + output
		}
		now := time.Now()
		if err != nil {
			curlLogFile := fmt.Sprintf("hard_timeout/%v", formatTime(now))
			Must(os.WriteFile(curlLogFile, []byte(output), 0644))
			cfmt.Printf("%v #r{%s}\n", formatTime(now), output)
			Must2(fmt.Fprintf(logFile, "%v %s\n", formatTime(now), output))
		} else {
			similarity := CosineSimilarity(referenceResponse, output)
			curlLogFile := fmt.Sprintf("hard_timeout/%v %f similarity", formatTime(now), similarity)
			Must(os.WriteFile(curlLogFile, []byte(output), 0644))
			cfmt.Printf("%v #yB{%f} similarity\n", formatTime(now), similarity)
			Must2(fmt.Fprintf(logFile, "%v %f similarity\n", formatTime(now), similarity))
		}
		time.Sleep(interval)
	}
}

func performInactivityTimeoutTest(curlCommand string) {
	cfmt.Printf("Performing #yB{'%s'} test...\n", inactivityTimeoutTest)
	if _, err := os.Stat("inactivity_timeout"); err == nil {
		if choose.YesNo("Remove previous test results?", choose.DEFAULT_NONE) {
			Must(os.RemoveAll("inactivity_timeout"))
		} else {
			fmt.Printf("Abort.\n")
			os.Exit(1)
		}
	}
	Must(os.Mkdir("inactivity_timeout", 0755))
	Must(os.WriteFile("inactivity_timeout/curl_command", []byte(curlCommand), 0644))
	logFile := Must2(os.OpenFile("inactivity_timeout/log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644))
	defer logFile.Close()
	interval := 0 * time.Minute
	startTime = time.Now()
	for {
		cfmt.Printf("Waiting for #yB{'%v'}\n", interval)
		time.Sleep(interval)
		cmd := exec.Command("bash", "-c", curlCommand)
		bytesOut, err := cmd.CombinedOutput()
		output := string(bytesOut)
		if err != nil {
			output = err.Error() + "\n" + output
		}
		now := time.Now()
		if err != nil {
			curlLogFile := fmt.Sprintf("inactivity_timeout/%v", formatTime(now))
			Must(os.WriteFile(curlLogFile, []byte(output), 0644))
			cfmt.Printf("%v #r{%s}\n", formatTime(now), output)
			Must2(fmt.Fprintf(logFile, "%v %s\n", formatTime(now), output))
		} else {
			similarity := CosineSimilarity(referenceResponse, output)
			curlLogFile := fmt.Sprintf("inactivity_timeout/%v %f similarity", formatTime(now), similarity)
			Must(os.WriteFile(curlLogFile, []byte(output), 0644))
			cfmt.Printf("%v #yB{%f} similarity\n", formatTime(now), similarity)
			Must2(fmt.Fprintf(logFile, "%v %f similarity\n", formatTime(now), similarity))
		}
		if intervalArg == 0 {
			interval += 15 * time.Minute
		} else {
			interval += intervalArg
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
		fmt.Print(ansi.DecorReset)
		fmt.Printf("\nAbort.\n")
		os.Exit(1)
	}()

	cfmt.Println("Welcome to #yB{wylmo}!")

	args := Args{}
	clap.Parse(&args)
	intervalArg = args.Interval

	typeOfTest, ok := choose.One("Please choose the type of test to perform", []string{
		hardTimeoutTest,
		inactivityTimeoutTest,
	})
	if ok {
		cfmt.Printf("Thank you for choosing #yB{'%s'}\n", typeOfTest)
		curlCommand := requestCurlCommand()
		performTest(typeOfTest, curlCommand)
	} else {
		fmt.Println("Abort.")
	}
}
