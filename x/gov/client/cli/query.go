package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	// Group gov queries under a subcommand
	govQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the governance module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	govQueryCmd.AddCommand(
		GetCmdQueryQuorums(),
		GetCmdQueryMinDeposit(),
		GetCmdQueryMinInitialDeposit(),
		GetCmdQueryParticipationEMAs(),
		GetCmdQueryGovernor(),
		GetCmdQueryGovernors(),
		GetCmdQueryGovernanceDelegation(),
		GetCmdQueryGovernanceDelegations(),
		GetCmdQueryGovernorValShares(),
	)

	return govQueryCmd
}

// GetCmdQueryQuorums implements the query quorums command.
func GetCmdQueryQuorums() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quorums",
		Short: "Query the current state of the dynamic quorums",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the current state of all the dynamic quorums.

Example:
$ %s query gov quorums
`,
				version.AppName,
			),
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := v1.NewQueryClient(clientCtx)

			// Query store for all 3 params
			ctx := cmd.Context()

			quorumRes, err := queryClient.Quorums(ctx, &v1.QueryQuorumsRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(quorumRes)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryMinDeposit implements the query min deposit command.
func GetCmdQueryMinDeposit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "min-deposit",
		Args:  cobra.ExactArgs(0),
		Short: "Query the minimum deposit currently needed for a proposal to enter voting period",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the minimum deposit needed for a proposal to enter voting period.

Example:
$ %s query gov min-deposit
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			queryClient := v1.NewQueryClient(clientCtx)

			resp, err := queryClient.MinDeposit(cmd.Context(), &v1.QueryMinDepositRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryMinInitialDeposit implements the query min initial deposit command.
func GetCmdQueryMinInitialDeposit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "min-initial-deposit",
		Args:  cobra.ExactArgs(0),
		Short: "Query the minimum initial deposit needed for a proposal to enter deposit period",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the minimum initial deposit needed for a proposal to enter deposit period.

Example:
$ %s query gov min-initial-deposit
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			queryClient := v1.NewQueryClient(clientCtx)

			resp, err := queryClient.MinInitialDeposit(cmd.Context(), &v1.QueryMinInitialDepositRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryParticipationEMAs implements the query ParticipationEMAs command.
func GetCmdQueryParticipationEMAs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "participation",
		Short: "Query the current state of the exponential moving average of the proposal participations",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the current state of the exponential moving average of the proposal participations.

Example:
$ %s query gov participation
`,
				version.AppName,
			),
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := v1.NewQueryClient(clientCtx)

			// Query store for all 3 params
			ctx := cmd.Context()

			participationEMARes, err := queryClient.ParticipationEMAs(ctx, &v1.QueryParticipationEMAsRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(participationEMARes)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryGovernor implements the query governor command.
func GetCmdQueryGovernor() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "governor [address]",
		Args:  cobra.ExactArgs(1),
		Short: "Query details of a single governor",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details for a governor by its address.
Example:
$ %s query gov governor cosmosgov1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			addr, err := types.GovernorAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1.NewQueryClient(clientCtx)

			res, err := queryClient.Governor(
				cmd.Context(),
				&v1.QueryGovernorRequest{GovernorAddress: addr.String()},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res.Governor)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryGovernors implements the query governors command.
func GetCmdQueryGovernors() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "governors",
		Short: "Query all governors",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all governors.
Example:
$ %s query gov governors
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1.NewQueryClient(clientCtx)

			res, err := queryClient.Governors(
				cmd.Context(),
				&v1.QueryGovernorsRequest{},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryGovernanceDelegation implements the query governance delegation command.
func GetCmdQueryGovernanceDelegation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegation [address]",
		Args:  cobra.ExactArgs(1),
		Short: "Query governance delegation for a delegator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details for a governance delegation by a delegator.
Example:
$ %s query gov delegation cosmos1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1.NewQueryClient(clientCtx)

			res, err := queryClient.GovernanceDelegation(
				cmd.Context(),
				&v1.QueryGovernanceDelegationRequest{DelegatorAddress: addr.String()},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryGovernanceDelegations implements the query governance delegations command.
func GetCmdQueryGovernanceDelegations() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegations [governor_address]",
		Short: "Query all governance delegations for a governor",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all governance delegations for a governor.
Example:
$ %s query gov delegations cosmosgov1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			addr, err := types.GovernorAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1.NewQueryClient(clientCtx)

			res, err := queryClient.GovernanceDelegations(
				cmd.Context(),
				&v1.QueryGovernanceDelegationsRequest{GovernorAddress: addr.String()},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryGovernorValShares implements the query governor validator shares command.
func GetCmdQueryGovernorValShares() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vshares [governor_address]",
		Args:  cobra.ExactArgs(1),
		Short: "Query governor virtual validators shares",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details for a governor virtual validators shares by its address.
Example:
$ %s query gov vshares cosmosgov1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			addr, err := types.GovernorAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1.NewQueryClient(clientCtx)

			res, err := queryClient.GovernorValShares(
				cmd.Context(),
				&v1.QueryGovernorValSharesRequest{GovernorAddress: addr.String()},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
