// Finance Tracker plugin for Cortex.
// This binary is launched as a subprocess by the Cortex host.
package main

import (
	"github.com/alvarotorresc/cortex/pkg/sdk"
)

func main() {
	sdk.Serve(&FinancePlugin{})
}
