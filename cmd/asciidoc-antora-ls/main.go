// Command asciidoc-antora-ls runs the AsciiDoc and Antora language server.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/bovinemagnet/asciidoc-antora-ls/pkg/lsp"
)

func main() {
	log.SetOutput(os.Stderr)
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("asciidoc-antora-ls", flag.ContinueOnError)
	flags.SetOutput(stderr)
	showVersion := flags.Bool("version", false, "print the version and exit")
	logFilePath := flags.String("log-file", "", "write server logs to this file")
	logLevelName := flags.String("log-level", "info", "log level: error, info, or debug")
	flags.Bool("stdio", false, "use standard input and output for LSP transport")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	if *showVersion {
		if _, err := fmt.Fprintln(stdout, lsp.VersionString()); err != nil {
			writeError(stderr, "write version: %v\n", err)
			return 1
		}
		return 0
	}

	level, err := lsp.ParseLogLevel(*logLevelName)
	if err != nil {
		writeError(stderr, "%v\n", err)
		return 2
	}

	logOutput := stderr
	var logFile *os.File
	if *logFilePath != "" {
		logFile, err = os.OpenFile(*logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			writeError(stderr, "open log file: %v\n", err)
			return 1
		}
		logOutput = logFile
	}

	logger := log.New(logOutput, "", log.LstdFlags)
	server := lsp.NewServer(lsp.WithLogger(logger), lsp.WithLogLevel(level))
	runErr := server.Run(context.Background(), stdin, stdout)
	exitCode := 0
	if runErr != nil {
		logger.Printf("language server failed: %v", runErr)
		exitCode = 1
	}

	if logFile != nil {
		if err := logFile.Close(); err != nil {
			writeError(stderr, "close log file: %v\n", err)
			return 1
		}
	}
	return exitCode
}

func writeError(writer io.Writer, format string, arguments ...any) {
	_, _ = fmt.Fprintf(writer, "asciidoc-antora-ls: "+format, arguments...)
}
