package cmd

import (
	"fmt"
	"os"
	"strings"
)

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

import (
	"github.com/Eagerod/kube-namesilo-dns/pkg/icanhazip"
	"github.com/Eagerod/kube-namesilo-dns/pkg/namesilo_api"
)

func updateCommand() *cobra.Command {
	var ingressClass string
	var domainName string

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "update dns records",
		RunE: func(cmd *cobra.Command, args []string) error {
			nsApiKey := os.Getenv("NAMESILO_API_KEY")
			api := namesilo_api.NewNamesiloApi(nsApiKey)

			if domainName == "" {
				return fmt.Errorf("Must provide a domain name to update DNS records.")
			}

			records, err := api.ListDNSRecords(domainName)
			if err != nil {
				return err
			}

			ip, err := icanhazip.GetPublicIP()
			if err != nil {
				return err
			}

			for _, record := range records {
				switch record.Type {
				case "A":
					if record.Value == ip {
						log.Infof("Skipping A record %s because it's already set correctly\n", record.Host)
					} else {
						if err := api.UpdateDNSRecord(domainName, "", record.RecordId, ip, 7207); err != nil {
							return err
						}

						log.Infof("Updated %s record %s to %s", record.Type, record.RecordId, ip)
					}
				case "CNAME":
					if record.Value == domainName {
						log.Infof("Skipping CNAME record %s because it's already set correctly\n", record.Host)
					} else {
						domainSuffix := fmt.Sprintf(".%s", domainName)
						subdomain := strings.TrimSuffix(record.Host, domainSuffix)
						if err := api.UpdateDNSRecord(domainName, subdomain, record.RecordId, domainName, 7207); err != nil {
							return err
						}

						log.Infof("Updated %s record %s to %s", record.Type, record.RecordId, domainName)
					}
				default:
					log.Infof("Can't handle record of type \"%s\"\n", record.Type)
				}
			}

			return nil
		},
	}

	updateCmd.Flags().StringVarP(&ingressClass, "ingress-class", "i", "", "ingress class to use for public DNS records")
	updateCmd.Flags().StringVarP(&domainName, "domain", "d", "", "domain name for API calls")
	return updateCmd
}
