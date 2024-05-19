package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Ant struct {
	id       int
	position string
	path     []string
	step     int
}

type Path struct {
	nodes []string
}

type Graph struct {
	V       int
	adj     map[string][]string
	paths   []Path
	ants    []Ant
	start   string
	end     string
	traffic map[string]bool // Tracks traffic on paths
}

func newGraph() *Graph {
	return &Graph{
		V:       0,
		adj:     make(map[string][]string),
		paths:   make([]Path, 0),
		ants:    make([]Ant, 0),
		start:   "",
		end:     "",
		traffic: make(map[string]bool),
	}
}

func (g *Graph) addEdge(v, w string) {
	g.adj[v] = append(g.adj[v], w)
	g.adj[w] = append(g.adj[w], v) // Add reverse direction
}

func (g *Graph) addPath(nodes []string) {
	g.paths = append(g.paths, Path{nodes: nodes})
}

func (g *Graph) addAnts(count int) {
	for i := 1; i <= count; i++ {
		g.ants = append(g.ants, Ant{id: i, position: g.start})
	}
}

func (g *Graph) parseInput(textfile string) (int, error) {
	content, err := ioutil.ReadFile("examples/" + textfile)
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(content), "\n")
	antCount, err := strconv.Atoi(lines[0])
	if antCount == 0 {
		return 0, fmt.Errorf("Error: İnvalid data format")
	}
	if err != nil {
		return 0, fmt.Errorf("invalid ant count")
	}
	startFound := false
	endFound := false

	for i, line := range lines {
		if strings.TrimSpace(line) == "##start" {
			startFound = true
			startLine := strings.Fields(lines[i+1])
			if len(startLine) < 3 {
				return 0, fmt.Errorf("invalid data format")
			}
			g.start = startLine[0]
		} else if strings.TrimSpace(line) == "##end" {
			endFound = true
			endLine := strings.Fields(lines[i+1])
			if len(endLine) < 3 {
				return 0, fmt.Errorf("invalid data format")
			}
			g.end = endLine[0]
		} else if strings.Contains(line, "-") {
			nodes := strings.Split(line, "-")
			if nodes[0] == nodes[1] {
				return 0, fmt.Errorf("ERROR: invalid data format")
			}
			if len(nodes) != 2 {
				return 0, fmt.Errorf("invalid data format")
			}
			g.addEdge(strings.TrimSpace(nodes[0]), strings.TrimSpace(nodes[1]))
		}
	}

	if !startFound || !endFound || g.start == "" || g.end == "" {
		return 0, fmt.Errorf("start or end node not found or invalid data format")
	}

	return antCount, nil
}

func (g *Graph) findAllPaths(start, end string, visited map[string]bool, path []string) {
	visited[start] = true
	path = append(path, start)

	if start == end {
		g.addPath(append([]string(nil), path...)) // Path slice'ını kopyalayarak kaydet
	} else {
		for _, i := range g.adj[start] {
			if !visited[i] {
				g.findAllPaths(i, end, visited, path)
			}
		}
	}

	visited[start] = false
}

func (p1 Path) Equals(p2 Path) bool {
	if len(p1.nodes) != len(p2.nodes) {
		return false
	}
	for i := range p1.nodes {
		if p1.nodes[i] != p2.nodes[i] {
			return false
		}
	}
	return true
}

func (g *Graph) hasCommonNode(nodes1, nodes2 []string) bool {
	set := make(map[string]struct{})
	for _, n := range nodes1 {
		set[n] = struct{}{}
	}
	for _, n := range nodes2 {
		if _, ok := set[n]; ok {
			return true
		}
	}
	return false
}

func (g *Graph) filterPaths() []Path {
	if len(g.paths) == 0 {
		return nil
	}

	// İlk en kısa yolu ekle
	validPaths := []Path{g.paths[0]}

	// İlk yolu aldıktan sonra, çakışmayan yolları ekle
	for _, path := range g.paths[1:] {
		// Mevcut geçerli yollarla çakışmayan bir yol bul
		valid := true
		for _, vp := range validPaths {
			if g.hasCommonNode(vp.nodes[1:len(vp.nodes)-1], path.nodes[1:len(path.nodes)-1]) {
				valid = false
				break
			}
		}

		if valid {
			validPaths = append(validPaths, path)
		}
	}

	return validPaths
}

type ByLength []Path

func (a ByLength) Len() int           { return len(a) }
func (a ByLength) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByLength) Less(i, j int) bool { return len(a[i].nodes) < len(a[j].nodes) }

func (g *Graph) moveAnts(validPaths []Path) {
	for {
		moved := false
		var output []string

		for i := range g.ants {
			ant := &g.ants[i]
			if ant.position == g.end {
				continue
			}

			if ant.position == g.start {
				for _, path := range validPaths {
					nextPos := path.nodes[1]
					if !g.traffic[nextPos] {
						ant.path = path.nodes
						ant.step = 1
						ant.position = nextPos
						g.traffic[nextPos] = true
						moved = true
						output = append(output, fmt.Sprintf("L%d-%s", ant.id, ant.position))
						break
					}
				}
			} else if ant.step < len(ant.path)-1 {
				nextPos := ant.path[ant.step+1]
				if !g.traffic[nextPos] {
					ant.step++
					ant.position = nextPos
					if ant.position == g.end {
						// Karınca end noktasına ulaştıysa, hareket etmeyi bitir
						output = append(output, fmt.Sprintf("L%d-%s", ant.id, "end"))
					} else {
						g.traffic[nextPos] = true
						output = append(output, fmt.Sprintf("L%d-%s", ant.id, ant.position))
					}
					moved = true
				} else {
					// Karınca hareket edemediği için mevcut pozisyonunu ekleyelim
					output = append(output, fmt.Sprintf("L%d-%s", ant.id, ant.position))
				}
			} else {
				// Karınca daha fazla hareket edemiyor
				output = append(output, fmt.Sprintf("L%d-%s", ant.id, ant.position))
			}
		}

		if len(output) > 0 {
			fmt.Println(strings.Join(output, " "))
		}

		// Reset the traffic map for the next turn
		g.traffic = make(map[string]bool)

		// Stop the loop if no ant moved
		if !moved {
			break
		}
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Kullanım: go run . <dosya_adı>")
		os.Exit(1)
	}

	textfile := os.Args[1]

	g := newGraph()
	antCount, err := g.parseInput(textfile)
	if err != nil {
		log.Fatal(err)
	}

	visited := make(map[string]bool)
	g.findAllPaths(g.start, g.end, visited, []string{})

	sort.Sort(ByLength(g.paths))

	// Geçerli yolları filtrele
	validPaths := g.filterPaths()

	// Karıncaları ekle
	g.addAnts(antCount)
	// Karıncaları validPaths yollarında hareket ettir
	g.moveAnts(validPaths)
}
