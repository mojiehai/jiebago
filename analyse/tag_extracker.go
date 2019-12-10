// Package analyse is the Golang implementation of Jieba's analyse module.
package analyse

import (
	"github.com/mojiehai/jiebago/posseg"
	"sort"
	"strings"
	"unicode/utf8"
)

// Segment represents a word with weight.
type Segment struct {
	text   string
	weight float64
}

// Text returns the segment's text.
func (s Segment) Text() string {
	return s.text
}

// Weight returns the segment's weight.
func (s Segment) Weight() float64 {
	return s.weight
}

// Segments represents a slice of Segment.
type Segments []Segment

func (ss Segments) Len() int {
	return len(ss)
}

func (ss Segments) Less(i, j int) bool {
	if ss[i].weight == ss[j].weight {
		return ss[i].text < ss[j].text
	}

	return ss[i].weight < ss[j].weight
}

func (ss Segments) Swap(i, j int) {
	ss[i], ss[j] = ss[j], ss[i]
}

// TagExtracter is used to extract tags from sentence.
type TagExtracter struct {
	seg      *posseg.Segmenter
	idf      *Idf
	stopWord *StopWord
}

// LoadDictionary reads the given filename and create a new dictionary.
func (t *TagExtracter) LoadDictionary(fileName string) error {
	t.stopWord = NewStopWord()
	t.seg = new(posseg.Segmenter)
	return t.seg.LoadDictionary(fileName)
}

// LoadIdf reads the given file and create a new Idf dictionary.
func (t *TagExtracter) LoadIdf(fileName string) error {
	t.idf = NewIdf()
	return t.idf.loadDictionary(fileName)
}

// LoadStopWords reads the given file and create a new StopWord dictionary.
func (t *TagExtracter) LoadStopWords(fileName string) error {
	t.stopWord = NewStopWord()
	return t.stopWord.loadDictionary(fileName)
}


// ExtractTags extracts the topK key words from sentence.
// Parameter allowPOS allows a customized pos list.
func (t *TagExtracter) ExtractTagsWithPOS(sentence string, topK int, allowPOS []string) (tags Segments) {
	freqMap := make(map[string]float64)

	// allowPos值转键
	allowPOSMap := map[string]interface{}{}
	for _, k := range allowPOS {
		allowPOSMap[k] = 1
	}

	for w := range t.seg.Cut(sentence, true) {
		// 词性过滤
		if len(allowPOS) > 0 {
			_, ok := allowPOSMap[w.Pos()]
			if !ok {
				continue
			}
		}

		// 长度<2或者在停止词中过滤
		word := strings.TrimSpace(w.Text())
		if utf8.RuneCountInString(word) < 2 {
			continue
		}

		if t.stopWord.IsStopWord(strings.ToLower(word)) {
			continue
		}
		if f, ok := freqMap[word]; ok {
			freqMap[word] = f + 1.0
		} else {
			freqMap[word] = 1.0
		}
	}
	total := 0.0
	for _, freq := range freqMap {
		total += freq
	}
	for k, v := range freqMap {
		freqMap[k] = v / total
	}
	ws := make(Segments, 0)
	var s Segment
	for k, v := range freqMap {
		if freq, ok := t.idf.Frequency(k); ok {
			s = Segment{text: k, weight: freq * v}
		} else {
			s = Segment{text: k, weight: t.idf.median * v}
		}
		ws = append(ws, s)
	}
	sort.Sort(sort.Reverse(ws))
	if len(ws) > topK {
		tags = ws[:topK]
	} else {
		tags = ws
	}
	return tags
}


// ExtractTags extracts the topK key words from sentence.
func (t *TagExtracter) ExtractTags(sentence string, topK int) (tags Segments) {
	return t.ExtractTagsWithPOS(sentence, topK, defaultAllowPOS)
}
