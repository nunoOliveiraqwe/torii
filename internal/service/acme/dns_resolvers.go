package acme

import "strings"

// NormalizeDNSResolvers trims empty entries and removes duplicates while
// preserving the first spelling supplied by the user.
func NormalizeDNSResolvers(resolvers []string) []string {
	normalized := make([]string, 0, len(resolvers))
	seen := make(map[string]struct{}, len(resolvers))

	for _, resolver := range resolvers {
		r := strings.TrimSpace(resolver)
		if r == "" {
			continue
		}
		key := strings.ToLower(r)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, r)
	}

	return normalized
}
