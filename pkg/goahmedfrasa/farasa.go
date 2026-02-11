package goahmedfrasa

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Farasa is the core Arabic segmentation engine
type Farasa struct {
	hmPreviouslySeenTokenizations map[string][]string
	hmWordPossibleSplits          map[string][]string
	hmListMorph                   map[string]int
	hmListGaz                     map[string]int
	hmAraLexCom                   map[string]int
	hmBuck                        map[string]int
	hmLocations                   map[string]int
	hmPeople                      map[string]int
	hmStop                        map[string]int
	hPrefixes                     map[string]int
	hSuffixes                     map[string]int
	hmValidSuffixes               map[string]bool
	hmValidPrefixes               map[string]bool
	hmTemplateCount               map[string]float64
	HmSeenBefore                  map[string]string
	hmValidSuffixesSegmented      map[string]bool
	hmValidPrefixesSegmented      map[string]bool
	wordCount                     map[string]float64
	probPrefixes                  map[string]float64
	probSuffixes                  map[string]float64
	probCondPrefixes              map[string]float64
	probCondSuffixes              map[string]float64
	seenTemplates                 map[string]float64
	probPrefixSuffix              map[string]map[string]float64
	probSuffixPrefix              map[string]map[string]float64
	generalVariables              map[string]float64
	ft                            *FitTemplateClass
}

// NewFarasa creates a new Farasa instance and loads all data
func NewFarasa(dataDir string) (*Farasa, error) {
	f := &Farasa{
		hmPreviouslySeenTokenizations: make(map[string][]string),
		hmWordPossibleSplits:          make(map[string][]string),
		hmListMorph:                   make(map[string]int),
		hmListGaz:                     make(map[string]int),
		hmAraLexCom:                   make(map[string]int),
		hmBuck:                        make(map[string]int),
		hmLocations:                   make(map[string]int),
		hmPeople:                      make(map[string]int),
		hmStop:                        make(map[string]int),
		hPrefixes:                     make(map[string]int),
		hSuffixes:                     make(map[string]int),
		hmValidSuffixes:               make(map[string]bool),
		hmValidPrefixes:               make(map[string]bool),
		hmTemplateCount:               make(map[string]float64),
		HmSeenBefore:                  make(map[string]string),
		hmValidSuffixesSegmented:      make(map[string]bool),
		hmValidPrefixesSegmented:      make(map[string]bool),
		wordCount:                     make(map[string]float64),
		probPrefixes:                  make(map[string]float64),
		probSuffixes:                  make(map[string]float64),
		probCondPrefixes:              make(map[string]float64),
		probCondSuffixes:              make(map[string]float64),
		seenTemplates:                 make(map[string]float64),
		probPrefixSuffix:              make(map[string]map[string]float64),
		probSuffixPrefix:              make(map[string]map[string]float64),
		generalVariables:              make(map[string]float64),
	}

	var err error
	f.ft, err = NewFitTemplateClass(dataDir)
	if err != nil {
		return nil, fmt.Errorf("loading fit template: %w", err)
	}

	if err := f.loadStoredData(dataDir); err != nil {
		return nil, fmt.Errorf("loading stored data: %w", err)
	}
	return f, nil
}

