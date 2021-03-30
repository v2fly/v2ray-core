package control

import (
	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/platform/filesystem"
	"github.com/v2fly/v2ray-core/v4/common/platform/securedload"
	"github.com/v2fly/v2ray-core/v4/common/templates"
	"os"
	"text/template"
)

type SubscriptionParseCommand struct {
}

func (s SubscriptionParseCommand) Name() string {
	return "subscriptionParse"
}

func (s SubscriptionParseCommand) Description() Description {
	return Description{
		Short: "a tool to parse all kind of subscription into V2Ray outbound",
		Usage: []string{
			"v2ctl subscriptionParse <subscription file>",
		},
	}
}

func (s SubscriptionParseCommand) Execute(args []string) error {
	templatedef, err := securedload.GetAssetSecured("subscriptions/subscriptionsDefinition.v2flyTemplate")
	if err != nil {
		return newError("Cannot load subscription template file").Base(err)
	}

	templatedata, errtempl := template.New("").Funcs(templates.AssistFunctions).Parse(string(templatedef))
	if errtempl != nil {
		return newError("Cannot parse subscription template file").Base(errtempl)
	}
	subscription, errsubscription := filesystem.ReadFile(args[0])
	if errsubscription != nil {
		return newError("cannot read subscriptions file")
	}

	dot := templates.NewUniversalDot(subscription)

	if errtemp := templatedata.Execute(os.Stdout, dot); errtemp != nil {
		return newError("Cannot load execute template file").Base(errtemp)
	}

	return nil
}

func init() {
	common.Must(RegisterCommand(SubscriptionParseCommand{}))
}
