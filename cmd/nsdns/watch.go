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
	"github.com/Eagerod/kube-namesilo-dns/pkg/icanhazip"
	"github.com/Eagerod/kube-namesilo-dns/pkg/namesilo_api"
)

func watchCommand() *cobra.Command {
	var ingressClass string
	var domainName string

	watchCmd := &cobra.Command{
		Use:   "watch",
		Short: "watch ingress objects, and create dns records",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.SetLevel(log.DebugLevel)

			if domainName == "" {
				return fmt.Errorf("Must provide a domain name to update DNS records.")
			}

			if ingressClass == "" {
				return fmt.Errorf("Must provide an ingress class to select DNS records.")
			}

			nsApiKey := os.Getenv("NAMESILO_API_KEY")
			if nsApiKey == "" {
				return fmt.Errorf("Failed to find NAMESILO_API_KEY in environment. Cannot proceed.")
			}

			api := namesilo_api.NewNamesiloApi(domainName, nsApiKey)

			rrMutex := sync.RWMutex{}
			var records []namesilo_api.ResourceRecord
			var ip string

			refreshState := func() error {
				var err error
				rrMutex.Lock()
				defer rrMutex.Unlock()

				records, err = api.ListDNSRecords()
				if err != nil {
					return err
				}
				ip, err = icanhazip.GetPublicIP()
				if err != nil {
					return err
				}

				log.Debug("Updated local resource records and IP address")

				return nil
			}

			// Don't actually start the informer until basic information is
			//   available.
			done := make(chan struct{})
			go func() {
				if err := refreshState(); err != nil {
					panic(err)
				}

				done <- struct{}{}

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
						if !ShouldProcessIngress(ingressClass, ingress) {
							return
						}

						updateRecord, err := OneUpdatedOrAddedResourceRecord(records, ingress, domainName, ip, false)
						if err != nil {
							log.Error(err)
							return
						}

						if updateRecord == nil {
							return
						}

						log.Infof("Adding record for %s", updateRecord.Host)
						if err := api.AddDNSRecord(*updateRecord); err != nil {
							log.Error(err)
						}

						if err := refreshState(); err != nil {
							log.Error(err)
						}
					},
					DeleteFunc: func(obj interface{}) {
						ingress := obj.(*networkingv1.Ingress)
						if !ShouldProcessIngress(ingressClass, ingress) {
							return
						}

						updateRecord, err := OneUpdatedOrAddedResourceRecord(records, ingress, domainName, ip, true)
						if err != nil {
							log.Error(err)
							return
						}

						if updateRecord == nil {
							return
						}

						log.Infof("Deleting resource record %s", updateRecord.RecordId)
						err = api.DeleteDNSRecord(*updateRecord)
						if err != nil {
							log.Error(err.Error())
							return
						}

						if err := refreshState(); err != nil {
							log.Error(err)
						}
					},
					UpdateFunc: func(old, new interface{}) {
						ingress := new.(*networkingv1.Ingress)
						if !ShouldProcessIngress(ingressClass, ingress) {
							return
						}

						updateRecord, err := OneUpdatedOrAddedResourceRecord(records, ingress, domainName, ip, false)
						if err != nil {
							log.Error(err)
							return
						}

						if updateRecord == nil {
							return
						}

						if err := api.UpdateDNSRecord(*updateRecord); err != nil {
							log.Error(err)
						}

						if err := refreshState(); err != nil {
							log.Error(err)
						}
					},
				},
			)

			<-done

			stop := make(chan struct{})
			informerFactory.Start(stop)
			informerFactory.WaitForCacheSync(stop)

			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

			<-sig

			return nil
		},
	}

	watchCmd.Flags().StringVarP(&ingressClass, "ingress-class", "i", "", "ingress class to use for public DNS records")
	watchCmd.Flags().StringVarP(&domainName, "domain", "d", "", "domain name for API calls")

	return watchCmd
}
