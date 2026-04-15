package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultpatch/internal/quota"
)

func init() {
	var (
		rulesFlag []string
		jsonOut   bool
	)

	cmd := &cobra.Command{
		Use:   "quota",
		Short: "Check secret counts against per-path quotas",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runQuota(rulesFlag, jsonOut)
		},
	}

	cmd.Flags().StringArrayVarP(&rulesFlag, "rule", "r", nil,
		"Quota rule in path=limit format (repeatable)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output violations as JSON")
	_ = cmd.MarkFlagRequired("rule")

	rootCmd.AddCommand(cmd)
}

func runQuota(rulesFlag []string, jsonOut bool) error {
	vc, err := buildClient()
	if err != nil {
		return err
	}

	rules, err := parseQuotaRules(rulesFlag)
	if err != nil {
		return err
	}

	checker := quota.NewChecker(vc, rules)
	violations, err := checker.Check()
	if err != nil {
		return err
	}

	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(violations)
	}

	out := quota.Format(violations)
	if out == "" {
		fmt.Println("all quotas satisfied")
		return nil
	}
	fmt.Println(out)
	if len(violations) > 0 {
		os.Exit(1)
	}
	return nil
}

func parseQuotaRules(flags []string) ([]quota.Rule, error) {
	rules := make([]quota.Rule, 0, len(flags))
	for _, f := range flags {
		parts := strings.SplitN(f, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid rule %q: expected path=limit", f)
		}
		limit, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid limit in rule %q: %w", f, err)
		}
		rules = append(rules, quota.Rule{Path: parts[0], Limit: limit})
	}
	return rules, nil
}
