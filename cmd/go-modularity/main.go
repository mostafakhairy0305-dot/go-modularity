// Command go-modularity computes type-level and package-level modularity
// metrics for a Go module.
//
//	go-modularity [flags] [patterns...]
//	go-modularity ./...
//	go-modularity ./... --format=json
//	go-modularity ./internal/... --metrics=amc,lcom1,lcom96b,tcc
//
// Logs go to stderr; the report goes to stdout or --output.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	gomodularity "github.com/mostafakhairy0305-dot/go-modularity"
	reporting "github.com/mostafakhairy0305-dot/go-modularity/internal/features/reporting/application"
	reportingdomain "github.com/mostafakhairy0305-dot/go-modularity/internal/features/reporting/domain"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/features/reporting/ports/outbound"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/infrastructure/profiling"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/infrastructure/sinks"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/version"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("go-modularity", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: go-modularity [flags] [patterns...]\n\n")
		fs.PrintDefaults()
	}

	var (
		format          = fs.String("format", "text", "report format: text, json, or csv")
		output          = fs.String("output", "", "write the report to this file instead of stdout")
		explain         = fs.Bool("explain", false, "include reasons for n/a and dropped-component metrics in the text report")
		metricList      = fs.String("metrics", "", "comma-separated metrics to report (default: all except cbo)")
		workers         = fs.Int("workers", 0, "concurrent package workers (0 = min(GOMAXPROCS, packages))")
		fieldUsage      = fs.String("field-usage", "direct", "field usage resolution: direct or transitive")
		dependencyScope = fs.String("dependency-scope", "module", "dependency scope: project, module, or all")
		buildTags       = fs.String("build-tags", "", "comma-separated build tags")
		includeTests    = fs.Bool("tests", false, "include test files and test packages")
		generated       = fs.Bool("generated", false, "include generated files")
		continueOnError = fs.Bool("continue-on-error", false, "skip packages that fail to load or type-check")
		cpuProfile      = fs.String("cpu-profile", "", "write a CPU profile to this file")
		memoryProfile   = fs.String("memory-profile", "", "write a memory profile to this file")
		showVersion     = fs.Bool("version", false, "print the version and exit")
		verbose         = fs.Bool("verbose", false, "verbose logging to stderr")
	)
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *showVersion {
		fmt.Println("go-modularity " + version.Version)
		return 0
	}

	level := slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	reportFormat, ok := reportingdomain.ParseFormat(*format)
	if !ok {
		logger.Error("invalid format", "format", *format, "want", "text, json, or csv")
		return 2
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if *cpuProfile != "" {
		stopProfile, err := profiling.StartCPU(*cpuProfile)
		if err != nil {
			logger.Error("cpu profiling failed", "error", err)
			return 1
		}
		defer func() {
			if err := stopProfile(); err != nil {
				logger.Error("cpu profiling failed", "error", err)
			}
		}()
	}

	config := gomodularity.Config{
		Patterns:         fs.Args(),
		IncludeTests:     *includeTests,
		IncludeGenerated: *generated,
		BuildTags:        splitList(*buildTags),
		Workers:          *workers,
		DependencyScope:  gomodularity.DependencyScope(*dependencyScope),
		FieldUsageMode:   gomodularity.FieldUsageMode(*fieldUsage),
		SelectedMetrics:  parseMetrics(*metricList),
		ContinueOnError:  *continueOnError,
	}

	start := time.Now()
	report, err := gomodularity.Analyze(ctx, config)
	if err != nil {
		logger.Error("analysis failed", "error", err)
		return 1
	}
	logger.Debug("analysis complete",
		"packages", len(report.Packages), "duration", time.Since(start))

	if *memoryProfile != "" {
		if err := profiling.WriteHeap(*memoryProfile); err != nil {
			logger.Error("memory profiling failed", "error", err)
			return 1
		}
	}

	var sink outbound.Sink = sinks.StdoutSink{}
	if *output != "" {
		sink = sinks.FileSink{Path: *output}
	}
	textOptions := reportingdomain.TextOptions{
		Color:   *output == "" && os.Getenv("NO_COLOR") == "" && stdoutIsTerminal(),
		Explain: *explain,
	}
	if err := reporting.Write(report, reportFormat, sink, textOptions); err != nil {
		logger.Error("writing report failed", "error", err)
		return 1
	}
	return 0
}

// stdoutIsTerminal reports whether stdout is a character device, so ANSI
// colors never leak into pipes or redirected files.
func stdoutIsTerminal() bool {
	info, err := os.Stdout.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func splitList(list string) []string {
	if list == "" {
		return nil
	}
	parts := strings.Split(list, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

func parseMetrics(list string) []gomodularity.MetricName {
	parts := splitList(list)
	if len(parts) == 0 {
		return nil
	}
	out := make([]gomodularity.MetricName, len(parts))
	for i, part := range parts {
		out[i] = gomodularity.MetricName(part)
	}
	return out
}
