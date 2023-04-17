package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

import (
	"github.com/Eagerod/kube-namesilo-dns/pkg/nsdns"
)

func updateCommand() *cobra.Command {
	var ingressClass string
	var domainName string

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "update dns records",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.SetLevel(log.DebugLevel)

			dm, err := nsdns.NewDnsManager(domainName, ingressClass)
			if err != nil {
				return err
			}

			if err := dm.UpdateCache(); err != nil {
				return err
			}

			ingresses, err := GetIngresses("default")
			if err != nil {
				return err
			}

			for _, i := range ingresses {
				if err := dm.HandleIngressExists(&i); err != nil {
					return err
				}
			}

			return nil
		},
	}

	updateCmd.Flags().StringVarP(&ingressClass, "ingress-class", "i", "", "ingress class to use for public DNS records")
	updateCmd.Flags().StringVarP(&domainName, "domain", "d", "", "domain name for API calls")
	return updateCmd
}
