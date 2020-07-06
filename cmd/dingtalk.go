package cmd

import (
	"github.com/bitnami-labs/kubewatch/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var dingTalkConfigCmd = &cobra.Command{
	Short: "specific dingtalk configuration",
	Long:  "specific dingtalk configuration",
	Use:   "dingtalk",
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := config.New()
		if err != nil {
			logrus.Fatal(err)
		}
		url, err := cmd.Flags().GetString("url")
		if err == nil {
			if len(url) > 0 {
				conf.Handler.DingTalk.Url = url
			}
		} else {
			logrus.Fatal(err)
		}

		secret, err := cmd.Flags().GetString("secret")
		if err == nil {
			if len(url) > 0 {
				conf.Handler.DingTalk.Secret = secret
			}
		} else {
			logrus.Fatal(err)
		}

		if err = conf.Write(); err != nil {
			logrus.Fatal(err)
		}

	},
}

func init() {
	dingTalkConfigCmd.Flags().StringP("url", "u", "", "Specify dingtalk url")
	dingTalkConfigCmd.Flags().StringP("secret", "s", "", "Specify dingtalk secret")
}
