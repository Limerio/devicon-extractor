package main

import "log"

func main() {
	log.Println("Starting devicon SVG extraction process...")

	extractor := NewIconExtractor()

	if err := extractor.Run(); err != nil {
		log.Fatalf("Extraction process failed: %v", err)
	}

	log.Printf("Process completed successfully! Icons extracted to: %s", OutputDir)
}
