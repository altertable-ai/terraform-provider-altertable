package provider

import "testing"

func TestParseEnvScopedImportID(t *testing.T) {
	tests := []struct {
		name            string
		in              string
		wantEnv, wantID string
		wantOK          bool
	}{
		{"valid", "env_1:db_2", "env_1", "db_2", true},
		{"colon in id is preserved", "env_1:db:2", "env_1", "db:2", true},
		{"missing colon", "env_1", "", "", false},
		{"empty env", ":db_2", "", "", false},
		{"empty id", "env_1:", "", "", false},
		{"empty", "", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env, id, ok := parseEnvScopedImportID(tt.in)
			if ok != tt.wantOK || env != tt.wantEnv || id != tt.wantID {
				t.Fatalf("parseEnvScopedImportID(%q) = (%q,%q,%v), want (%q,%q,%v)",
					tt.in, env, id, ok, tt.wantEnv, tt.wantID, tt.wantOK)
			}
		})
	}
}

func TestParseCredentialImportID(t *testing.T) {
	tests := []struct {
		name                             string
		in                               string
		wantPT, wantPID, wantEnv, wantID string
		wantOK                           bool
	}{
		{"valid", "service_account:sa_1:env_2:cred_3", "service_account", "sa_1", "env_2", "cred_3", true},
		{"too few", "sa_1:env_2:cred_3", "", "", "", "", false},
		{"too many", "a:b:c:d:e", "", "", "", "", false},
		{"empty segment", "service_account::env_2:cred_3", "", "", "", "", false},
		{"empty", "", "", "", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt, pid, env, id, ok := parseCredentialImportID(tt.in)
			if ok != tt.wantOK || pt != tt.wantPT || pid != tt.wantPID || env != tt.wantEnv || id != tt.wantID {
				t.Fatalf("parseCredentialImportID(%q) = (%q,%q,%q,%q,%v), want (%q,%q,%q,%q,%v)",
					tt.in, pt, pid, env, id, ok, tt.wantPT, tt.wantPID, tt.wantEnv, tt.wantID, tt.wantOK)
			}
		})
	}
}
