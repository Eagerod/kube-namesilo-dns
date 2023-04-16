package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

import (
	"github.com/Eagerod/kube-namesilo-dns/pkg/namesilo_api"
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

			dmCache := nsdns.NewDnsManagerCache()
			if err := nsdns.UpdateCachedRecords(dmCache, dm.Api); err != nil {
				return err
			}

			log.Debugf("Received %d records from Namesilo", len(dmCache.CurrentRecords))

			if err := nsdns.UpdateIpAddress(dmCache); err != nil {
				return err
			}

			ingressRecords, err := GetResourcesFromKubernetesIngresses(dm, dmCache.CurrentIpAddress)
			if err != nil {
				return err
			}

			log.Debugf("Received %d records from Ingress objects", len(ingressRecords))

			rr := ReconcileRecords(dmCache.CurrentRecords, ingressRecords)

			for _, r := range rr.NoOp {
				log.Infof("Skipping %s because it's already up to date.", r.Host)
			}

			for _, record := range rr.Add {
				log.Infof("Adding record for %s", record.Host)
				if err := dm.Api.AddDNSRecord(record); err != nil {
					return err
				}
			}

			for _, record := range rr.Update {
				switch record.Type {
				case "A":
					if err := updateRecordIfNeeded(dm.Api, record, dmCache.CurrentIpAddress); err != nil {
						return err
					}
				case "CNAME":
					if err := updateRecordIfNeeded(dm.Api, record, dm.BareDomainName); err != nil {
						return err
					}
				default:
					log.Infof("Can't update record of type \"%s\"\n", record.Type)
				}
			}

			return nil
		},
	}

	updateCmd.Flags().StringVarP(&ingressClass, "ingress-class", "i", "", "ingress class to use for public DNS records")
	updateCmd.Flags().StringVarP(&domainName, "domain", "d", "", "domain name for API calls")
	return updateCmd
}

func updateRecordIfNeeded(api *namesilo_api.NamesiloApi, record namesilo_api.ResourceRecord, value string) error {
	if record.Value == value {
		log.Infof("Skipping %s record %s because it's already set correctly\n", record.Type, record.Host)
		return nil
	}

	record.Value = value
	if err := api.UpdateDNSRecord(record); err != nil {
		log.Errorf("Failed to update %s record with %s", record.Type, err.Error())
		return err
	}

	log.Infof("Updated %s record %s to %s", record.Type, record.RecordId, value)
	return nil
}
