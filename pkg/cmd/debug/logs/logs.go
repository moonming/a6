package logs

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/api7/a6/internal/config"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/iostreams"
)

type Options struct {
	IO     *iostreams.IOStreams
	Config func() (config.Config, error)

	Follow    bool
	Tail      int
	Since     string
	LogType   string
	Container string
	FilePath  string
	Output    string
}

func NewCmdLogs(f *cmd.Factory) *cobra.Command {
	opts := &Options{
		IO:      f.IOStreams,
		Config:  f.Config,
		Tail:    100,
		LogType: "all",
	}

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Stream APISIX logs from Docker or files",
		RunE: func(_ *cobra.Command, _ []string) error {
			return logsRun(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Follow, "follow", "f", false, "Stream logs continuously")
	cmd.Flags().IntVarP(&opts.Tail, "tail", "n", 100, "Number of recent lines to show")
	cmd.Flags().StringVar(&opts.Since, "since", "", "Show logs since duration (e.g., 5m, 1h, 24h)")
	cmd.Flags().StringVarP(&opts.LogType, "type", "t", "all", "Log type: error, access, all")
	cmd.Flags().StringVarP(&opts.Container, "container", "c", "", "Docker container name (auto-detect if empty)")
	cmd.Flags().StringVar(&opts.FilePath, "file", "", "Path to log file (use file tailing instead of Docker)")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output format: table, json, yaml")

	return cmd
}

func logsRun(opts *Options) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if opts.FilePath != "" {
		return tailFile(ctx, opts)
	}

	container := opts.Container
	if container == "" {
		autoContainer, err := detectContainer()
		if err != nil {
			return err
		}
		container = autoContainer
	}

	return runDockerLogs(ctx, opts, container)
}

func detectContainer() (string, error) {
	cmd := exec.Command("docker", "ps", "--filter", "name=apisix", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", fmt.Errorf("docker binary not found in PATH")
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			stderr := strings.TrimSpace(string(exitErr.Stderr))
			if stderr == "" {
				stderr = strings.TrimSpace(err.Error())
			}
			return "", fmt.Errorf("failed to detect APISIX containers: %s", stderr)
		}
		return "", fmt.Errorf("failed to detect APISIX containers: %w", err)
	}

	containers := parseDockerPSNames(string(output))
	return chooseContainer(containers)
}

func runDockerLogs(ctx context.Context, opts *Options, container string) error {
	args := buildDockerArgs(opts, container)
	cmd := exec.CommandContext(ctx, "docker", args...)
	var stderr bytes.Buffer
	cmd.Stdout = opts.IO.Out
	cmd.Stderr = io.MultiWriter(opts.IO.Out, &stderr)

	err := cmd.Run()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return fmt.Errorf("docker binary not found in PATH")
		}
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return fmt.Errorf("%s", msg)
		}
		return err
	}

	return nil
}

func buildDockerArgs(opts *Options, container string) []string {
	args := []string{"logs"}
	if opts.Follow {
		args = append(args, "--follow")
	}
	if opts.Tail > 0 {
		args = append(args, "--tail", strconv.Itoa(opts.Tail))
	}
	if opts.Since != "" {
		args = append(args, "--since", opts.Since)
	}
	return append(args, container)
}

func parseDockerPSNames(output string) []string {
	lines := strings.Split(output, "\n")
	containers := make([]string, 0, len(lines))
	for _, line := range lines {
		name := strings.TrimSpace(line)
		if name != "" {
			containers = append(containers, name)
		}
	}
	return containers
}

func chooseContainer(containers []string) (string, error) {
	if len(containers) == 0 {
		return "", fmt.Errorf("no APISIX container found. Use --container to specify or --file to tail a log file directly")
	}
	if len(containers) > 1 {
		return "", fmt.Errorf("multiple APISIX containers found: %s. Use --container to specify one", strings.Join(containers, ", "))
	}

	return containers[0], nil
}

func tailFile(ctx context.Context, opts *Options) error {
	if opts.Tail > 0 {
		lines, err := readLastLines(opts.FilePath, opts.Tail)
		if err != nil {
			return err
		}
		for _, line := range lines {
			_, _ = fmt.Fprintln(opts.IO.Out, line)
		}
	}

	if !opts.Follow {
		return nil
	}

	file, err := os.Open(opts.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open log file %q: %w", opts.FilePath, err)
	}
	defer file.Close()

	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		return fmt.Errorf("failed to seek log file %q: %w", opts.FilePath, err)
	}

	reader := bufio.NewReader(file)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			line, err := reader.ReadString('\n')
			if err == nil {
				_, _ = io.WriteString(opts.IO.Out, line)
				continue
			}

			if errors.Is(err, io.EOF) {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			return fmt.Errorf("failed reading log file %q: %w", opts.FilePath, err)
		}
	}
}

func readLastLines(filePath string, n int) ([]string, error) {
	if n <= 0 {
		return nil, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %q: %w", filePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := make([]string, 0, n)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > n {
			lines = lines[1:]
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read log file %q: %w", filePath, err)
	}

	return lines, nil
}
