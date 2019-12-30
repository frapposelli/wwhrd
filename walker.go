package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/emicklei/dot"
	log "github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-license-detector.v3/licensedb"
)

var overrides = map[string]string{}

// dependencies are tracked as a graph, but the graph itself is not used to build the nodelist
type dependencies struct {
	nodes     []*node
	nodesList map[string]bool
	edges     map[node][]*node
	dotGraph  *dot.Graph
	sync.RWMutex
}

type node struct {
	pkg    string
	dir    string
	vendor string
}

func newGraph() *dependencies {
	var g dependencies
	g.nodesList = make(map[string]bool)
	return &g
}

// AddNode adds a node to the graph
func (g *dependencies) addNode(n *node) error {
	log.Debugf("[%s] current nodesList status %+v", n.pkg, g.nodesList)
	// check if Node has been visited, this is done raw by caching it in a global hashtable
	if !g.nodesList[n.pkg] {
		g.Lock()
		g.nodes = append(g.nodes, n)
		g.nodesList[n.pkg] = true
		g.Unlock()
		return nil
	}
	return fmt.Errorf("[%s] node already visited", n.pkg)
}

// addEdge adds an edge to the graph
func (g *dependencies) addEdge(n1, n2 *node) {
	g.Lock()
	if g.edges == nil {
		g.edges = make(map[node][]*node)
	}
	g.edges[*n1] = append(g.edges[*n1], n2)
	g.edges[*n2] = append(g.edges[*n2], n1)
	g.Unlock()
}

func (g *dependencies) getDotGraph() string {
	g.dotGraph = dot.NewGraph(dot.Directed)
	g.generateDotGraph(func(n *node) {})
	return g.dotGraph.String()
}

type nodeQueue struct {
	items []node
	sync.RWMutex
}

func (s *nodeQueue) new() *nodeQueue {
	s.Lock()
	s.items = []node{}
	s.Unlock()
	return s
}

func (s *nodeQueue) enqueue(t node) {
	s.Lock()
	s.items = append(s.items, t)
	s.Unlock()
}

func (s *nodeQueue) dequeue() *node {
	s.Lock()
	item := s.items[0]
	s.items = s.items[1:len(s.items)]
	s.Unlock()
	return &item
}

func (s *nodeQueue) isEmpty() bool {
	s.RLock()
	defer s.RUnlock()
	return len(s.items) == 0
}

// do a BFS on the graph and generate dot.Graph
func (g *dependencies) generateDotGraph(f func(*node)) {
	g.RLock()
	q := nodeQueue{}
	q.new()
	n := g.nodes[0]
	q.enqueue(*n)
	visited := make(map[*node]bool)
	for {
		if q.isEmpty() {
			break
		}
		node := q.dequeue()
		// add dotGraph node after dequeing
		dGN := g.dotGraph.Node(node.pkg)

		visited[node] = true
		near := g.edges[*node]

		for i := 0; i < len(near); i++ {
			j := near[i]
			if !visited[j] {
				// add unvisited node to dotGraph
				edGN := g.dotGraph.Node(j.pkg)
				// add an edge in the dotGraph between ancestor and descendant
				g.dotGraph.Edge(dGN, edGN)
				q.enqueue(*j)
				visited[j] = true
			}
		}
		if f != nil {
			f(node)
		}
	}
	g.RUnlock()
}

func (g *dependencies) WalkNode(n *node) {
	var walkFn = func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		log.Debugf("walking %q", path)

		// check if we need to skip this
		if ok, err := shouldSkip(path, info); ok {
			return err
		}

		fs := token.NewFileSet()
		f, err := parser.ParseFile(fs, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}

		for _, s := range f.Imports {
			vendorpkg := strings.Replace(s.Path.Value, "\"", "", -1)
			log.Debugf("found import %q", vendorpkg)
			pkgdir := filepath.Join(n.vendor, "vendor", vendorpkg)
			if _, err := os.Stat(pkgdir); !os.IsNotExist(err) {

				// Add imported pkg to the graph
				vendornode := node{pkg: vendorpkg, dir: pkgdir, vendor: n.vendor}
				log.Debugf("[%s] adding node", vendornode.pkg)
				if err := g.addNode(&vendornode); err != nil {
					log.Debug(err.Error())
					continue
				}
				log.Debugf("[%s] adding node as edge of %s", vendornode.pkg, n.pkg)
				g.addEdge(n, &vendornode)
				log.Debugf("[%s] walking node", vendornode.pkg)
				g.WalkNode(&vendornode)
			}

		}
		return nil
	}

	if err := filepath.Walk(n.dir, walkFn); err != nil {
		return
	}

}

func WalkImports(root string) (map[string]bool, error) {

	graph := newGraph()
	rootNode := node{pkg: "root", dir: root, vendor: root}
	if err := graph.addNode(&rootNode); err != nil {
		log.Debug(err.Error())
	}

	log.Debugf("[%s] walking root node", rootNode.pkg)
	graph.WalkNode(&rootNode)

	return graph.nodesList, nil
}

func GraphImports(root string) (string, error) {

	graph := newGraph()
	rootNode := node{pkg: "root", dir: root, vendor: root}
	if err := graph.addNode(&rootNode); err != nil {
		log.Debug(err.Error())
	}

	log.Debugf("[%s] walking root node", rootNode.pkg)
	graph.WalkNode(&rootNode)

	return graph.getDotGraph(), nil
}

func GetLicenses(root string, list map[string]bool) map[string]string {
	lics := make(map[string]string)

	if !strings.HasSuffix(root, "vendor") {
		root = filepath.Join(root, "vendor")
	}

	for k := range list {
		fpath := filepath.Join(root, k)
		pkg, err := os.Stat(fpath)
		if err != nil {
			continue
		}
		if pkg.IsDir() {

			log.Debugf("[%s] Analyzing %q", k, fpath)
			l := licensedb.Analyse(fpath)
			if l[0].ErrStr != "" {
				var found bool
				// the package might be part of the overrides, let's check for that
				for p, l := range overrides {
					if strings.Contains(k, p) {
						lics[k] = l
						found = true
					}
				}
				// the package might be nested inside a larger package, we try to find
				// the license walking back to the beginning of the path.
				pak := strings.Split(k, "/")
				var path string
				for y := range pak {
					path = filepath.Join(root, strings.Join(pak[:len(pak)-y], "/"))
					log.Debugf("[%s] Analyzing %q", k, path)
					bal := licensedb.Analyse(path)
					if bal[0].ErrStr != "" {
						continue
					} else {
						// We found a license in the leftmost package, that's enough for now
						if len(bal[0].Matches) > 0 {
							lics[k] = bal[0].Matches[0].License
							found = true
						}
						break
					}
				}

				if !found {
					// if our search didn't bear any fruit, ¯\_(ツ)_/¯
					lics[k] = "UNKNOWN"
				}
				continue
			}
			lics[k] = l[0].Matches[0].License
		}
	}

	return lics
}

func shouldSkip(path string, info os.FileInfo) (bool, error) {
	if info.IsDir() {
		name := info.Name()
		// check if directory is in the blocklist
		if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") || name == "testdata" || name == "vendor" {
			log.Debugf("skipping %q: directory in blocklist", path)
			return true, filepath.SkipDir
		}
		return true, nil
	}
	// if it's not a .go file, skip
	if filepath.Ext(path) != ".go" {
		log.Debugf("skipping %q: not a go file", path)
		return true, nil
	}
	// if it's a test file, skip
	if strings.HasSuffix(path, "_test.go") {
		log.Debugf("skipping %q: test file", path)
		return true, nil
	}
	return false, nil
}
