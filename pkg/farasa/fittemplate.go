package farasa

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// FitTemplateClass handles Arabic root/template morphological pattern matching
type FitTemplateClass struct {
	hmRoot    map[string]float64
	hmTemplate map[string]float64
	templates  map[int][]string
}

// NewFitTemplateClass creates and initializes a FitTemplateClass from data directory
func NewFitTemplateClass(dataDir string) (*FitTemplateClass, error) {
	ft := &FitTemplateClass{
		hmRoot:    make(map[string]float64),
		hmTemplate: make(map[string]float64),
		templates:  make(map[int][]string),
	}
	if err := ft.initVariables(dataDir); err != nil {
		return nil, err
	}
	return ft, nil
}

func (ft *FitTemplateClass) initVariables(dataDir string) error {
	// Load roots
	rootFile, err := os.Open(dataDir + "roots.txt")
	if err != nil {
		return err
	}
	defer rootFile.Close()

	scanner := bufio.NewScanner(rootFile)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")
		if len(parts) == 2 {
			val, err := strconv.ParseFloat(parts[1], 64)
			if err == nil {
				ft.hmRoot[parts[0]] = val
			}
		}
	}

	// Load templates
	templateFile, err := os.Open(dataDir + "template-count.txt")
	if err != nil {
		return err
	}
	defer templateFile.Close()

	scanner = bufio.NewScanner(templateFile)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")
		if len(parts) == 2 {
			tmplLen := len([]rune(parts[0]))
			ft.templates[tmplLen] = append(ft.templates[tmplLen], parts[0])
			if _, exists := ft.hmTemplate[parts[0]]; !exists {
				val, err := strconv.ParseFloat(parts[1], 64)
				if err == nil {
					ft.hmTemplate[parts[0]] = val
				}
			}
		}
	}
	return nil
}

// FitTemplate tries to match a word to a known Arabic morphological template
func (ft *FitTemplateClass) FitTemplate(line string) string {
	tmp := ft.fitStemTemplate(UTF82Buck(line))

	// ends with ta marbouta or yeh
	if strings.Contains(tmp, "Y") && (strings.HasSuffix(line, "\u0629") || strings.HasSuffix(line, "\u064a")) {
		runes := []rune(line)
		tmp = ft.fitStemTemplate(UTF82Buck(string(runes[:len(runes)-1])))
	}
	// ends with ya + ta marbouta
	if strings.Contains(tmp, "Y") && strings.HasSuffix(line, "\u064a\u0629") {
		runes := []rune(line)
		tmp = ft.fitStemTemplate(UTF82Buck(string(runes[:len(runes)-2])))
	}
	// ends with alef maqsoura
	if strings.Contains(tmp, "Y") && strings.HasSuffix(line, "\u0649") {
		runes := []rune(line)
		tmp = ft.fitStemTemplate(UTF82Buck(string(runes[:len(runes)-1]) + "\u064a"))
	}
	// contains any form of alef
	if strings.Contains(tmp, "Y") && (strings.Contains(line, "\u0623") || strings.Contains(line, "\u0622") || strings.Contains(line, "\u0625")) {
		normalized := strings.ReplaceAll(line, "\u0625", "\u0627")
		normalized = strings.ReplaceAll(normalized, "\u0623", "\u0627")
		normalized = strings.ReplaceAll(normalized, "\u0622", "\u0627")
		tmp = ft.fitStemTemplate(UTF82Buck(normalized))
	}
	// double last letter
	if strings.Contains(tmp, "Y") && len([]rune(line)) > 1 {
		runes := []rune(line)
		tmp = ft.fitStemTemplate(UTF82Buck(line + string(runes[len(runes)-1])))
	}
	// starts with "ات"
	if strings.Contains(tmp, "Y") && strings.HasPrefix(line, "\u0627\u062a") {
		runes := []rune(line)
		tmp = ft.fitStemTemplate(UTF82Buck(string(runes[0:1]) + "\u0648" + string(runes[1:])))
	}
	// check for Ta/Dal at position 2
	if strings.Contains(tmp, "Y") && len([]rune(line)) >= 5 {
		runes := []rune(line)
		ch := string(runes[2])
		if ch == "\u0637" || ch == "\u062f" {
			potential := ft.fitStemTemplate(UTF82Buck(string(runes[0:2]) + "\u062a" + string(runes[3:])))
			if len([]rune(potential)) > 3 && string([]rune(potential)[2]) == "t" {
				tmp = potential
			}
		}
	}
	// contains آ (alef madda)
	if strings.Contains(tmp, "Y") && strings.Contains(line, "\u0622") {
		tmp = ft.fitStemTemplate(UTF82Buck(strings.ReplaceAll(line, "\u0622", "\u0623\u0627")))
	}
	// contains ئ or ؤ
	if strings.Contains(tmp, "Y") && (strings.Contains(line, "\u0626") || strings.Contains(line, "\u0624")) {
		replaced := strings.ReplaceAll(line, "\u0626", "\u0621")
		replaced = strings.ReplaceAll(replaced, "\u0624", "\u0621")
		tmp = ft.fitStemTemplate(UTF82Buck(replaced))
	}
	return tmp
}

