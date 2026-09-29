package kion

import (
	"regexp"
	"testing"
)

// Every sub-package has a recorded release, and a v3_NN release is either a
// Kion release tag inside that minor (3.NN.P) or its unreleased support branch
// (support-3.NN.x). Anything else means stamp-kion-releases wrote a value that
// KionRelease's doc comment does not describe.
func TestKionReleasesMatchKionVersions(t *testing.T) {
	for sub, minor := range KionVersions {
		rel, ok := KionRelease(sub)
		if !ok {
			t.Errorf("KionRelease(%q) missing — spec/%s/kion-release not stamped?", sub, sub)
			continue
		}
		if sub == "master" {
			continue
		}
		tag := regexp.MustCompile(`^` + regexp.QuoteMeta(minor) + `\.[0-9]+$`)
		if !tag.MatchString(rel) && rel != "support-"+minor+".x" {
			t.Errorf("KionRelease(%q) = %q, want %s.<patch> or support-%s.x", sub, rel, minor, minor)
		}
	}
	for sub := range KionReleases {
		if _, ok := KionVersions[sub]; !ok {
			t.Errorf("KionReleases has %q, which KionVersions does not", sub)
		}
	}
}