func loadJSONFile(path string, target interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

// loadJSONIntMap loads a map[string]int from a JSON file where values may be float64
func loadJSONIntMap(path string) (map[string]int, error) {
	var raw map[string]float64
	if err := loadJSONFile(path, &raw); err != nil {
		return nil, err
	}
	result := make(map[string]int, len(raw))
	for k, v := range raw {
		result[k] = int(v)
	}
	return result, nil
}

func (f *Farasa) loadStoredData(dataDir string) error {
	var err error

	// Load int maps
	if f.hmListMorph, err = loadJSONIntMap(dataDir + "hmListMorph.json"); err != nil {
		return err
	}
	if f.hmListGaz, err = loadJSONIntMap(dataDir + "hmListGaz.json"); err != nil {
		return err
	}
	if f.hmAraLexCom, err = loadJSONIntMap(dataDir + "hmAraLexCom.json"); err != nil {
		return err
	}
	if f.hmBuck, err = loadJSONIntMap(dataDir + "hmBuck.json"); err != nil {
		return err
	}
	if f.hmLocations, err = loadJSONIntMap(dataDir + "hmLocations.json"); err != nil {
		return err
	}
	if f.hmPeople, err = loadJSONIntMap(dataDir + "hmPeople.json"); err != nil {
		return err
	}
	if f.hmStop, err = loadJSONIntMap(dataDir + "hmStop.json"); err != nil {
		return err
	}
	if f.hPrefixes, err = loadJSONIntMap(dataDir + "hPrefixes.json"); err != nil {
		return err
	}
	if f.hSuffixes, err = loadJSONIntMap(dataDir + "hSuffixes.json"); err != nil {
		return err
	}

	// Load bool maps
	if err = loadJSONFile(dataDir+"hmValidSuffixes.json", &f.hmValidSuffixes); err != nil {
		return err
	}
	if err = loadJSONFile(dataDir+"hmValidPrefixes.json", &f.hmValidPrefixes); err != nil {
		return err
	}
	if err = loadJSONFile(dataDir+"hmValidSuffixesSegmented.json", &f.hmValidSuffixesSegmented); err != nil {
		return err
	}
	if err = loadJSONFile(dataDir+"hmValidPrefixesSegmented.json", &f.hmValidPrefixesSegmented); err != nil {
		return err
	}

	// Load double maps
	if err = loadJSONFile(dataDir+"hmTemplateCount.json", &f.hmTemplateCount); err != nil {
		return err
	}
	if err = loadJSONFile(dataDir+"wordCount.json", &f.wordCount); err != nil {
		return err
	}
	if err = loadJSONFile(dataDir+"probPrefixes.json", &f.probPrefixes); err != nil {
		return err
	}
	if err = loadJSONFile(dataDir+"probSuffixes.json", &f.probSuffixes); err != nil {
		return err
	}
	if err = loadJSONFile(dataDir+"probCondPrefixes.json", &f.probCondPrefixes); err != nil {
		return err
	}
	if err = loadJSONFile(dataDir+"probCondSuffixes.json", &f.probCondSuffixes); err != nil {
		return err
	}
	if err = loadJSONFile(dataDir+"seenTemplates.json", &f.seenTemplates); err != nil {
		return err
	}
	if err = loadJSONFile(dataDir+"generalVariables.json", &f.generalVariables); err != nil {
		return err
	}

	// Load list maps
	if err = loadJSONFile(dataDir+"hmPreviouslySeenTokenizations.json", &f.hmPreviouslySeenTokenizations); err != nil {
		return err
	}
	if err = loadJSONFile(dataDir+"hmWordPossibleSplits.json", &f.hmWordPossibleSplits); err != nil {
		return err
	}

	// Load nested maps
	if err = loadJSONFile(dataDir+"probPrefixSuffix.json", &f.probPrefixSuffix); err != nil {
		return err
	}
	if err = loadJSONFile(dataDir+"probSuffixPrefix.json", &f.probSuffixPrefix); err != nil {
		return err
	}

	// Load seen before
	if err = loadJSONFile(dataDir+"SeenBefore.json", &f.HmSeenBefore); err != nil {
		return err
	}

	return nil
}

// ScorePartition scores a prefix;stem;suffix partition
func (f *Farasa) ScorePartition(parts []string) float64 {
	score := 0.0
	prefix := strings.TrimSpace(parts[0])
	suffix := strings.TrimSpace(parts[2])
	stem := strings.TrimSpace(parts[1])

	magicNumbersStr := "1:-0.097825818 2:-0.03893654 3:0.13109569 4:0.18436976 5:0.11448806 6:0.53001714 7:0.21098258 8:-0.17760228 9:0.44223878 10:0.26183113 11:-0.05603376 12:0.055829503 13:-0.17745291 14:0.015865559 15:0.66909122 16:0.16948195 17:0.15397599 18:0.60355717"
	magicParts := regexp.MustCompile(` +`).Split(magicNumbersStr, -1)
	magicNo := make([]float64, len(magicParts))
	for i, m := range magicParts {
		idx := strings.Index(m, ":")
		magicNo[i], _ = strconv.ParseFloat(m[idx+1:], 64)
	}

	// Feature 0: prefix probability
	if v, ok := f.probPrefixes[prefix]; ok {
		score += magicNo[0] * math.Log(v)
	} else {
		score += magicNo[0] * -10
	}

	// Feature 1: suffix probability
	if v, ok := f.probSuffixes[suffix]; ok {
		score += magicNo[1] * math.Log(v)
	} else {
		score += magicNo[1] * -10
	}

	trimmedTemp := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(suffix, "+", ""), ";", ""), ",", "")
	altStem := ""
	if strings.HasPrefix(trimmedTemp, "\u062a") && len([]rune(trimmedTemp)) > 1 {
		altStem = stem + "\u0629"
	}

	// Feature 2: stem word count
	stemWordCount := -10.0
	if v, ok := f.wordCount[stem]; ok {
		stemWordCount = v
	} else if len(altStem) > 1 {
		if v, ok := f.wordCount[altStem]; ok {
			stemWordCount = v
		}
	}
	score += magicNo[2] * stemWordCount

	// Feature 3: prefix-suffix co-occurrence
	if inner, ok := f.probPrefixSuffix[prefix]; ok {
		if v, ok := inner[suffix]; ok {
			score += magicNo[3] * math.Log(v)
		} else {
			score += magicNo[3] * -20
		}
	} else {
		score += magicNo[3] * -20
	}

	// Feature 4: suffix-prefix co-occurrence
	if inner, ok := f.probSuffixPrefix[suffix]; ok {
		if v, ok := inner[prefix]; ok {
			score += magicNo[4] * math.Log(v)
		} else {
			score += magicNo[4] * -20
		}
	} else {
		score += magicNo[4] * -20
	}

	// Feature 5: template fit
	if f.ft.FitTemplate(stem) != "Y" {
		score += magicNo[5] * math.Log(f.generalVariables["hasTemplate"])
	} else {
		score += magicNo[5] * math.Log(1-f.generalVariables["hasTemplate"])
	}

	// Feature 6: in morph list
	_, inMorph := f.hmListMorph[stem]
	if !inMorph && strings.HasSuffix(stem, "\u064a") {
		runes := []rune(stem)
		_, inMorph = f.hmListMorph[string(runes[:len(runes)-1])+"\u0649"]
	}
	if inMorph {
		score += magicNo[6] * math.Log(f.generalVariables["inMorphList"])
	} else {
		score += magicNo[6] * math.Log(1-f.generalVariables["inMorphList"])
	}

	// Feature 7: in gazetteer list
	_, inGaz := f.hmListGaz[stem]
	if !inGaz && strings.HasSuffix(stem, "\u064a") {
		runes := []rune(stem)
		_, inGaz = f.hmListGaz[string(runes[:len(runes)-1])+"\u0649"]
	}
	if inGaz {
		score += magicNo[7] * math.Log(f.generalVariables["inGazList"])
	} else {
		score += magicNo[7] * math.Log(1-f.generalVariables["inGazList"])
	}

	// Feature 8: conditional prefix probability
	if v, ok := f.probCondPrefixes[prefix]; ok {
		score += magicNo[8] * math.Log(v)
	} else {
		score += magicNo[8] * -20
	}

	// Feature 9: conditional suffix probability
	if v, ok := f.probCondSuffixes[suffix]; ok {
		score += magicNo[9] * math.Log(v)
	} else {
		score += magicNo[9] * -20
	}

	// Feature 10: stem + first suffix word count
	stemPlusFirstSuffix := stem
	if len(suffix) > 1 {
		if idx := strings.Index(suffix[1:], "+"); idx > 0 {
			stemPlusFirstSuffix += suffix[1 : idx+1]
		} else {
			stemPlusFirstSuffix += suffix
		}
	} else {
		stemPlusFirstSuffix += suffix
	}
	stemWordCount = -10.0
	if v, ok := f.wordCount[stemPlusFirstSuffix]; ok {
		stemWordCount = v
	} else if strings.HasSuffix(stem, "\u064a") {
		runes := []rune(stem)
		alt := string(runes[:len(runes)-1]) + "\u0649"
		if v, ok := f.wordCount[alt]; ok {
			stemWordCount = v
		}
	}
	if stemWordCount == -10 && strings.HasSuffix(stemPlusFirstSuffix, "\u062a") {
		runes := []rune(stemPlusFirstSuffix)
		alt := string(runes[:len(runes)-1]) + "\u0629"
		if v, ok := f.wordCount[alt]; ok {
			stemWordCount = v
		}
	}
	score += magicNo[10] * stemWordCount

	// Feature 11: template count
	template := f.ft.FitTemplate(stem)
	if v, ok := f.hmTemplateCount[template]; ok {
		score += magicNo[11] * math.Log(v)
	} else {
		score += magicNo[11] * -10
	}

	// Feature 12: difference from average stem length
	score += magicNo[12] * math.Log(math.Abs(float64(len([]rune(stem)))-f.generalVariables["averageStemLength"]))

	// Feature 13: AraLexCom
	trimmedTemp = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(suffix, "+", ""), ";", ""), ",", "")
	altStem = ""
	if strings.HasPrefix(trimmedTemp, "\u062a") && len([]rune(trimmedTemp)) > 1 {
		altStem = stem + "\u0629"
	}
	if _, ok := f.hmAraLexCom[stem]; ok {
		if v, ok := f.wordCount[stem]; ok {
			score += magicNo[13] * v
		} else {
			score += magicNo[13] * -10
		}
	} else if strings.HasSuffix(stem, "\u064a") {
		runes := []rune(stem)
		alt := string(runes[:len(runes)-1]) + "\u0649"
		if _, ok := f.hmAraLexCom[alt]; ok {
			if v, ok := f.wordCount[alt]; ok {
				score += magicNo[13] * v
			} else {
				score += magicNo[13] * -10
			}
		} else if len(strings.TrimSpace(altStem)) > 0 {
			if _, ok := f.hmAraLexCom[altStem]; ok {
				if v, ok := f.wordCount[altStem]; ok {
					score += magicNo[13] * v
				} else {
					score += magicNo[13] * -10
				}
			} else {
				score += magicNo[13] * -20
			}
		} else {
			score += magicNo[13] * -20
		}
	} else if len(strings.TrimSpace(altStem)) > 0 {
		if _, ok := f.hmAraLexCom[altStem]; ok {
			if v, ok := f.wordCount[altStem]; ok {
				score += magicNo[13] * v
			} else {
				score += magicNo[13] * -10
			}
		} else {
			score += magicNo[13] * -20
		}
	} else {
		score += magicNo[13] * -20
	}

	// Feature 14: in Buck list
	_, inBuck := f.hmBuck[stem]
	if !inBuck && strings.HasSuffix(stem, "\u064a") {
		runes := []rune(stem)
		_, inBuck = f.hmBuck[string(runes[:len(runes)-1])+"\u0649"]
	}
	if inBuck {
		score += magicNo[14]
	} else {
		score += -1 * magicNo[14]
	}

	// Feature 15: in locations list
	if _, ok := f.hmLocations[stem]; ok {
		score += magicNo[15]
	} else {
		score += -1 * magicNo[15]
	}

	// Feature 16: in people list
	if _, ok := f.hmPeople[stem]; ok {
		score += magicNo[16]
	} else {
		score += -1 * magicNo[16]
	}

	// Feature 17: in stop words list
	_, inStop := f.hmStop[stem]
	if !inStop && strings.HasSuffix(stem, "\u064a") {
		runes := []rune(stem)
		_, inStop = f.hmStop[string(runes[:len(runes)-1])+"\u0649"]
	}
	if inStop {
		score += magicNo[17]
	} else {
		score += -1 * magicNo[17]
	}

	return score
}

