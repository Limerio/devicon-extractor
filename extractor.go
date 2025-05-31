package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
)

type IconExtractor struct {
	cloneDir  string
	outputDir string
}

func NewIconExtractor() *IconExtractor {
	return &IconExtractor{
		cloneDir:  CloneDir,
		outputDir: OutputDir,
	}
}

func (ie *IconExtractor) Run() error {
	steps := []struct {
		name string
		fn   func() error
	}{
		{"Clone repository", ie.CloneRepository},
		{"Cleanup clone", ie.CleanupClone},
		{"Create output directory", ie.CreateOutputDirectory},
		{"Extract SVG icons", ie.ExtractSVGIcons},
		{"Cleanup temporary files", ie.Cleanup},
	}

	for _, step := range steps {
		if err := step.fn(); err != nil {
			if step.name == "Cleanup temporary files" {
				log.Printf("Warning: %s failed: %v", step.name, err)
				continue
			}
			return fmt.Errorf("%s failed: %w", step.name, err)
		}
	}

	return nil
}

func (ie *IconExtractor) CloneRepository() error {
	log.Println("Cloning devicons repository...")

	if err := os.RemoveAll(ie.cloneDir); err != nil {
		return fmt.Errorf("failed to remove existing clone directory: %w", err)
	}

	cmd := exec.Command("git", "clone", DeviconRepo, ie.cloneDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to clone repository: %w\nOutput: %s", err, output)
	}

	log.Println("Repository cloned successfully")
	return nil
}

func (ie *IconExtractor) CleanupClone() error {
	log.Println("Cleaning up cloned repository...")

	iconsPath := filepath.Join(ie.cloneDir, IconsDir)
	if _, err := os.Stat(iconsPath); os.IsNotExist(err) {
		return fmt.Errorf("icons directory not found in cloned repository")
	}

	entries, err := os.ReadDir(ie.cloneDir)
	if err != nil {
		return fmt.Errorf("failed to read clone directory: %w", err)
	}

	for _, entry := range entries {
		if entry.Name() != IconsDir {
			entryPath := filepath.Join(ie.cloneDir, entry.Name())
			if err := os.RemoveAll(entryPath); err != nil {
				log.Printf("Warning: failed to remove %s: %v", entryPath, err)
			}
		}
	}

	log.Println("Cleanup completed")
	return nil
}

func (ie *IconExtractor) CreateOutputDirectory() error {
	if err := os.RemoveAll(ie.outputDir); err != nil {
		return fmt.Errorf("failed to remove existing output directory: %w", err)
	}

	if err := os.MkdirAll(ie.outputDir, DirPermissions); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	log.Printf("Output directory created: %s", ie.outputDir)
	return nil
}

// TechDirJob represents a technology directory processing job
type TechDirJob struct {
	techName string
	techPath string
}

// ProcessingResult represents the result of processing a technology directory
type ProcessingResult struct {
	techName string
	success  bool
	error    error
}

func (ie *IconExtractor) ExtractSVGIcons() error {
	log.Println("Extracting SVG icons...")

	iconsPath := filepath.Join(ie.cloneDir, IconsDir)

	entries, err := os.ReadDir(iconsPath)
	if err != nil {
		return fmt.Errorf("failed to read icons directory: %w", err)
	}

	// Filter directories and create jobs
	var jobs []TechDirJob
	for _, entry := range entries {
		if entry.IsDir() {
			jobs = append(jobs, TechDirJob{
				techName: entry.Name(),
				techPath: filepath.Join(iconsPath, entry.Name()),
			})
		}
	}

	// Process directories in parallel
	processedCount, skippedCount := ie.processDirectoriesParallel(jobs)

	log.Printf("Extraction completed. Processed: %d, Skipped: %d", processedCount, skippedCount)
	return nil
}

func (ie *IconExtractor) processDirectoriesParallel(jobs []TechDirJob) (int64, int64) {
	// Determine optimal number of workers (don't overwhelm the filesystem)
	numWorkers := runtime.NumCPU()
	if numWorkers > 8 {
		numWorkers = 8 // Cap at 8 to avoid too many concurrent file operations
	}
	if len(jobs) < numWorkers {
		numWorkers = len(jobs)
	}

	log.Printf("Processing %d directories using %d workers...", len(jobs), numWorkers)

	// Channels for communication
	jobChan := make(chan TechDirJob, len(jobs))
	resultChan := make(chan ProcessingResult, len(jobs))

	// Atomic counters for thread-safe counting
	var processedCount, skippedCount int64

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go ie.worker(jobChan, resultChan, &wg)
	}

	// Send jobs to workers
	go func() {
		defer close(jobChan)
		for _, job := range jobs {
			jobChan <- job
		}
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Process results and update counters
	for result := range resultChan {
		if result.success {
			atomic.AddInt64(&processedCount, 1)
		} else {
			atomic.AddInt64(&skippedCount, 1)
			log.Printf("Warning: failed to process %s: %v", result.techName, result.error)
		}
	}

	return processedCount, skippedCount
}

func (ie *IconExtractor) worker(jobChan <-chan TechDirJob, resultChan chan<- ProcessingResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobChan {
		err := ie.processTechDirectory(job.techPath, job.techName)
		result := ProcessingResult{
			techName: job.techName,
			success:  err == nil,
			error:    err,
		}
		resultChan <- result
	}
}

func (ie *IconExtractor) processTechDirectory(techPath, techName string) error {
	fileUtils := NewFileUtils()
	svgFiles, err := fileUtils.FindSVGFiles(techPath)
	if err != nil {
		return fmt.Errorf("failed to find SVG files: %w", err)
	}

	if len(svgFiles) == 0 {
		return fmt.Errorf("no SVG files found")
	}

	selectedFile := ie.selectBestSVG(svgFiles, techName)
	if selectedFile == "" {
		return fmt.Errorf("no suitable SVG file found")
	}

	outputFileName := fmt.Sprintf("%s.svg", techName)
	outputPath := filepath.Join(ie.outputDir, outputFileName)

	if err := fileUtils.CopyFile(selectedFile, outputPath); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	log.Printf("Extracted: %s -> %s", selectedFile, outputFileName)
	return nil
}

func (ie *IconExtractor) selectBestSVG(svgFiles []string, techName string) string {
	for _, file := range svgFiles {
		fileName := filepath.Base(file)
		if strings.Contains(strings.ToLower(fileName), "original") {
			return file
		}
	}

	for _, file := range svgFiles {
		fileName := strings.TrimSuffix(filepath.Base(file), ".svg")
		if strings.EqualFold(fileName, techName) {
			return file
		}
	}

	for _, file := range svgFiles {
		fileName := filepath.Base(file)
		if strings.Contains(strings.ToLower(fileName), "plain") {
			return file
		}
	}

	if len(svgFiles) == 1 {
		return svgFiles[0]
	}

	if len(svgFiles) > 0 {
		return svgFiles[0]
	}

	return ""
}

func (ie *IconExtractor) Cleanup() error {
	log.Println("Cleaning up temporary files...")
	if err := os.RemoveAll(ie.cloneDir); err != nil {
		return fmt.Errorf("failed to cleanup clone directory: %w", err)
	}
	log.Println("Cleanup completed")
	return nil
}
