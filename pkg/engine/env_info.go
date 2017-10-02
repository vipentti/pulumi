// Copyright 2017, Pulumi Corporation.  All rights reserved.

package engine

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi/pkg/tokens"

	"github.com/pkg/errors"
)

func (eng *Engine) GetCurrentEnvName() tokens.QName {
	return eng.getCurrentEnv()
}

func (eng *Engine) EnvInfo(envName tokens.QName, showIDs bool, showURNs bool) error {
	curr := envName
	if curr == "" {
		curr = eng.getCurrentEnv()
	}
	if curr == "" {
		return errors.New("no current environment; either `pulumi env init` or `pulumi env select` one")
	}

	fmt.Fprintf(eng.Stdout, "Current environment is %v\n", curr)
	fmt.Fprintf(eng.Stdout, "    (use `pulumi env select` to change environments; `pulumi env ls` lists known ones)\n")
	target, snapshot, checkpoint, err := eng.Environment.GetEnvironment(curr)
	if err != nil {
		return err
	}
	if checkpoint.Latest != nil {
		fmt.Fprintf(eng.Stdout, "Last update at %v\n", checkpoint.Latest.Time)
		if checkpoint.Latest.Info != nil {
			info, err := json.MarshalIndent(checkpoint.Latest.Info, "    ", "    ")
			if err != nil {
				return err
			}
			fmt.Fprintf(eng.Stdout, "Additional update info:\n    %s\n", string(info))
		}
	}
	if target.Config != nil && len(target.Config) > 0 {
		fmt.Fprintf(eng.Stdout,
			"%v configuration variables set (see `pulumi config` for details)\n", len(target.Config))
	}
	if snapshot == nil || len(snapshot.Resources) == 0 {
		fmt.Fprintf(eng.Stdout, "No resources currently in this environment\n")
	} else {
		fmt.Fprintf(eng.Stdout, "%v resources currently in this environment:\n", len(snapshot.Resources))
		fmt.Fprintf(eng.Stdout, "\n")
		fmt.Fprintf(eng.Stdout, "%-48s %s\n", "TYPE", "NAME")
		for _, res := range snapshot.Resources {
			fmt.Fprintf(eng.Stdout, "%-48s %s\n", res.Type, res.URN.Name())

			// If the ID and/or URN is requested, show it on the following line.  It would be nice to do this
			// on a single line, but they can get quite lengthy and so this formatting makes more sense.
			if showIDs {
				fmt.Fprintf(eng.Stdout, "\tID: %s\n", res.ID)
			}
			if showURNs {
				fmt.Fprintf(eng.Stdout, "\tURN: %s\n", res.URN)
			}
		}
	}
	return nil
}
