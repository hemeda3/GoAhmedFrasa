# FarasaGo

Go port of QCRI's Farasa Arabic morphological segmenter. Takes Arabic text, splits each word into prefix + stem + suffix.

## What it does

Input: `للتواصل الله يعرفون بالمحكمة`

Output: `ل+ال+تواصل الله يعرف+ون ب+ال+محكم+ه`

Each word is broken into its morphological components separated by `+`. The stem is the core meaning-bearing part.

### Segmentation examples

| Input | Output | Breakdown |
|---|---|---|
| `للتواصل` | `ل+ال+تواصل` | prefix `ل` + prefix `ال` + stem `تواصل` |
| `يعرفون` | `يعرف+ون` | stem `يعرف` + suffix `ون` |
| `بالمحكمة` | `ب+ال+محكم+ه` | prefix `ب` + prefix `ال` + stem `محكم` + suffix `ه` |
| `محمد` | `محمد` | stem only (no affixes) |
| `استقلالها` | `استقلال+ها` | stem `استقلال` + suffix `ها` |
| `والتنمية` | `و+ال+تنمي+ه` | prefix `و` + prefix `ال` + stem `تنمي` + suffix `ه` |
| `فهم` | `فهم` | stem only |
| `له` | `ل+ه` | prefix `ل` + stem `ه` |

## How it works (end-to-end flow)

```
Arabic text input
      |
      v
1. Remove diacritics (تشكيل)
      |
      v
2. Tokenize into words
      |
      v
3. For each word:
   a. Check SeenBefore cache (99K pre-computed segmentations)
      -> if found, return cached result
      -> if not found, continue:
   b. Generate all possible prefix+stem+suffix splits
   c. For each split, score using 18 weighted features:
      - prefix/suffix probability
      - stem word frequency (from 613K word corpus)
      - prefix-suffix co-occurrence
      - morphological template matching (فعل patterns)
      - dictionary lookups (212K stems, 66K gazetteer, 47K Buckwalter)
      - named entity lists (12K people, 12K locations)
      - stop word list
   d. Return the highest-scoring split
      |
      v
4. Output: segmented text with + between morphemes
```

## Data files

26 JSON dictionaries in `data/`. Converted from Java `.ser` serialization files.

| File | Entries | Purpose |
|---|---|---|
| `wordCount.json` | 613,116 | Word frequencies (log probability) |
| `hmListMorph.json` | 212,205 | Known valid Arabic stems |
| `hmListGaz.json` | 66,327 | Gazetteer (known entities) |
| `hmBuck.json` | 47,421 | Buckwalter morphological analyzer entries |
| `hmAraLexCom.json` | 25,936 | Arabic lexicon |
| `hmPeople.json` | 12,476 | Person names |
| `hmLocations.json` | 12,608 | Location names |
| `hmStop.json` | 102 | Stop words |
| `SeenBefore.json` | 99,069 | Pre-computed segmentation cache |
| `hmPreviouslySeenTokenizations.json` | 17,308 | Known correct tokenizations |
| `probPrefixes.json` | - | Prefix probabilities |
| `probSuffixes.json` | - | Suffix probabilities |
| `probCondPrefixes.json` | - | Conditional prefix probabilities |
| `probCondSuffixes.json` | - | Conditional suffix probabilities |
| `probPrefixSuffix.json` | - | Prefix-suffix co-occurrence |
| `probSuffixPrefix.json` | - | Suffix-prefix co-occurrence |
| `hmTemplateCount.json` | - | Morphological template frequencies |
| `generalVariables.json` | - | Model parameters (avg stem length, etc.) |
| `roots.txt` | - | Arabic root list |
| `template-count.txt` | - | Template frequency counts |

## Build

```
go build -o farasa ./cmd/farasa/
```

## Usage

### From stdin

```
echo "مؤتمر الأمم المتحدة للتجارة والتنمية" | ./farasa -d ./data/
```

Output:
```
مءتمر ال+امم ال+متحد+ه ل+ال+تجار+ه و+ال+تنمي+ه
```

### From file

```
./farasa -d ./data/ -i input.txt -o output.txt
```

### ATB scheme (Arabic Treebank segmentation)

```
echo "بالمحكمة" | ./farasa -d ./data/ -c atb
```

### All flags

```
-d    Data directory path (default: ./data/ or $FarasaDataDir env var)
-i    Input file path (default: stdin)
-o    Output file path (default: stdout)
-c    Segmentation scheme. Use "atb" for Arabic Treebank style
-n    Normalization true/false (default: true)
```

