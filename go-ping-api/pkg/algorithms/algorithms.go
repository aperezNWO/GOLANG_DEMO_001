package algorithms

import (
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func GenerateRandomPoints(vertexSize, sampleSizeRaw, sourcePoint int) string {
	sampleSize := sampleSizeRaw - 2

	graph := make([][]int, vertexSize)
	for i := range graph {
		graph[i] = make([]int, vertexSize)
	}

	currentTime := time.Now().UnixNano() / int64(time.Millisecond)
	randX := rand.New(rand.NewSource(currentTime / 2))
	randY := rand.New(rand.NewSource(currentTime * 2))

	vertexX := fisherYates(sampleSize, randX)
	vertexY := fisherYates(sampleSize, randY)

	vertexArray := make([]string, 0, vertexSize)
	for i := 0; i < vertexSize; i++ {
		separator := ""
		if i < vertexSize-1 {
			separator = "|"
		}
		vertexArray = append(vertexArray, fmt.Sprintf("[%d,%d]%s", vertexX[i], vertexY[i], separator))
	}

	vertexArrayString := strings.Join(vertexArray, "")
	separator2 := "■"
	vertexMatrix := generateRandomMatrix(vertexArray, graph, vertexSize)
	vertexList := dijkstra(vertexArray, graph, vertexSize, sampleSize, sourcePoint)

	sortedListEncoded := strings.ReplaceAll(strings.ReplaceAll(vertexList, ",", "<br/>"), "\t", "&nbsp;")
	return fmt.Sprintf("%s%s%s%s%s", vertexArrayString, separator2, vertexMatrix, separator2, sortedListEncoded)
}

func generateRandomMatrix(vertexString []string, graph [][]int, vertexSize int) string {
	for i := 0; i < vertexSize; i++ {
		graph[i][i] = 0
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano() % 1000))
	for indexX := 0; indexX < vertexSize; indexX++ {
		for indexY := indexX + 1; indexY < vertexSize; indexY++ {
			randomValue := float64(rnd.Intn(2))
			if randomValue == 1.0 {
				randomValue = getHipotemuza(vertexString, indexX, indexY)
			}
			graph[indexX][indexY] = int(randomValue)
			graph[indexY][indexX] = int(randomValue)
		}
	}

	for indexX := 0; indexX < vertexSize; indexX++ {
		zeroCount := 0
		for indexY := 0; indexY < vertexSize; indexY++ {
			if indexX != indexY && graph[indexX][indexY] == 0 {
				zeroCount++
				if zeroCount == vertexSize-1 {
					hipotemuza := int(getHipotemuza(vertexString, indexX, indexY))
					graph[indexX][indexY] = hipotemuza
					graph[indexY][indexX] = hipotemuza
				}
			}
		}
	}

	var sb strings.Builder
	for indexX := 0; indexX < vertexSize; indexX++ {
		separator1 := ""
		if indexX < vertexSize-1 {
			separator1 = "|"
		}
		rowValues := make([]string, vertexSize)
		for indexY := 0; indexY < vertexSize; indexY++ {
			rowValues[indexY] = strconv.Itoa(graph[indexX][indexY])
		}
		sb.WriteString(fmt.Sprintf("{%s}%s", strings.Join(rowValues, ","), separator1))
	}
	return sb.String()
}

func getHipotemuza(vertexString []string, indexX, indexY int) float64 {
	re := regexp.MustCompile(`[|\[\]]`)
	coordSource := strings.Split(re.ReplaceAllString(vertexString[indexY], ""), ",")
	coordDest := strings.Split(re.ReplaceAllString(vertexString[indexX], ""), ",")

	sourceX, _ := strconv.ParseFloat(coordSource[0], 64)
	sourceY, _ := strconv.ParseFloat(coordSource[1], 64)
	destX, _ := strconv.ParseFloat(coordDest[0], 64)
	destY, _ := strconv.ParseFloat(coordDest[1], 64)

	return math.Hypot(math.Abs(destX-sourceX), math.Abs(destY-sourceY))
}

func fisherYates(count int, r *rand.Rand) []int {
	deck := make([]int, count)
	for i := 0; i < count; i++ {
		deck[i] = i + 1
	}

	for i := 0; i <= count-2; i++ {
		j := r.Intn(count - i)
		if j > 0 {
			deck[i], deck[i+j] = deck[i+j], deck[i]
		}
	}

	for i := count - 1; i >= 1; i-- {
		j := r.Intn(i + 1)
		if j != i {
			deck[i], deck[j] = deck[j], deck[i]
		}
	}
	return deck
}

func dijkstra(vertex []string, graph [][]int, vertexSize, sampleSize, sourcePoint int) string {
	dist, path := runDijkstra(graph, sourcePoint, vertexSize)

	var sb strings.Builder
	for index := 0; index < len(dist); index++ {
		if dist[index] >= math.MaxInt32 {
			dist[index] = 0
		}

		separator := ""
		if index < len(dist)-1 {
			separator = ","
		}

		vClean := strings.ReplaceAll(strings.ReplaceAll(vertex[index], ",", ";"), "|", "")
		pClean := strings.ReplaceAll(path[index], ",", ";")

		sb.WriteString(fmt.Sprintf("%02d<%s>-%02d-%s%s", index, vClean, dist[index], pClean, separator))
	}
	return sb.String()
}

func runDijkstra(graph [][]int, src, vertexSize int) ([]int, []string) {
	dist := make([]int, vertexSize)
	previous := make([]int, vertexSize)
	visited := make([]bool, vertexSize)
	path := make([]string, vertexSize)

	for i := 0; i < vertexSize; i++ {
		dist[i] = math.MaxInt32
		previous[i] = -1
	}
	dist[src] = 0

	for i := 0; i < vertexSize; i++ {
		u := -1
		minDist := math.MaxInt32
		for v := 0; v < vertexSize; v++ {
			if !visited[v] && dist[v] < minDist {
				minDist = dist[v]
				u = v
			}
		}

		if u == -1 {
			break
		}
		visited[u] = true

		for v := 0; v < vertexSize; v++ {
			weight := graph[u][v]
			if !visited[v] && weight > 0 && dist[u] != math.MaxInt32 {
				newDist := dist[u] + weight
				if newDist < dist[v] {
					dist[v] = newDist
					previous[v] = u
				}
			}
		}
	}

	for v := 0; v < vertexSize; v++ {
		path[v] = buildPathString(previous, src, v)
	}

	return dist, path
}

func buildPathString(previous []int, src, dest int) string {
	if dest == src {
		return ""
	}

	steps := []int{}
	cur := dest
	for cur != -1 {
		steps = append(steps, cur)
		cur = previous[cur]
	}

	for i, j := 0, len(steps)-1; i < j; i, j = i+1, j-1 {
		steps[i], steps[j] = steps[j], steps[i]
	}

	if len(steps) == 0 || steps[0] != src {
		return ""
	}

	var sb strings.Builder
	for i := 0; i < len(steps)-1; i++ {
		sb.WriteString(fmt.Sprintf("[%d;%d]≡", steps[i], steps[i+1]))
	}
	return sb.String()
}