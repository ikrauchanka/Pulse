package pulse

import (
	"fmt"
	"sort"
	"unicode"
)

type outputFunc func(string)

type variation struct {
	text       string
	numMatches int64
}

type token struct {
	word       string
	variable   bool
	required   bool
	variations []variation
}

type pattern struct {
	tokens    []token
	frequency float64
}

type vertex struct {
	x                      int
	y                      int
	startsSequenceOfLength int
}

type vertexDistance struct {
	distance int
	index    int
}

type distArray []vertexDistance

//Channel to receive log data from consuming application
var input <-chan string
var report outputFunc
var patternCreationRate float64
var unmatched []string
var patterns []pattern
var tokenMap [2048][]string

func (s distArray) Len() int           { return len(s) }
func (s distArray) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s distArray) Less(i, j int) bool { return s[i].distance < s[j].distance }

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func getTokens(value string) []string {
	var buffer []rune
	var result []string
	chars := []rune(value)
	for i, r := range chars {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && !unicode.IsDigit(r) && !unicode.IsSpace(r) {
			if len(buffer) > 0 {
				result = append(result, string(buffer))
				buffer = nil
			}
			result = append(result, string(r))
		} else if unicode.IsSpace(r) {
			if len(buffer) > 0 {
				result = append(result, string(buffer))
			}
			buffer = nil
		} else {
			buffer = append(buffer, r)
			if i == len(chars)-1 {
				result = append(result, string(buffer))
			}
		}
	}
	return result
}

func addUpdateVertex(newValue vertex, list []vertex) []vertex {
	var done = false
	for i := range list {
		if newValue.x == list[i].x && newValue.y == list[i].y {
			list[i].startsSequenceOfLength = newValue.startsSequenceOfLength
			done = true
			break
		}
	}

	if !done {
		list = append(list, newValue)
	}

	return list
}

func getNextVertex(value vertex, vertices []vertex) (bool, vertex) {
	x := value.x
	y := value.y

	var distances []vertexDistance
	nextVertexExists := false

	for i := range vertices {
		v := vertices[i]
		if v.x < x || v.y < y {
			continue
		}

		nextVertexExists = true
		distances = append(distances, vertexDistance{(v.x - x) + (v.y - y), i})
	}

	if !nextVertexExists {
		return false, vertex{0, 0, 0}
	}

	sort.Sort(distArray(distances))

	var minDistance = distances[0]
	var nextVertex = vertices[minDistance.index]
	var nextMin = vertexDistance{0, 0}
	if len(distances) > 1 {
		nextMin = distances[1]
		var difference = nextMin.distance - minDistance.distance
		if difference <= 3 && vertices[nextMin.index].startsSequenceOfLength > nextVertex.startsSequenceOfLength {
			nextVertex = vertices[nextMin.index]
		}
	}

	return true, nextVertex
}

func removeVertexFromList(val vertex, vertices []vertex) []vertex {
	for i := range vertices {
		if vertices[i].x == val.x && vertices[i].y == val.y {
			vertices = append(vertices[:i], vertices[i+1:]...)
			break
		}
	}
	return vertices
}

//returns sorted list of tokens in pattern
func analyzeMatrix(matrix [][]int, vertices []vertex) (bool, []vertex) {
	//start with {0, 0}
	var tokens []vertex
	if matrix[0][0] > 0 {
		tokens = append(tokens, vertices[0])
		vertices = removeVertexFromList(vertices[0], vertices)
	}
	var start = vertex{0, 0, 0}
	var foundNextPoint, nextPoint = getNextVertex(start, vertices)
	for foundNextPoint {
		tokens = append(tokens, nextPoint)
		vertices = removeVertexFromList(nextPoint, vertices)
		foundNextPoint, nextPoint = getNextVertex(nextPoint, vertices)
	}
	return len(tokens) > 0, tokens
}

