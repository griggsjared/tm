package status

import (
	"testing"
)

type mockTmuxClient struct {
	available bool
	path      string
}

func (m *mockTmuxClient) IsAvailable() bool {
	return m.available
}

func (m *mockTmuxClient) Path() string {
	return m.path
}

type mockFzfRunner struct {
	available bool
	path      string
}

func (m *mockFzfRunner) IsAvailable() bool {
	return m.available
}

func (m *mockFzfRunner) Path() string {
	return m.path
}

func TestStatus_Run(t *testing.T) {
	tests := []struct {
		name      string
		tmuxAvail bool
		tmuxPath  string
		fzfAvail  bool
		fzfPath   string
		wantExit  int
	}{
		{
			name:      "both available",
			tmuxAvail: true,
			tmuxPath:  "/usr/bin/tmux",
			fzfAvail:  true,
			fzfPath:   "/usr/bin/fzf",
			wantExit:  0,
		},
		{
			name:      "tmux available, fzf missing",
			tmuxAvail: true,
			tmuxPath:  "/usr/bin/tmux",
			fzfAvail:  false,
			fzfPath:   "",
			wantExit:  0,
		},
		{
			name:      "tmux missing, fzf available",
			tmuxAvail: false,
			tmuxPath:  "",
			fzfAvail:  true,
			fzfPath:   "/usr/bin/fzf",
			wantExit:  1,
		},
		{
			name:      "both missing",
			tmuxAvail: false,
			tmuxPath:  "",
			fzfAvail:  false,
			fzfPath:   "",
			wantExit:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmuxMock := &mockTmuxClient{available: tt.tmuxAvail, path: tt.tmuxPath}
			fzfMock := &mockFzfRunner{available: tt.fzfAvail, path: tt.fzfPath}
			s := New("test", tmuxMock, fzfMock)

			got := s.Run()
			if got != tt.wantExit {
				t.Errorf("Run() = %v, want %v", got, tt.wantExit)
			}
		})
	}
}

var _ TmuxClient = &mockTmuxClient{}
var _ FzfRunner = &mockFzfRunner{}
