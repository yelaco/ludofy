package stofinet

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/chess-vn/slchess/pkg/logging"
	"go.uber.org/zap"
)

type options struct {
	threads  int
	hash     int
	multiPvs int
}

func defaultOptions() options {
	return options{
		threads:  1,
		hash:     128,
		multiPvs: 3,
	}
}

func parsePvsLines(lines []string) Evaluation {
	// Improved regex pattern
	re := regexp.MustCompile(`depth (\d+).*?score cp (-?\d+).*?nodes (\d+).*?pv (.+)`)

	var eval Evaluation
	eval.Pvs = []Pv{}

	for _, line := range lines {
		match := re.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		depth, err1 := strconv.Atoi(match[1])
		cp, err2 := strconv.Atoi(match[2])
		nodes, err3 := strconv.Atoi(match[3])
		moves := strings.TrimSpace(match[4])

		if err1 != nil || err2 != nil || err3 != nil {
			logging.Error(
				"Error converting values:",
				zap.Error(err1),
				zap.Error(err2),
				zap.Error(err3),
			)
			continue
		}

		// Set depth and knodes once
		if eval.Depth == 0 {
			eval.Depth = depth
			eval.Knodes = nodes
		}

		// Append move sequence
		eval.Pvs = append(eval.Pvs, Pv{
			Cp:    cp,
			Moves: moves,
		})
	}

	return eval
}

func runStockfish(path string, fen string, depth int, opts options) ([]string, error) {
	cmd := exec.Command(path)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	writer := bufio.NewWriter(stdin)
	reader := bufio.NewScanner(stdout)

	options := []string{
		"uci",
		fmt.Sprintf("setoption name Threads value %d", opts.threads),
		fmt.Sprintf("setoption name Hash value %d", opts.hash),
		fmt.Sprintf("setoption name MultiPV value %d", opts.multiPvs),
		"isready",
	}
	for _, option := range options {
		fmt.Fprintln(writer, option)
	}
	writer.Flush()

	// Wait for Stockfish to be ready
	for reader.Scan() {
		if reader.Text() == "readyok" {
			break
		}
	}

	// Set position and start analysis
	fmt.Fprintln(writer, "position fen "+fen)
	fmt.Fprintf(writer, "go depth %d\n", depth)
	writer.Flush()

	// Read Stockfish output and extract MultiPV lines
	stopStr := fmt.Sprintf("info depth %d", depth)
	var pvLines []string
	for reader.Scan() {
		line := reader.Text()
		if strings.Contains(line, "bestmove") {
			break // Stop reading once bestmove is received
		}
		if strings.Contains(line, stopStr) && strings.Contains(line, " multipv ") {
			pvLines = append(pvLines, line)
		}
	}

	stdin.Close()
	stdout.Close()
	cmd.Wait()
	return pvLines, nil
}
