package internal

import "testing"

func TestParseRoleRepoSource(t *testing.T) {
	tests := []struct {
		in   string
		want string
		err  bool
	}{
		{"owner/repo", "owner/repo", false},
		{"https://github.com/owner/repo", "owner/repo", false},
		{"https://github.com/owner/repo.git", "owner/repo", false},
		{"git@github.com:owner/repo.git", "owner/repo", false},
		{"gitlab.com/owner/repo", "", true},
	}

	for _, tt := range tests {
		s, err := ParseRoleRepoSource(tt.in)
		if tt.err {
			if err == nil {
				t.Fatalf("ParseRoleRepoSource(%q) expected error", tt.in)
			}
			continue
		}
		if err != nil {
			t.Fatalf("ParseRoleRepoSource(%q): %v", tt.in, err)
		}
		if got := s.FullName(); got != tt.want {
			t.Fatalf("ParseRoleRepoSource(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestParseRolePathFromYAMLPath(t *testing.T) {
	tests := []struct {
		path           string
		wantRole       string
		wantRolePath   string
		wantAcceptPath bool
	}{
		{"skills/frontend/references/role.yaml", "frontend", "skills/frontend", true},
		{".agents/teams/backend/references/role.yaml", "backend", ".agents/teams/backend", true},
		{"skills/frontend/role.yaml", "", "", false},
		{"agents/teams/backend/references/role.yaml", "", "", false},
	}

	for _, tt := range tests {
		role, rolePath, ok := ParseRolePathFromYAMLPath(tt.path)
		if ok != tt.wantAcceptPath {
			t.Fatalf("ParseRolePathFromYAMLPath(%q) ok=%v, want %v", tt.path, ok, tt.wantAcceptPath)
		}
		if !ok {
			continue
		}
		if role != tt.wantRole || rolePath != tt.wantRolePath {
			t.Fatalf("ParseRolePathFromYAMLPath(%q) = (%q,%q), want (%q,%q)", tt.path, role, rolePath, tt.wantRole, tt.wantRolePath)
		}
	}
}
