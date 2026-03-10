package ownership

import "testing"

func TestNormalizeGitOwnerID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "plain email",
			in:   "Alice@example.com",
			want: "alice",
		},
		{
			name: "email plus alias",
			in:   "alice+ci@example.com",
			want: "alice",
		},
		{
			name: "github noreply numeric prefix",
			in:   "126325182+stuckvgn@users.noreply.github.com",
			want: "stuckvgn",
		},
		{
			name: "github noreply username",
			in:   "octocat@users.noreply.github.com",
			want: "octocat",
		},
		{
			name: "raw username",
			in:   "  @Team-Platform  ",
			want: "team-platform",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := normalizeGitOwnerID(tc.in)
			if got != tc.want {
				t.Fatalf("normalizeGitOwnerID(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