func findPattern(shortTokens []string, longTokens []string) bool {
	foundPattern := false
	var vertices []vertex
	matrix := make([][]int, len(shortTokens))
	for i := range shortTokens {
		matrix[i] = make([]int, len(longTokens))
		for j := range matrix[i] {
			var matches = 0
			if shortTokens[i] == longTokens[j] {
				matches++
				vertices = addUpdateVertex(vertex{i, j, matches}, vertices)
				var prevRow = j - 1
				var prevCol = i - 1
				for prevRow > 0 && prevCol > 0 {
					if shortTokens[prevCol] == longTokens[prevRow] {
						matches++
						vertices = addUpdateVertex(vertex{prevCol, prevRow, matches}, vertices)
						prevRow--
						prevCol--
					} else {
						break
					}
				}
			}
			matrix[i][j] = matches
		}
	}

	for j := range longTokens {
		fmt.Printf("\n")
		for i := range shortTokens {
			fmt.Printf("%v ", matrix[i][j])
		}
	}

	for i := range vertices {
		var vertex = vertices[i]
		fmt.Printf("%v \n", vertex)
	}

	foundPattern, vertices = analyzeMatrix(matrix, vertices)
	if foundPattern {
		fmt.Println("Found a pattern...")
		var p pattern

		lastPoint := vertex{0, 0, 0}
		for i := range vertices {
			var vertex = vertices[i]
			var distance = (vertex.x - lastPoint.x) + (vertex.y - lastPoint.y)
			if distance <= 1 {
				lastPoint = vertex
				text := shortTokens[lastPoint.x]
				p.tokens = append(p.tokens, token{text, false, true, nil})
			} else {
				xDiff := vertex.x - lastPoint.x
				yDiff := vertex.y - lastPoint.y

				skippedColText := ""
				skippedRowText := ""
				if xDiff > 1 {
					var skipped = shortTokens[lastPoint.x+1 : vertex.x]
					for x := range skipped {
						skippedColText += skipped[x]
					}
					fmt.Println("Skipped col: " + skippedColText)
				}

				if yDiff > 1 {
					var skipped = longTokens[lastPoint.y+1 : vertex.y]
					for y := range skipped {
						skippedRowText += skipped[y]
					}
					fmt.Println("Skipped row: " + skippedRowText)
				}

				var variableText []variation
				if skippedColText != "" {
					variableText = append(variableText, variation{skippedColText, 1})
				}
				if skippedRowText != "" {
					variableText = append(variableText, variation{skippedRowText, 1})
				}
				lastPoint = vertex
				text := shortTokens[lastPoint.x]
				//add wildcard token to sequence
				p.tokens = append(p.tokens, token{"!WILDCARD!", true, len(variableText) > 1, variableText})
				//add static token to sequence
				p.tokens = append(p.tokens, token{text, false, true, nil})
			}
			fmt.Printf("%v \n", vertex)
		}
		fmt.Printf("Pattern: %v \n", p)
	}
	return foundPattern
}

func analyze(line string) {
	index := -1
	maxScore := 0.0
	patternFound := false

	//search for existing pattern using token map

	//if no pattern found, compare to unmatched lines, see if a new pattern can be detected
	for i := range unmatched {
		var compare = unmatched[i]
		var distance = ld(line, compare)
		var maxLength = max(len(line), len(compare))
		var score = float64(maxLength-distance) / float64(maxLength)
		if score > maxScore {
			maxScore = score
			index = i
		}
	}

	if maxScore >= 0.5 {
		report("Looking for pattern...")
		var lineTokens = getTokens(line)
		var unmatchedTokens = getTokens(unmatched[index])
		if len(lineTokens) < len(unmatchedTokens) {
			patternFound = findPattern(lineTokens, unmatchedTokens)
		} else {
			patternFound = findPattern(unmatchedTokens, lineTokens)
		}
	}

	if !patternFound {
		unmatched = append(unmatched, line)
		report("Added line to unmatched")
	}
}

//Copied from http://rosettacode.org/wiki/Levenshtein_distance#Go
func ld(s, t string) int {
	d := make([][]int, len(s)+1)
	for i := range d {
		d[i] = make([]int, len(t)+1)
	}
	for i := range d {
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}
	for j := 1; j <= len(t); j++ {
		for i := 1; i <= len(s); i++ {
			if s[i-1] == t[j-1] {
				d[i][j] = d[i-1][j-1]
			} else {
				min := d[i-1][j]
				if d[i][j-1] < min {
					min = d[i][j-1]
				}
				if d[i-1][j-1] < min {
					min = d[i-1][j-1]
				}
				d[i][j] = min + 1
			}
		}

	}
	return d[len(s)][len(t)]
}

//Run starts the pulse package
func Run(in <-chan string, out outputFunc) {
	input = in
	report = out
	analyze("monkey x [michaeld] Hello World")
	analyze("monkey x y x [bob] Hello World")
	/*go func() {
		for value := range in {
			analyze(value)
			//report(value)
		}
	}()*/
}
