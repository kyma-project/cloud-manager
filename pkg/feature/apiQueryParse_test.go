package feature

import "testing"

// TestQueryConstrainsProvider exercises queryConstrainsProvider directly across all three
// GoFF provider-constraint syntaxes plus the reject cases, so the reverse-equality and
// empty-list branches are armored independently of the fixture-driven lint tests.
func TestQueryConstrainsProvider(t *testing.T) {
	cases := []struct {
		name  string
		query string
		want  bool
	}{
		{"forward equality", `provider == "gcp" and feature in ["nfs"]`, true},
		{"reverse equality", `"gcp" == provider and feature in ["nfs"]`, true},
		{"eq synonym forward", `provider eq "gcp" and feature in ["nfs"]`, true},
		{"EQ synonym uppercase", `provider EQ "gcp"`, true},
		{"in-list single", `provider in ["gcp"] and feature in ["nfs"]`, true},
		{"in-list multi", `provider in ["gcp", "aws"]`, true},
		{"IN uppercase list", `provider IN ["gcp"]`, true},
		{"feature-only is not a provider constraint", `feature in ["nfs", "peering"]`, false},
		{"feature equality is not a provider constraint", `feature == "nfs"`, false},
		{"empty in-list is not a provider constraint", `provider in [] and feature in ["nfs"]`, false},
		{"empty query", ``, false},
		{"brokerPlan only", `brokerPlan == "trial"`, false},
		{"or-composed enable is not bounded", `provider == "gcp" or feature in ["nfs"]`, false},
		{"negated provider is not bounded", `not (provider == "openstack")`, false},
		{"not-equal provider is not bounded", `provider != "openstack"`, false},
		{"ne synonym is not bounded", `provider ne "openstack"`, false},
		{"or inside a quoted value does not trip the scan", `provider == "gcp" and feature == "a or b"`, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := queryConstrainsProvider(tc.query); got != tc.want {
				t.Errorf("queryConstrainsProvider(%q) = %v, want %v", tc.query, got, tc.want)
			}
		})
	}
}