func (ft *FitTemplateClass) fitStemTemplate(stem string) string {
	stemRunes := []rune(stem)
	stemLen := len(stemRunes)

	templates, exists := ft.templates[stemLen]
	if !exists {
		return "Y"
	}

	if stemLen == 2 {
		root := Buck2Morph(stem + string(stemRunes[1]))
		if _, ok := ft.hmRoot[root]; ok {
			return "fE"
		}
		return "Y"
	}

	var templateResults []string

	for _, s := range templates {
		sRunes := []rune(s)
		root := ""
		lastF := -1
		lastL := -1
		broken := false

		for i := 0; i < len(sRunes) && !broken; i++ {
			ch := string(sRunes[i])
			stemCh := string(stemRunes[i])

			if ch == "f" {
				root += stemCh
			} else if ch == "E" {
				if lastF == -1 {
					root += stemCh
					lastF = i
				} else {
					if stemCh != string(stemRunes[lastF]) {
						broken = true
					}
				}
			} else if ch == "l" {
				if lastL == -1 {
					root += stemCh
					lastL = i
				} else {
					if stemCh != string(stemRunes[lastL]) {
						broken = true
					}
				}
			} else if ch == "C" {
				root += stemCh
			} else {
				if stemCh != ch {
					broken = true
				}
			}
		}

		root = Buck2Morph(root)

		var altRoots []string
		if !broken {
			if _, ok := ft.hmRoot[root]; !ok {
				rootRunes := []rune(root)
				for j := 0; j < len(rootRunes); j++ {
					ch := string(rootRunes[j])
					if ch == "y" || ch == "A" || ch == "w" {
						head := string(rootRunes[:j])
						tail := string(rootRunes[j+1:])
						if _, ok := ft.hmRoot[head+"w"+tail]; ok {
							altRoots = append(altRoots, head+"w"+tail)
						}
						if _, ok := ft.hmRoot[head+"y"+tail]; ok {
							altRoots = append(altRoots, head+"y"+tail)
						}
						if _, ok := ft.hmRoot[head+"A"+tail]; ok {
							altRoots = append(altRoots, head+"A"+tail)
						}
					}
				}
			}
		}

		if !broken {
			if _, ok := ft.hmRoot[root]; ok {
				templateResults = append(templateResults, s+"/"+root)
			}
		}
		for _, alt := range altRoots {
			templateResults = append(templateResults, s+"/"+alt)
		}
	}

	if len(templateResults) == 0 {
		return "Y"
	}

	var withC, withoutC []string
	for _, ss := range templateResults {
		if strings.Contains(ss, "C") {
			withC = append(withC, ss)
		} else {
			withoutC = append(withoutC, ss)
		}
	}

	if len(withoutC) == 0 {
		return ft.getBestTemplate(templateResults)
	}
	return ft.getBestTemplate(withoutC)
}

func (ft *FitTemplateClass) getBestTemplate(templates []string) string {
	bestScore := 0.0
	bestTemplate := ""
	for _, s := range templates {
		parts := strings.SplitN(s, "/", 2)
		if len(parts) == 2 {
			rootScore, rok := ft.hmRoot[parts[1]]
			tmplScore, tok := ft.hmTemplate[parts[0]]
			if rok && tok {
				score := rootScore * tmplScore
				if bestScore < score {
					bestScore = score
					bestTemplate = parts[0]
				}
			}
		}
	}
	return bestTemplate
}
