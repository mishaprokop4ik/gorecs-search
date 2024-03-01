package ranker

import (
	gorecslices "github.com/mishaprokop4ik/gorecs-search/pkg/slices"
	"math"
	"slices"
	"sort"
	"time"
)

// NewModel generates new tf-idf ranking function model.
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

// AddDocuments adds combination of document's path and document's tokens to the ranking model.
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

type DocumentStorer interface {
	Save(path Path, doc Doc) error
	Get(path Path) (Doc, error)
	Remove(path Path) error
}

type RankStorer interface {
	Save(keyWordsHash string, paths []Path) error
	Get(keyWordsHash string) ([]Path, error)
	Remove(keyWordsHash string) error
}

// Model represents API for tf-idf ranking function model.
// Provides results for Model's Docs.
type Model struct {
	Docs Docs
	// TODO: maybe it should be moved to another struct as Model shouldn't think about Storing stuff, it should think only about Ranking...
	DocumentStore DocumentStorer
	RankStore     RankStorer
}

type Docs map[Path]Doc

type Doc struct {
	Terms map[string]uint

	lastModified time.Time
}

type DocFreq map[Path]float64

type Path string

// Rank returns sorted by if-idf rank function paths.
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

// computeTermFrequency calculates tf for a term by a documentPath.
//
// Formula: tf = x / y
//
// Where:
//
// x - the total number of term count in a document;
//
// y - count of all terms by documentPath
func (m *Model) computeTermFrequency(term string, documentPath Path) float64 {
	document := m.Docs[documentPath]
	termFreq := document.Terms[term]
	return float64(termFreq) / float64(len(document.Terms))
}

// computeInverseDocumentFrequency calculates idf for a term.
//
// Formula: idf = log(x / y), if y == 0 => idf = log(x)
//
// Where:
//
// x - total number of documents;
//
// y - number of terms count in all documents.
func (m *Model) computeInverseDocumentFrequency(term string) float64 {
	totalDocNumber := float64(len(m.Docs))
	termAppearsCount := float64(0)

	for _, doc := range m.Docs {
		if termCount := doc.Terms[term]; termCount > 0 {
			termAppearsCount++
		}
	}

	return math.Log10(totalDocNumber / math.Max(termAppearsCount, 1))
}

// computeTFIDF calculates multiplication of computeTermFrequency and computeInverseDocumentFrequency.
func (m *Model) computeTFIDF(term string, documentPath Path) float64 {
	return m.computeTermFrequency(term, documentPath) * m.computeInverseDocumentFrequency(term)
}
