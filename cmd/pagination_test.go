package cmd

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/windosx/zentao-cli/pkg/zentao"
)

func TestPagination_FlagsAndApply(t *testing.T) {
	tests := []struct {
		name          string
		pageArg       string
		limitArg      string
		expectErr     bool
		expectedErrIs error
		checkParams   func(t *testing.T, p zentao.Params)
	}{
		{
			name:     "defaults",
			pageArg:  "1",
			limitArg: "100",
			checkParams: func(t *testing.T, p zentao.Params) {
				if p.Get("recPerPage") != "100" {
					t.Errorf("expected recPerPage=100, got %s", p.Get("recPerPage"))
				}
				if p.Get("recTotal") != fullPullMarker {
					t.Errorf("expected recTotal=%s, got %s", fullPullMarker, p.Get("recTotal"))
				}
				if p.Get("pageID") != "" {
					t.Errorf("page 1 should not set pageID, got %s", p.Get("pageID"))
				}
			},
		},
		{
			name:     "page 2 with limit 20",
			pageArg:  "2",
			limitArg: "20",
			checkParams: func(t *testing.T, p zentao.Params) {
				if p.Get("recPerPage") != "20" {
					t.Errorf("expected recPerPage=20, got %s", p.Get("recPerPage"))
				}
				if p.Get("pageID") != "2" {
					t.Errorf("expected pageID=2, got %s", p.Get("pageID"))
				}
			},
		},
		{
			name:     "limit all",
			pageArg:  "1",
			limitArg: "all",
			checkParams: func(t *testing.T, p zentao.Params) {
				if p.Get("recPerPage") != fullPullMarker {
					t.Errorf("expected recPerPage=%s, got %s", fullPullMarker, p.Get("recPerPage"))
				}
				if p.Get("pageID") != "" {
					t.Errorf("expected pageID to be unset for all, got %s", p.Get("pageID"))
				}
			},
		},
		{
			name:     "limit empty string translates to all",
			pageArg:  "1",
			limitArg: " ",
			checkParams: func(t *testing.T, p zentao.Params) {
				if p.Get("recPerPage") != fullPullMarker {
					t.Errorf("expected recPerPage=%s, got %s", fullPullMarker, p.Get("recPerPage"))
				}
			},
		},
		{
			name:          "invalid page zero",
			pageArg:       "0",
			limitArg:      "100",
			expectErr:     true,
			expectedErrIs: zentao.ErrValidation,
		},
		{
			name:          "invalid page negative",
			pageArg:       "-1",
			limitArg:      "100",
			expectErr:     true,
			expectedErrIs: zentao.ErrValidation,
		},
		{
			name:          "invalid limit negative",
			pageArg:       "1",
			limitArg:      "-5",
			expectErr:     true,
			expectedErrIs: zentao.ErrValidation,
		},
		{
			name:          "invalid limit zero",
			pageArg:       "1",
			limitArg:      "0",
			expectErr:     true,
			expectedErrIs: zentao.ErrValidation,
		},
		{
			name:          "invalid limit non-numeric text",
			pageArg:       "1",
			limitArg:      "foo",
			expectErr:     true,
			expectedErrIs: zentao.ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dummyCmd := &cobra.Command{Use: "test"}
			addPaginationFlags(dummyCmd)

			if tt.pageArg != "" {
				_ = dummyCmd.Flags().Set("page", tt.pageArg)
			}
			if tt.limitArg != "" {
				_ = dummyCmd.Flags().Set("limit", tt.limitArg)
			}

			params := zentao.Params{}
			err := applyPagination(dummyCmd, params)

			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.expectedErrIs != nil && !errors.Is(err, tt.expectedErrIs) {
					t.Errorf("expected error wrapping %v, got %v", tt.expectedErrIs, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.checkParams != nil {
				tt.checkParams(t, params)
			}
		})
	}
}

func TestPagination_CommandWithoutFlags_IsNoop(t *testing.T) {
	cmdWithoutFlags := &cobra.Command{Use: "test"}
	params := zentao.Params{"foo": {"bar"}}
	if err := applyPagination(cmdWithoutFlags, params); err != nil {
		t.Fatalf("applyPagination on command without flags should return nil, got: %v", err)
	}
	if params.Get("foo") != "bar" {
		t.Errorf("expected params to remain unchanged, got %v", params)
	}
}
