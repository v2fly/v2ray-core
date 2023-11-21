package engineering

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/v2fly/v2ray-core/v5/app/subscription/entries/nonnative"
	"github.com/v2fly/v2ray-core/v5/main/commands/base"
)

type valueContainer struct {
	key, value string
}

type orderedValueContainer []valueContainer

func (o *orderedValueContainer) Len() int {
	return len(*o)
}

func (o *orderedValueContainer) Less(i, j int) bool {
	return strings.Compare((*o)[i].key, (*o)[j].key) < 0
}

func (o *orderedValueContainer) Swap(i, j int) {
	(*o)[i], (*o)[j] = (*o)[j], (*o)[i]
}

var cmdNonNativeLinkExtract = &base.Command{
	UsageLine: "{{.Exec}} engineering nonnativelinkextract",
	Flag: func() flag.FlagSet {
		fs := flag.NewFlagSet("", flag.ExitOnError)
		return *fs
	}(),
	Run: func(cmd *base.Command, args []string) {
		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			base.Fatalf("%s", err)
		}
		flattenedLink := nonnative.ExtractAllValuesFromBytes(content)
		var valueContainerOrdered orderedValueContainer

		for key, value := range flattenedLink.Values {
			valueContainerOrdered = append(valueContainerOrdered, valueContainer{key, value})
		}
		sort.Sort(&valueContainerOrdered)
		for _, valueContainer := range valueContainerOrdered {
			io.WriteString(os.Stdout, fmt.Sprintf("%s=%s\n", valueContainer.key, valueContainer.value))
		}
	},
}
