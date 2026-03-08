// Infrastructure Validator for go-agent
// Scans Docker containers, validates Qdrant collections, tests API connectivity
// Auto-fixes common issues and runs monitoring loop during development
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

var (
	flagOnce          = flag.Bool("once", false, "Run validation once and exit")
	flagInterval      = flag.Duration("interval", 10*time.Minute, "Monitoring interval (0 to disable)")
	flagQdrantURL     = flag.String("qdrant-url", "http://localhost:6333", "Qdrant base URL")
	flagQdrantAPIKey  = flag.String("qdrant-api-key", "", "Qdrant API key (optional)")
	flagExpectedDim   = flag.Int("expected-dim", 768, "Expected vector dimension for collections")
	flagAutoFix       = flag.Bool("auto-fix", true, "Automatically fix detected issues")
	flagReportDir     = flag.String("report-dir", "", "Directory to save validation reports")
	flagJSON          = flag.Bool("json", false, "Output in JSON format")
	flagVerbose       = flag.Bool("v", false, "Verbose output")
)

// ValidationResult represents the result of a single validation check
type ValidationResult struct {
	Name        string    `json:"name"`
	Status      string    `json:"status"` // "ok", "warning", "error", "fixed"
	Message     string    `json:"message,omitempty"`
	Details     any       `json:"details,omitempty"`
	FixedBy     string    `json:"fixed_by,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// ValidationReport is the complete validation report
type ValidationReport struct {
	Timestamp       time.Time            `json:"timestamp"`
	Duration       time.Duration        `json:"duration"`
	OverallStatus  string               `json:"overall_status"` // "healthy", "degraded", "unhealthy"
	Docker         []ValidationResult   `json:"docker,omitempty"`
	Qdrant         []ValidationResult   `json:"qdrant,omitempty"`
	APIs           []ValidationResult   `json:"apis,omitempty"`
	Environment    []ValidationResult   `json:"environment,omitempty"`
	Summary        Summary              `json:"summary"`
}

// Summary contains counts for the report
type Summary struct {
	TotalChecks  int `json:"total_checks"`
	Passed       int `json:"passed"`
	Warnings     int `json:"warnings"`
	Errors       int `json:"errors"`
	Fixed        int `json:"fixed"`
}

// Validator is the main validator struct
type Validator struct {
	qdrantURL      string
	qdrantAPIKey   string
	expectedDim    int
	autoFix        bool
	verbose        bool
	reportDir      string

	mu             sync.RWMutex
	lastReport     *ValidationReport
}

// NewValidator creates a new validator instance
func NewValidator(qdrantURL, qdrantAPIKey string, expectedDim int, autoFix, verbose bool, reportDir string) *Validator {
	return &Validator{
		qdrantURL:    qdrantURL,
		qdrantAPIKey: qdrantAPIKey,
		expectedDim:  expectedDim,
		autoFix:      autoFix,
		verbose:      verbose,
		reportDir:    reportDir,
	}
}

// RunAll runs all validations and returns a report
func (v *Validator) RunAll(ctx context.Context) *ValidationReport {
	start := time.Now()
	report := &ValidationReport{
		Timestamp: start,
	}

	// Run all validation checks
	report.Docker = v.validateDocker(ctx)
	report.Qdrant = v.validateQdrant(ctx)
	report.APIs = v.validateAPIs(ctx)
	report.Environment = v.validateEnvironment(ctx)

	// Calculate summary
	v.calculateSummary(report)

	// Determine overall status
	report.OverallStatus = v.determineOverallStatus(report)
	report.Duration = time.Since(start)

	// Store last report
	v.mu.Lock()
	v.lastReport = report
	v.mu.Unlock()

	// Save report if directory specified
	if v.reportDir != "" {
		v.saveReport(report)
	}

	return report
}

func (v *Validator) calculateSummary(report *ValidationReport) {
	for _, r := range report.Docker {
		report.Summary.TotalChecks++
		switch r.Status {
		case "ok": report.Summary.Passed++
		case "warning": report.Summary.Warnings++
		case "error": report.Summary.Errors++
		case "fixed": report.Summary.Fixed++; report.Summary.Passed++
		}
	}
	for _, r := range report.Qdrant {
		report.Summary.TotalChecks++
		switch r.Status {
		case "ok": report.Summary.Passed++
		case "warning": report.Summary.Warnings++
		case "error": report.Summary.Errors++
		case "fixed": report.Summary.Fixed++; report.Summary.Passed++
		}
	}
	for _, r := range report.APIs {
		report.Summary.TotalChecks++
		switch r.Status {
		case "ok": report.Summary.Passed++
		case "warning": report.Summary.Warnings++
		case "error": report.Summary.Errors++
		case "fixed": report.Summary.Fixed++; report.Summary.Passed++
		}
	}
	for _, r := range report.Environment {
		report.Summary.TotalChecks++
		switch r.Status {
		case "ok": report.Summary.Passed++
		case "warning": report.Summary.Warnings++
		case "error": report.Summary.Errors++
		case "fixed": report.Summary.Fixed++; report.Summary.Passed++
		}
	}
}

func (v *Validator) determineOverallStatus(report *ValidationReport) string {
	if report.Summary.Errors > 0 {
		return "unhealthy"
	}
	if report.Summary.Warnings > 0 {
		return "degraded"
	}
	return "healthy"
}

func (v *Validator) saveReport(report *ValidationReport) {
	if err := os.MkdirAll(v.reportDir, 0755); err != nil {
		log.Printf("Failed to create report directory: %v", err)
		return
	}

	filename := filepath.Join(v.reportDir, fmt.Sprintf("validation-%s.json", report.Timestamp.Format("20060102-150405")))
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal report: %v", err)
		return
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		log.Printf("Failed to save report: %v", err)
		return
	}

	if v.verbose {
		log.Printf("Report saved to: %s", filename)
	}
}

// GetLastReport returns the last validation report
func (v *Validator) GetLastReport() *ValidationReport {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.lastReport
}

// Print prints the validation report
func (r *ValidationReport) Print(jsonFormat bool) {
	if jsonFormat {
		data, _ := json.MarshalIndent(r, "", "  ")
		fmt.Println(string(data))
		return
	}

	// Human-readable output
	fmt.Printf("\n╔══════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  Infrastructure Validation Report                              ║\n")
	fmt.Printf("║  %s | Duration: %v                      ║\n", r.Timestamp.Format("2006-01-02 15:04:05"), r.Duration.Round(time.Millisecond))
	fmt.Printf("╠══════════════════════════════════════════════════════════════╣\n")

	// Status indicator
	statusIcon := map[string]string{
		"healthy":   "✅",
		"degraded":  "⚠️",
		"unhealthy": "❌",
	}
	fmt.Printf("║  Overall: %s %s                                              ║\n", 
		statusIcon[r.OverallStatus], r.OverallStatus)
	fmt.Printf("╠══════════════════════════════════════════════════════════════╣\n")

	// Summary
	fmt.Printf("║  Summary: %d checks | %d passed | %d warnings | %d errors | %d fixed  ║\n",
		r.Summary.TotalChecks, r.Summary.Passed, r.Summary.Warnings, r.Summary.Errors, r.Summary.Fixed)
	fmt.Printf("╠══════════════════════════════════════════════════════════════╣\n")

	// Docker
	if len(r.Docker) > 0 {
		fmt.Printf("║  🐳 Docker Containers                                         ║\n")
		for _, result := range r.Docker {
			r.printResult("║    ", result)
		}
	}

	// Qdrant
	if len(r.Qdrant) > 0 {
		fmt.Printf("║  🗄️  Qdrant Vector Database                                   ║\n")
		for _, result := range r.Qdrant {
			r.printResult("║    ", result)
		}
	}

	// APIs
	if len(r.APIs) > 0 {
		fmt.Printf("║  🔌 API Connectivity                                          ║\n")
		for _, result := range r.APIs {
			r.printResult("║    ", result)
		}
	}

	// Environment
	if len(r.Environment) > 0 {
		fmt.Printf("║  🔧 Environment Variables                                     ║\n")
		for _, result := range r.Environment {
			r.printResult("║    ", result)
		}
	}

	fmt.Printf("╚══════════════════════════════════════════════════════════════╝\n\n")
}

func (r *ValidationReport) printResult(prefix string, result ValidationResult) {
	statusIcon := map[string]string{
		"ok":      "✅",
		"warning": "⚠️",
		"error":   "❌",
		"fixed":   "🔧",
	}
	fmt.Printf("%s%s %s: %s\n", prefix, statusIcon[result.Status], result.Name, result.Message)
	if result.FixedBy != "" {
		fmt.Printf("%s   └─ Fixed by: %s\n", prefix, result.FixedBy)
	}
}

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	validator := NewValidator(
		*flagQdrantURL,
		*flagQdrantAPIKey,
		*flagExpectedDim,
		*flagAutoFix,
		*flagVerbose,
		*flagReportDir,
	)

	// Run initial validation
	fmt.Println("🔍 Running infrastructure validation...")
	report := validator.RunAll(ctx)
	report.Print(*flagJSON)

	// If -once flag is set, exit after one run
	if *flagOnce {
		if report.OverallStatus == "unhealthy" {
			os.Exit(1)
		}
		return
	}

	// If interval is 0, exit after one run
	if *flagInterval == 0 {
		return
	}

	// Start monitoring loop
	fmt.Printf("\n📡 Starting monitoring loop (interval: %v)...\n", *flagInterval)
	fmt.Println("Press Ctrl+C to stop")

	ticker := time.NewTicker(*flagInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case sig := <-sigChan:
			fmt.Printf("\nReceived signal %v, shutting down...\n", sig)
			cancel()
			return
		case <-ticker.C:
			fmt.Printf("\n🔍 [%s] Running periodic validation...\n", time.Now().Format("15:04:05"))
			report := validator.RunAll(ctx)
			report.Print(*flagJSON)

			if report.Summary.Fixed > 0 {
				fmt.Printf("🔧 Auto-fixed %d issue(s)\n", report.Summary.Fixed)
			}
		}
	}
}