package ranker

import (
	gorecslices "github.com/mishaprokop4ik/gorecs-search/pkg/slices"
	"math"
	"slices"
	"sort"
	"time"
)

func NewModel(docs map[string][]string) *Model {
	modelDocs := make(Docs)
	for path, terms := range docs {
		doc := Doc{
			Terms: map[string]uint{},

			lastModified: time.Now(),
		}

		for _, term := range terms {
			doc.Terms[term] = gorecslices.Count(term, terms)
		}

		modelDocs[Path(path)] = doc
	}

	return &Model{
		Docs: modelDocs,
	}
}

func (m *Model) AddDocuments(d map[string][]string) *Model {
	for path, terms := range d {
		doc := Doc{
			Terms: map[string]uint{},

			lastModified: time.Now(),
		}

		for _, term := range terms {
			doc.Terms[term] = gorecslices.Count(term, terms)
		}

		m.Docs[Path(path)] = doc
	}

	return m
}

type Model struct {
	Docs Docs
}

type Docs map[Path]Doc

type Doc struct {
	Terms map[string]uint

	lastModified time.Time
}

type DocFreq map[Path]float64

type Path string

func (m *Model) Rank(keyWords ...string) []Path {
	docFreq := DocFreq{}

	for path := range m.Docs {
		rank := float64(0)
		for _, term := range keyWords {
			rank += m.computeTFIDF(term, path)
		}
		docFreq[path] = rank
	}

	keys := make([]Path, 0, len(docFreq))
	for key := range docFreq {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return docFreq[keys[i]] > docFreq[keys[j]] })

	keys = slices.DeleteFunc(keys, func(path Path) bool {
		return docFreq[path] <= 0
	})

	return keys
}

// tf - x / y
// where x - the total number of term count in a document
// y - the count of all terms
func (m *Model) computeTF(term string, documentPath Path) float64 {
	document := m.Docs[documentPath]
	termFreq := document.Terms[term]
	return float64(termFreq) / float64(len(document.Terms))
}

// idf - log(x / 1 + y)
// x - total number of documents
// y - number of terms count in all documents
func (m *Model) computeIDF(term string) float64 {
	totalDocNumber := float64(len(m.Docs))
	termAppearsCount := float64(0)

	for _, doc := range m.Docs {
		if termCount := doc.Terms[term]; termCount > 0 {
			termAppearsCount++
		}
	}

	return math.Log10(totalDocNumber / math.Max(termAppearsCount, 1))
}

// tfidf = tf * idf
func (m *Model) computeTFIDF(term string, documentPath Path) float64 {
	return m.computeTF(term, documentPath) * m.computeIDF(term)
}
