package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

import (
	"github.com/Eagerod/kube-namesilo-dns/pkg/nsdns"
)

func watchCommand() *cobra.Command {
	var ingressClass string
	var domainName string

	watchCmd := &cobra.Command{
		Use:   "watch",
		Short: "watch ingress objects, and create dns records",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.SetLevel(log.DebugLevel)

			dm, err := nsdns.NewDnsManager(domainName, ingressClass)
			if err != nil {
				return err
			}

			dmCache := nsdns.NewDnsManagerCache()

			rrMutex := sync.RWMutex{}
			refreshState := func() error {
				rrMutex.Lock()
				defer rrMutex.Unlock()

				if err := nsdns.UpdateCachedRecords(dmCache, dm.Api); err != nil {
					return err
				}

				if err := nsdns.UpdateIpAddress(dmCache); err != nil {
					return err
				}

				log.Debug("Updated local resource records and IP address")

				return nil
			}

			if err := refreshState(); err != nil {
				return err
			}

			go func() {
				for range time.Tick(time.Hour) {
					if err := refreshState(); err != nil {
						panic(err)
					}
				}
			}()

			clientset, err := GetKubernetesClientSet()
			if err != nil {
				return err
			}

			informerFactory := informers.NewSharedInformerFactory(clientset, time.Minute)

			informerFactory.Networking().V1().Ingresses().Informer().AddEventHandler(
				cache.ResourceEventHandlerFuncs{
					AddFunc: func(obj interface{}) {
						ingress := obj.(*networkingv1.Ingress)
						if !dm.ShouldProcessIngress(ingress) {
							return
						}

						record, err := nsdns.NamesiloRecordFromIngress(ingress, dm.BareDomainName, dmCache.CurrentIpAddress)
						if err != nil {
							log.Error(err)
							return
						}

						existingRecord, _ := RecordMatching(dmCache.CurrentRecords, *record)
						if existingRecord != nil {
							return
						}

						log.Infof("Adding record for %s", record.Host)
						if err := dm.Api.AddDNSRecord(*record); err != nil {
							log.Error(err)
						}

						if err := refreshState(); err != nil {
							log.Error(err)
						}
					},
					DeleteFunc: func(obj interface{}) {
						ingress := obj.(*networkingv1.Ingress)
						if !dm.ShouldProcessIngress(ingress) {
							return
						}

						record, err := nsdns.NamesiloRecordFromIngress(ingress, dm.BareDomainName, dmCache.CurrentIpAddress)
						if err != nil {
							log.Error(err)
							return
						}

						deleteRecord, _ := RecordMatching(dmCache.CurrentRecords, *record)

						log.Infof("Deleting resource record %s", deleteRecord.RecordId)
						err = dm.Api.DeleteDNSRecord(*deleteRecord)
						if err != nil {
							log.Error(err)
							return
						}

						if err := refreshState(); err != nil {
							log.Error(err)
						}
					},
					UpdateFunc: func(old, new interface{}) {
						ingress := new.(*networkingv1.Ingress)
						if !dm.ShouldProcessIngress(ingress) {
							return
						}

						record, err := nsdns.NamesiloRecordFromIngress(ingress, dm.BareDomainName, dmCache.CurrentIpAddress)
						if err != nil {
							log.Error(err)
							return
						}

						// Only update if something actionable changed.
						updateRecord, _ := RecordMatching(dmCache.CurrentRecords, *record)
						if record.EqualsRecord(*updateRecord) {
							return
						}

						log.Infof("Updating record %s", updateRecord.RecordId)
						if err := dm.Api.UpdateDNSRecord(*updateRecord); err != nil {
							log.Error(err)
						}

						if err := refreshState(); err != nil {
							log.Error(err)
						}
					},
				},
			)

			stop := make(chan struct{})
			informerFactory.Start(stop)
			informerFactory.WaitForCacheSync(stop)

			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

			<-sig
			stop <- struct{}{}

			return nil
		},
	}

	watchCmd.Flags().StringVarP(&ingressClass, "ingress-class", "i", "", "ingress class to use for public DNS records")
	watchCmd.Flags().StringVarP(&domainName, "domain", "d", "", "domain name for API calls")

	return watchCmd
}