// ScoredPartition holds a score and its corresponding partition string
type ScoredPartition struct {
	score     float64
	partition string
}

// GetPartition returns the partition string
func (sp ScoredPartition) GetPartition() string {
	return sp.partition
}

// GetScore returns the score
func (sp ScoredPartition) GetScore() float64 {
	return sp.score
}

// MostLikelyPartition returns the top N segmentations for a word
func (f *Farasa) MostLikelyPartition(word string, numberOfSolutions int) []ScoredPartition {
	word = strings.TrimSpace(word)
	possiblePartitions := f.GetAllPossiblePartitionsOfString(word)

	if strings.HasPrefix(word, "\u0644\u0644") {
		possiblePartitions = append(possiblePartitions, f.GetAllPossiblePartitionsOfString("\u0644\u0627\u0644"+word[len("\u0644\u0644"):])...)
	} else if strings.HasPrefix(word, "\u0648\u0644\u0644") {
		possiblePartitions = append(possiblePartitions, f.GetAllPossiblePartitionsOfString("\u0648\u0644\u0627\u0644"+word[len("\u0648\u0644\u0644"):])...)
	} else if strings.HasPrefix(word, "\u0641\u0644\u0644") {
		possiblePartitions = append(possiblePartitions, f.GetAllPossiblePartitionsOfString("\u0641\u0644\u0627\u0644"+word[len("\u0641\u0644\u0644"):])...)
	}

	var scores []ScoredPartition

	cleanWord := strings.ReplaceAll(word, "+", "")
	if tokenizations, ok := f.hmPreviouslySeenTokenizations[cleanWord]; ok {
		for _, p := range tokenizations {
			pp := f.GetProperSegmentation(strings.ReplaceAll(p, ";", ""))
			parts := strings.Split(" "+pp+" ", ";")
			if len(parts) == 3 {
				sc := f.ScorePartition(parts)
				scores = append(scores, ScoredPartition{sc, pp})
			}
		}
	} else {
		for _, p := range possiblePartitions {
			pp := f.GetProperSegmentation(strings.ReplaceAll(p, ";", ""))
			parts := strings.Split(" "+pp+" ", ";")
			if len(parts) == 3 {
				sc := f.ScorePartition(parts)
				scores = append(scores, ScoredPartition{sc, pp})
			}
		}
	}

	// Sort by score ascending (Java TreeMap order)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score < scores[j].score
	})

	// Keep top N (last N in ascending order)
	if len(scores) > numberOfSolutions {
		scores = scores[len(scores)-numberOfSolutions:]
	}
	return scores
}

