package main

import (
	"testing"
)

var prepareGitMatches = []struct {
	name        string
	lhs         string
	rhs         string
	shouldMatch bool
}{
	{"identical - https", "https://github.com/cloud66/stacks-test.git", "https://github.com/cloud66/stacks-test.git", true},
	{"identical - http", "http://github.com/cloud66/stacks-test.git", "http://github.com/cloud66/stacks-test.git", true},
	{"identical - git", "git://github.com/cloud66/stacks-test.git", "git://github.com/cloud66/stacks-test.git", true},
	{"no .git - https", "https://github.com/cloud66/stacks-test", "git://github.com/cloud66/stacks-test", true},
	{"no .git - http", "http://github.com/cloud66/stacks-test", "http://github.com/cloud66/stacks-test", true},
	{"asym .git - https", "https://github.com/cloud66/stacks-test.git", "https://github.com/cloud66/stacks-test", true},
	{"asym .git - git", "git://github.com/cloud66/stacks-test.git", "git://github.com/cloud66/stacks-test", true},
	{"identical - git@", "git@github.com:cloud66/stacks-test.git", "git@github.com:cloud66/stacks-test.git", true},
	{"no .git - git@", "git@github.com:cloud66/stacks-test.git", "git@github.com:cloud66/stacks-test", true},
	{"http - git", "git://github.com/cloud66/stacks-test.git", "http://github.com/cloud66/stacks-test.git", true},
	{"http - git@", "git@github.com:cloud66/stacks-test.git", "http://github.com/cloud66/stacks-test.git", true},
	{"git - git@", "git@github.com:cloud66/stacks-test.git", "git://github.com/cloud66/stacks-test.git", true},
	{"different domains - git@ - git", "git@bitbucket.com:cloud66/stacks-test.git", "git://github.com/cloud66/stacks-test.git", false},
	{"different domains - git - git", "git://bitbucket.com/cloud66/stacks-test.git", "git://github.com/cloud66/stacks-test.git", false},
	{"different domains - http - http", "http://bitbucket.com/cloud66/stacks-test.git", "http://github.com/cloud66/stacks-test.git", false},
	{"different domains - https - http", "https://bitbucket.com/cloud66/stacks-test.git", "http://github.com/cloud66/stacks-test.git", false},
	{"invalid URLs", "this is totally wrong", "yes! absolutely", false},
	{"different URLs - http - http", "http://github.com/cloud666/stacks-test.git", "http://github.com/cloud66/stacks-test.git", false},
	{"different URLs - git - http", "git://github.com/cloud666/stacks-test.git", "http://github.com/cloud66/stacks-test.git", false},
	{"different URLs - @git - http", "git@github.com:cloud666/stacks-test.git", "http://github.com/cloud66/stacks-test.git", false},
	{"different URLs - @git - git", "git@github.com:cloud666/stacks-test.git", "git://github.com/cloud66/stacks-test.git", false},
	{"Same URL - different users - https - https", "https://a:b@github.com/cloud66/stacks-test.git", "https://x:y@github.com/cloud66/stacks-test.git", true},
	{"Same URL - different users - git - https", "git://a:b@github.com/cloud66/stacks-test.git", "https://x:y@github.com/cloud66/stacks-test.git", true},
	{"Same URL - different users - git - git@", "git://a:b@github.com/cloud66/stacks-test.git", "git@github.com:cloud66/stacks-test.git", true},
	{"identical - allow spaces - https", "https://github.com/cloud66/stacks-test.git ", " https://github.com/cloud66/stacks-test.git", true},
}

func TestAreSameRemotes(t *testing.T) {
	for _, m := range prepareGitMatches {
		result, err := areSameRemotes(m.lhs, m.rhs)
		if err != nil {
			t.Errorf("Error %v in %s\n", err, m.name)
		}

		if result != m.shouldMatch {
			t.Errorf("%s failed with with %t. Should have been %t\n", m.name, result, m.shouldMatch)
		}
	}
}