## Use as a Go package

```go
package main

import (
    "fmt"
    "strings"
    "farasa/pkg/farasa"
)

func main() {
    // Initialize
    f, err := farasa.NewFarasa("./data/")
    if err != nil {
        panic(err)
    }

    // Segment a word
    word := "للتواصل"
    solutions := f.MostLikelyPartition(farasa.Buck2UTF8(word), 1)
    if len(solutions) > 0 {
        result := solutions[0].GetPartition()
        result = strings.ReplaceAll(strings.ReplaceAll(result, ";", ""), "++", "+")
        fmt.Println(result) // ل+ال+تواصل
    }

    // Utilities
    text := "كِتَابٌ"
    clean := farasa.RemoveDiacritics(text)    // كتاب
    normalized := farasa.NormalizeFull(clean)  // كتاب
    tokens := farasa.Tokenize(clean)           // ["ktAb"] (Buckwalter)
    fmt.Println(clean, normalized, tokens)
}
```

## Source files

```
cmd/farasa/main.go          CLI entry point, stdin/file processing
pkg/farasa/farasa.go        Core segmenter: scoring, partitioning, dictionary lookups
pkg/farasa/arabicutils.go   Arabic text utilities: transliteration, normalization, tokenization
pkg/farasa/fittemplate.go   Morphological template matching (Arabic root/pattern system)
data/                       26 JSON dictionary files
```

### Core functions

**farasa.go:**
- `NewFarasa(dataDir)` — load all dictionaries, return segmenter instance
- `MostLikelyPartition(word, n)` — return top N segmentations for a word
- `ScorePartition(parts)` — score a prefix;stem;suffix split using 18 features
- `GetAllPossiblePartitionsOfString(s)` — generate all valid splits
- `GetProperSegmentation(input)` — convert raw split to prefix;stem;suffix format

**arabicutils.go:**
- `RemoveDiacritics(s)` — strip Arabic diacritics
- `NormalizeFull(s)` — normalize alef/taa marbuta/alef maqsura
- `Tokenize(s)` — split text into Buckwalter-encoded tokens
- `Buck2UTF8(s)` / `UTF82Buck(s)` — Buckwalter transliteration

**fittemplate.go:**
- `FitTemplate(word)` — match word to Arabic morphological template (e.g. فعل, فاعل, مفعول)

## Test results

Verified against original Java implementation. 100% match on all test cases.

### testfile.txt (39 lines, 622 words)

```
diff <(java output | normalize whitespace) <(go output | normalize whitespace)
EXIT CODE: 0
```

### StemExtractor test words (17 words)

```
فهم فك بلي والي للتواصل الله تللا فلل بنات زيت بيتي يعرفون زيتون يد أب له كي
```

| Word | Java | Go | Match |
|---|---|---|---|
| فهم | فهم | فهم | yes |
| فك | فك | فك | yes |
| بلي | بلي | بلي | yes |
| والي | والي | والي | yes |
| للتواصل | ل+ال+تواصل | ل+ال+تواصل | yes |
| الله | الله | الله | yes |
| تللا | تلل+ا | تلل+ا | yes |
| فلل | فلل | فلل | yes |
| بنات | بن+ات | بن+ات | yes |
| زيت | زيت | زيت | yes |
| بيتي | بيتي | بيتي | yes |
| يعرفون | يعرف+ون | يعرف+ون | yes |
| زيتون | زيتون | زيتون | yes |
| يد | يد | يد | yes |
| أب | اب | اب | yes |
| له | ل+ه | ل+ه | yes |
| كي | كي | كي | yes |

### DeserializeViewer test words (4 words)

| Word | Java | Go | Match |
|---|---|---|---|
| محمد | محمد | محمد | yes |
| رايان | راي+ان | راي+ان | yes |
| كتاب | كتاب | كتاب | yes |
| لاعب | لاعب | لاعب | yes |

Total: 21/21 words match, 622/622 words match on full document test. Zero differences.

## Origin

Ported from [QCRI Farasa](http://alt.qcri.org/tools/farasa/) Java implementation. The 18 scoring weights were trained on the Arabic Treebank (ATB) corpus. Dictionary data comes from multiple Arabic NLP resources (morphological analyzers, gazetteers, Buckwalter).

## Differences from Java version

- Data format: JSON files instead of Java `.ser` serialization
- No external dependencies (Java version needs commons-lang3, mapdb)
- Single static binary after compilation
- Same algorithm, same weights, same output