// GetAllPossiblePartitionsOfString generates all possible partitions of a string
func (f *Farasa) GetAllPossiblePartitionsOfString(s string) []string {
	var output []string
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return output
	}

	runes := []rune(s)
	fullPartition := string(runes[0:1])
	for i := 1; i < len(runes); i++ {
		fullPartition += "," + string(runes[i:i+1])
	}

	correctFull := f.GetProperSegmentation(strings.ReplaceAll(regexp.MustCompile(`\++`).ReplaceAllString(strings.ReplaceAll(fullPartition, ",", "+"), "+"), "", ""))
	// Hmm, let me redo this more carefully matching the Java
	cleaned := strings.ReplaceAll(fullPartition, ",", "+")
	cleaned = regexp.MustCompile(`\++`).ReplaceAllString(cleaned, "+")
	correctFull = f.GetProperSegmentation(cleaned)

	parts := strings.Split(" "+correctFull+" ", ";")
	if !containsSlice(output, correctFull) {
		if (len(parts) >= 2 && len(strings.TrimSpace(parts[1])) != 1) || len(runes) == 1 {
			output = append(output, correctFull)
		}
	}

	if strings.Contains(fullPartition, ",") {
		output = f.getSubPartitions(fullPartition, output)
	}
	return output
}

func (f *Farasa) getSubPartitions(s string, output []string) []string {
	if !strings.Contains(s, ",") {
		return output
	}
	parts := strings.Split(s, ",")
	for i := 0; i < len(parts)-1; i++ {
		var ss string
		for j := 0; j < i; j++ {
			if j == 0 {
				ss = parts[j]
			} else {
				ss += "," + parts[j]
			}
		}
		if i == 0 {
			ss = parts[i] + parts[i+1]
		} else {
			ss += "," + parts[i] + parts[i+1]
		}
		for k := i + 2; k < len(parts); k++ {
			if k == 0 {
				ss = parts[k]
			} else {
				ss += "," + parts[k]
			}
		}

		cleaned := strings.ReplaceAll(ss, ",", "+")
		cleaned = regexp.MustCompile(`\++`).ReplaceAllString(cleaned, "+")
		proper := f.GetProperSegmentation(cleaned)
		if !containsSlice(output, proper) {
			output = append(output, proper)
			if strings.Contains(ss, ",") {
				output = f.getSubPartitions(ss, output)
			}
		}
	}
	return output
}

