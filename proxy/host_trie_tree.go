package proxy

import (
	"net"
	"net/http"
	"strings"
	"sync"

	"go.uber.org/zap"
)

type HostTrie struct {
	mu   sync.RWMutex
	root *trieNode
}

type trieNode struct {
	children map[string]*trieNode
	handler  http.Handler
}

func NewHostTrie() *HostTrie {
	return &HostTrie{
		root: &trieNode{
			children: make(map[string]*trieNode),
		},
	}
}

func (h *HostTrie) HasAnyEntry() bool {
	return len(h.root.children) > 0
}

func (h *HostTrie) GetAllHosts() []string {
	domains := make([]string, 0)
	var dfs func(node *trieNode, path []string)
	dfs = func(node *trieNode, path []string) {
		if node.handler != nil {
			domainParts := make([]string, len(path))
			for i := 0; i < len(path); i++ {
				domainParts[i] = path[len(path)-1-i]
			}
			domain := strings.Join(domainParts, ".")
			domains = append(domains, domain)
		}
		for part, child := range node.children {
			dfs(child, append(path, part))
		}
	}
	dfs(h.root, []string{})
	return domains
}

func (h *HostTrie) insert(hostTokens []string, handler http.Handler) {
	h.mu.Lock()
	defer h.mu.Unlock()

	currentRoot := h.root

	for i := 0; i < len(hostTokens); i++ {
		part := hostTokens[i]
		node := currentRoot.children[part]

		if node == nil {
			node = &trieNode{
				children: make(map[string]*trieNode),
			}
			currentRoot.children[part] = node
		}

		currentRoot = node
	}

	currentRoot.handler = handler
}

func (h *HostTrie) Contains(host string) http.Handler {
	tokens := reverseAndTokenizeHost(host)
	h.mu.RLock()
	defer h.mu.RUnlock()

	currentNode := h.root

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		node := currentNode.children[token]
		if node == nil {
			node = currentNode.children["*"]
		}
		if node == nil {
			return nil
		}

		if i == len(tokens)-1 {
			return node.handler
		}

		currentNode = node
	}

	return nil
}

func (h *HostTrie) InsertHost(host string, handler http.Handler) {
	zap.S().Infof("Inserting host %q into HostTrie", host)
	tokens := reverseAndTokenizeHost(host)
	h.insert(tokens, handler)
}

func reverseAndTokenizeHost(host string) []string {
	zap.S().Debugf("Reversing host %s", host)
	host2, _, err := net.SplitHostPort(host)
	if err == nil {
		host = host2
	}
	trimmedHost := strings.TrimSuffix(host, ".")
	lowerAndTrimmedHost := strings.ToLower(trimmedHost)
	splitedHost := strings.Split(lowerAndTrimmedHost, ".")

	for i, j := 0, len(splitedHost)-1; i < j; i, j = i+1, j-1 {
		splitedHost[i], splitedHost[j] = splitedHost[j], splitedHost[i]
	}
	zap.S().Debugf("Reversed and tokenized host: %v", splitedHost)
	return splitedHost
}
