package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/windosx/zentao-cli/pkg/zentao"
)

// Pagination defaults for list commands. "all" keeps the legacy behavior of
// pulling every record (the recTotal=recPerPage=999999 trick that bypasses
// the ZenTao pager); a numeric limit pages server-side to protect large
// instances from full-table loads.
const (
	defaultPageLimit = "100"
	fullPullMarker   = "999999"
)

// addPaginationFlags registers the shared --page / --limit flags on a list
// command. Pair it with applyPagination in the command's RunE.
func addPaginationFlags(cmd *cobra.Command) {
	cmd.Flags().Int("page", 1, "页码 (从 1 开始)")
	cmd.Flags().String("limit", defaultPageLimit, "每页条数 (正整数，或 all 表示全量拉取)")
}

// applyPagination translates the --page / --limit flags into ZenTao pager
// params (recPerPage / pageID) on params. It is a no-op on commands without
// pagination flags. recTotal is always set to a large value so the server
// treats any pageID as valid.
func applyPagination(cmd *cobra.Command, params zentao.Params) error {
	pageFlag := cmd.Flags().Lookup("page")
	limitFlag := cmd.Flags().Lookup("limit")
	if pageFlag == nil || limitFlag == nil {
		return nil
	}

	page, err := strconv.Atoi(pageFlag.Value.String())
	if err != nil || page < 1 {
		return fmt.Errorf("%w: --page 必须是 >= 1 的整数", zentao.ErrValidation)
	}

	limit := strings.TrimSpace(limitFlag.Value.String())
	if limit == "" || strings.EqualFold(limit, "all") {
		params.Set("recPerPage", fullPullMarker)
		params.Set("recTotal", fullPullMarker)
		params.Del("pageID")
		return nil
	}

	n, err := strconv.Atoi(limit)
	if err != nil || n <= 0 {
		return fmt.Errorf("%w: --limit 必须是正整数或 all，当前值: %q", zentao.ErrValidation, limit)
	}
	params.Set("recPerPage", strconv.Itoa(n))
	params.Set("recTotal", fullPullMarker)
	if page > 1 {
		params.Set("pageID", strconv.Itoa(page))
	}
	return nil
}