func containsSlice(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// checkIfLeadingLettersCouldBePrefixes checks if head could be valid Arabic prefixes
func checkIfLeadingLettersCouldBePrefixes(head string) bool {
	matched, _ := regexp.MatchString(`^(\u0648|\u0641)?(\u0628|\u0643|\u0644)?(\u0627\u0644)?$`, head)
	return matched || head == "\u0633" || head == "\u0648\u0633" || head == "\u0641\u0633"
}

// getPrefixSplit splits a prefix string into its components
func getPrefixSplit(head string) string {
	output := ""
	runes := []rune(head)
	if len(runes) > 0 && (string(runes[0]) == "\u0648" || string(runes[0]) == "\u0641") {
		output += string(runes[0]) + ","
		runes = runes[1:]
	}
	if len(runes) > 0 && (string(runes[0]) == "\u0628" || string(runes[0]) == "\u0643" || string(runes[0]) == "\u0644" || string(runes[0]) == "\u0633") {
		output += string(runes[0]) + ","
		runes = runes[1:]
	}
	if len(runes) >= 2 && string(runes[0:2]) == "\u0627\u0644" {
		output += string(runes[0:2]) + ","
	}
	output = strings.TrimSuffix(output, ",")
	return output
}

// GetProperSegmentation converts a raw partition into prefix;stem;suffix format
func (f *Farasa) GetProperSegmentation(input string) string {
	if len(f.hPrefixes) == 0 {
		for _, p := range Prefixes {
			f.hPrefixes[p] = 1
		}
	}
	if len(f.hSuffixes) == 0 {
		for _, s := range Suffixes {
			f.hSuffixes[s] = 1
		}
	}

	word := strings.Split(input, "+")
	iValidPrefix := -1
	for iValidPrefix+1 < len(word) {
		if _, ok := f.hPrefixes[word[iValidPrefix+1]]; ok {
			iValidPrefix++
		} else {
			break
		}
	}

	iValidSuffix := len(word)
	for iValidSuffix > max(iValidPrefix, 0) {
		w := word[iValidSuffix-1]
		_, isSuffix := f.hSuffixes[w]
		if isSuffix || w == "_" {
			iValidSuffix--
		} else {
			break
		}
	}

	currentPrefix := ""
	for i := 0; i <= iValidPrefix; i++ {
		currentPrefix += word[i] + "+"
	}

	stemPart := ""
	for i := iValidPrefix + 1; i < iValidSuffix; i++ {
		stemPart += word[i]
	}

	if iValidSuffix == iValidPrefix {
		iValidSuffix++
	}

	currentSuffix := ""
	for i := iValidSuffix; i < len(word) && iValidSuffix != iValidPrefix; i++ {
		currentSuffix += "+" + word[i]
	}

	// Handle sin prefix
	if strings.HasSuffix(currentPrefix, "\u0633+") {
		sinMatch, _ := regexp.MatchString(`^[\u064a\u0646\u0623\u062a]`, stemPart)
		if !sinMatch {
			currentPrefix = currentPrefix[:len(currentPrefix)-len("\u0633+")]
			stemPart = "\u0633" + stemPart
		}
	}

	output := currentPrefix + ";" + stemPart + ";" + currentSuffix
	if strings.HasPrefix(output, "+") {
		output = output[1:]
	}
	if strings.HasSuffix(output, "+") {
		output = output[:len(output)-1]
	}
	return strings.ReplaceAll(output, "++", "+")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
