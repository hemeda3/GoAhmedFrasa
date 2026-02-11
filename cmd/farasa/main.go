package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"farasa/pkg/farasa"
)

func main() {
	inputFile := flag.String("i", "", "Input file path")
	outputFile := flag.String("o", "", "Output file path")
	scheme := flag.String("c", "", "Segmentation scheme (atb)")
	normFlag := flag.Bool("n", true, "Normalization (true/false)")
	dataDir := flag.String("d", "", "Data directory path")
	flag.Parse()

	// Determine data directory
	dir := *dataDir
	if dir == "" {
		dir = os.Getenv("FarasaDataDir")
	}
	if dir == "" {
		// Try default relative path
		dir = "data/"
	}
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}

	fmt.Fprint(os.Stderr, "Initializing the system ....")

	nbt, err := farasa.NewFarasa(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nError initializing Farasa: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprint(os.Stderr, "\r")
	fmt.Fprintln(os.Stderr, "System ready!               ")

	// Set up reader
	var reader *bufio.Reader
	if *inputFile != "" {
		f, err := os.Open(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		reader = bufio.NewReader(f)
	} else {
		reader = bufio.NewReader(os.Stdin)
	}

	// Set up writer
	var writer *bufio.Writer
	if *outputFile != "" {
		f, err := os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		writer = bufio.NewWriter(f)
	} else {
		writer = bufio.NewWriter(os.Stdout)
	}
	defer writer.Flush()

	processBuffer(reader, writer, nbt, *scheme, *normFlag)
}

func processBuffer(reader *bufio.Reader, writer *bufio.Writer, nbt *farasa.Farasa, scheme string, norm bool) {
	scanner := bufio.NewScanner(reader)
	// Increase scanner buffer for long lines
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		words := farasa.Tokenize(farasa.RemoveDiacritics(line))

		for _, w := range words {
			if cached, ok := nbt.HmSeenBefore[w]; !ok {
				solutions := nbt.MostLikelyPartition(farasa.Buck2UTF8(w), 1)
				topSolution := w
				if len(solutions) > 0 {
					topSolution = solutions[0].GetPartition()
				}
				topSolution = strings.ReplaceAll(strings.ReplaceAll(topSolution, ";", ""), "++", "+")
				nbt.HmSeenBefore[w] = topSolution

				if scheme == "atb" {
					topSolution = produceSpecialSegmentation(topSolution, nbt, norm)
				} else {
					if norm {
						topSolution = farasa.NormalizeFull(topSolution)
					}
				}
				nbt.HmSeenBefore[w] = topSolution
				writer.WriteString(topSolution + " ")
				writer.Flush()
			} else {
				var topSolution string
				if scheme == "atb" {
					cleaned := strings.ReplaceAll(strings.ReplaceAll(cached, ";", ""), "++", "+")
					topSolution = produceSpecialSegmentation(cleaned, nbt, norm)
				} else {
					topSolution = strings.ReplaceAll(strings.ReplaceAll(cached, ";", ""), "++", "+")
					if norm {
						topSolution = farasa.NormalizeFull(topSolution)
					}
				}
				writer.WriteString(topSolution + " ")
			}
		}
		writer.WriteString("\n")
	}
}

func produceSpecialSegmentation(segmentedWord string, nbt *farasa.Farasa, norm bool) string {
	tmp := nbt.GetProperSegmentation(segmentedWord)

	// attach Al to the word
	tmp = strings.ReplaceAll(tmp, "\u0627\u0644+;", ";\u0627\u0644")

	// attach ta marbouta
	tmp = strings.ReplaceAll(tmp, ";+\u0629", "\u0629;")

	// normalize output
	if norm {
		tmp = farasa.NormalizeFull(tmp)
	}

	// concat all prefixes and all suffixes
	parts := strings.Split(" "+tmp+" ", ";")

	output := ""

	// handle prefix
	prefixPart := strings.ReplaceAll(strings.TrimSpace(parts[0]), "+", "")
	if len(prefixPart) > 0 {
		output += prefixPart + "+ "
	}

	// handle stem
	output += strings.TrimSpace(parts[1])

	// handle suffix
	suffixPart := strings.ReplaceAll(strings.TrimSpace(parts[2]), "+", "")
	if len(suffixPart) > 0 {
		output += " +" + suffixPart
	}

	output = strings.TrimSpace(output)
	for strings.HasPrefix(output, "+") {
		output = output[1:]
	}
	for strings.HasSuffix(output, "+") {
		output = output[:len(output)-1]
	}
	return output
}
