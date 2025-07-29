package cosine_similarity

import (
	"math"
	"strings"

	"github.com/tobiashort/orderedmap-go"
)

func CosineSimilarity(s1, s2 string) float64 {
	fields1 := strings.Fields(s1)
	fields2 := strings.Fields(s2)

	union := orderedmap.NewOrderedMap[string, bool]()

	for _, field := range fields1 {
		union.Put(field, true)
	}

	for _, field := range fields2 {
		union.Put(field, true)
	}

	fieldCount1 := orderedmap.NewOrderedMap[string, float64]()
	fieldCount2 := orderedmap.NewOrderedMap[string, float64]()

	for field := range union.Iterate() {
		fieldCount1.Put(field, 0)
		fieldCount2.Put(field, 0)
	}

	for _, field := range fields1 {
		count, _ := fieldCount1.Get(field)
		fieldCount1.Put(field, count+1)
	}

	for _, field := range fields2 {
		count, _ := fieldCount2.Get(field)
		fieldCount2.Put(field, count+1)
	}

	vecA := fieldCount1.Values()
	vecB := fieldCount2.Values()

	var dividend float64 = 0
	for i := 0; i < len(vecA); i++ {
		dividend += vecA[i] * vecB[i]
	}

	var divisorA float64 = 0
	for _, a := range vecA {
		divisorA += a * a
	}
	divisorA = math.Sqrt(divisorA)

	var divisorB float64 = 0
	for _, b := range vecB {
		divisorB += b * b
	}
	divisorB = math.Sqrt(divisorB)

	divisor := divisorA * divisorB

	return dividend / divisor
}
