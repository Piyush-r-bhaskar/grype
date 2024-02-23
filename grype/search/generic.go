package search

import (
	"fmt"
	"github.com/anchore/grype/grype/distro"
	"github.com/anchore/grype/grype/match"
	"github.com/anchore/grype/grype/pkg"
	"github.com/anchore/grype/grype/version"
	"github.com/anchore/grype/grype/vulnerability"
)

func GenericPackage(store vulnerability.Provider, d *distro.Distro, p pkg.Package, upstreamMatcher match.MatcherType) ([]match.Match, error) {
	verObj, err := version.NewVersionFromPkg(p)
	if err != nil {
		return nil, fmt.Errorf("matcher failed to parse version pkg=%q ver=%q: %w", p.Name, p.Version, err)
	}

	allPkgVulns, err := store.GetByPURLType(p)
	if err != nil {
		return nil, fmt.Errorf("matcher failed to fetch language=%q pkg=%q: %w", p.Language, p.Name, err)
	}

	applicableVulns, err := onlyQualifiedPackages(d, p, allPkgVulns)
	if err != nil {
		return nil, fmt.Errorf("unable to filter language-related vulnerabilities: %w", err)
	}

	// TODO: Port this over to a qualifier and remove
	applicableVulns, err = onlyVulnerableVersions(verObj, applicableVulns)
	if err != nil {
		return nil, fmt.Errorf("unable to filter language-related vulnerabilities: %w", err)
	}

	var matches []match.Match
	for _, vuln := range applicableVulns {
		matches = append(matches, match.Match{
			Vulnerability: vuln,
			Package:       p,
			Details: []match.Detail{
				{
					Type: match.ExactDirectMatch,
					SearchedBy: map[string]any{
						"namespace": vuln.Namespace,
						"package": map[string]string{
							"name":    p.Name,
							"version": p.Version,
							"purl":    p.PURL,
						},
					},
					Found: map[string]any{
						"vulnerabilityID":   vuln.ID,
						"versionConstraint": vuln.Constraint.String(),
					},
					Matcher:    upstreamMatcher,
					Confidence: 1.0, // TODO: field is hard-coded to full confidence for now
				},
			},
		})
	}
	return matches, err
}
