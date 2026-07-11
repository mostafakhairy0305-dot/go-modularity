// Command go-modularity computes type-level and package-level modularity
// metrics for a Go module.
//
//	go-modularity [flags] [patterns...]
//	go-modularity ./...
//	go-modularity ./... --format=json
//	go-modularity --web ./...
//	go-modularity ./internal/... --metrics=amc,lcom1,lcom96b,tcc
//	go-modularity --help --web
//
// Logs go to stderr; the report goes to stdout or --output. --help --web
// opens an illustrated guide to every reported metric in the browser.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	gomodularity "github.com/mostafakhairy0305-dot/go-modularity"
	reporting "github.com/mostafakhairy0305-dot/go-modularity/internal/features/reporting/application"
	reportingdomain "github.com/mostafakhairy0305-dot/go-modularity/internal/features/reporting/domain"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/features/reporting/ports/outbound"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/infrastructure/browser"
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
		fmt.Fprintf(os.Stderr, "\nFor an illustrated guide to every reported metric:\n  go-modularity --help --web\n")
	}

	var (
		format          = fs.String("format", "text", "report format: text, json, csv, or web")
		webReport       = fs.Bool("web", false, "shorthand for -format=web: write a self-contained HTML report and open it")
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
		// --help --web (either order) opens the metrics guide instead of
		// failing; parsing aborts at --help, so the raw args are scanned.
		if errors.Is(err, flag.ErrHelp) && wantsWebHelp(args) {
			return runWebHelp()
		}

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

	if *webReport {
		if formatWasSet(fs) && *format != string(reportingdomain.FormatWeb) {
			logger.Error("conflicting flags: -web implies -format=web", "format", *format)

			return 2
		}

		*format = string(reportingdomain.FormatWeb)
	}

	reportFormat, ok := reportingdomain.ParseFormat(*format)
	if !ok {
		logger.Error("invalid format", "format", *format, "want", "text, json, csv, or web")

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
			err := stopProfile()
			if err != nil {
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

		// A signal (Ctrl-C / SIGTERM) cancels the context; report it with
		// the conventional 130 so CI can tell cancellation from failure.
		if errors.Is(err, context.Canceled) {
			return 130
		}

		return 1
	}

	logger.Debug("analysis complete",
		"packages", len(report.Packages), "duration", time.Since(start))

	if *memoryProfile != "" {
		err := profiling.WriteHeap(*memoryProfile)
		if err != nil {
			logger.Error("memory profiling failed", "error", err)

			return 1
		}
	}

	outputPath := *output

	// A web report is unreadable on stdout: default it to a well-known file.
	webToDefaultFile := outputPath == "" && reportFormat == reportingdomain.FormatWeb
	if webToDefaultFile {
		outputPath = defaultWebReportName
	}

	var sink outbound.Sink = sinks.StdoutSink{}
	if outputPath != "" {
		sink = sinks.FileSink{Path: outputPath}
	}

	textOptions := reportingdomain.TextOptions{
		Color:   outputPath == "" && os.Getenv("NO_COLOR") == "" && stdoutIsTerminal(),
		Explain: *explain,
	}
	if err := reporting.Write(report, reportFormat, sink, textOptions); err != nil {
		logger.Error("writing report failed", "error", err)

		return 1
	}

	// Open the defaulted web report only when a human is watching; an
	// explicit -output signals scripting, and pipes signal CI.
	if webToDefaultFile {
		logger.Info("report written", "path", outputPath)

		if stdoutIsTerminal() {
			if err := browser.Open(outputPath); err != nil {
				logger.Warn("opening the report in a browser failed", "error", err)
			}
		}
	}

	return 0
}

// defaultWebReportName is where the web report lands when -output is unset.
const defaultWebReportName = "modularity-report.html"

// formatWasSet reports whether the -format flag was given explicitly.
func formatWasSet(fs *flag.FlagSet) bool {
	set := false
	fs.Visit(func(f *flag.Flag) { set = set || f.Name == "format" })

	return set
}

// wantsWebHelp reports whether the raw arguments request the web metrics
// guide: a truthy -web / --web token before any "--" terminator. Raw args
// are scanned because --help aborts flag parsing before -web is seen.
func wantsWebHelp(args []string) bool {
	for _, arg := range args {
		if arg == "--" {
			return false
		}

		if arg == "-web" || arg == "--web" {
			return true
		}

		value, found := strings.CutPrefix(arg, "-web=")
		if !found {
			value, found = strings.CutPrefix(arg, "--web=")
		}

		if found {
			truthy, err := strconv.ParseBool(value)

			return err == nil && truthy
		}
	}

	return false
}

// runWebHelp writes the metrics guide to the OS temp dir and, when a human
// is watching, opens it in the browser. The file is always written so
// scripts and CI can read the logged path.
func runWebHelp() int {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))

	path, err := writeHelpDocs()
	if err != nil {
		logger.Error("writing the metrics guide failed", "error", err)

		return 1
	}

	logger.Info("metrics guide written", "path", path)

	if stdoutIsTerminal() {
		if err := browser.Open(path); err != nil {
			logger.Warn("opening the metrics guide in a browser failed", "error", err)
		}
	}

	return 0
}

// writeHelpDocs renders the metrics guide into a fresh temp file and
// returns its path.
func writeHelpDocs() (string, error) {
	file, err := os.CreateTemp("", "go-modularity-help-*.html")
	if err != nil {
		return "", err
	}

	path := file.Name()
	if err := file.Close(); err != nil {
		return "", err
	}

	if err := reporting.WriteDocs(sinks.FileSink{Path: path}, version.Version); err != nil {
		return "", err
	}

	return path, nil
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
