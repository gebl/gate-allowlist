package allowlist

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"slices"

	"github.com/go-logr/logr"
	"github.com/oschwald/maxminddb-golang"
	"github.com/robinbraemer/event"
	c "go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/netutil"
)

// Plugin is a boss bar plugin that displays a boss bar to the player upon login.
var Plugin = proxy.Plugin{
	Name: "AllowList",
	Init: func(ctx context.Context, p *proxy.Proxy) error {
		log := logr.FromContextOrDiscard(ctx)
		log.Info("Hello from AllowList plugin!")

		event.Subscribe(p.Event(), 0, allowListUserPreLogin(ctx))
		event.Subscribe(p.Event(), 0, allowListUserLogin(ctx))

		return nil
	},
}

func allowListUserPreLogin(ctx context.Context) func(*proxy.PreLoginEvent) {
	return func(e *proxy.PreLoginEvent) {
		var arr []string

		log := logr.FromContextOrDiscard(ctx)
		log.Info(fmt.Sprintf("Hello %s! ", e.Username()))
		content, err := os.ReadFile("./allowlist.json")
		if err != nil {
			log.Info("Error during Unmarshal(): ", err)
			e.Deny(&c.Text{Content: "You must be a subscriber to join this server."})
		} else {
			_ = json.Unmarshal([]byte(content), &arr)
			if !slices.Contains(arr, e.Username()) {
				e.Deny(&c.Text{Content: "You must be a subscriber to join this server."})
				log.Info("User is NOT a subscriber.")
			} else {
				log.Info("User is a subscriber.")
			}
		}

		// Open the GeoLite2 database.
		db, err := maxminddb.Open("/GeoLite2-Country.mmdb")

		if err != nil {
			log.Info("Issue open maxmind GeoLite2 Database")
		}
		defer db.Close()

		var record struct {
			Country struct {
				ISOCode string `maxminddb:"iso_code"`
			} `maxminddb:"country"`
		}

		ip := net.ParseIP(netutil.Host(e.Conn().RemoteAddr()))
		log.Info(fmt.Sprintf("RemoteAddr: %s", ip.String()))

		log.Info(fmt.Sprintf("IP: %s", ip.String()))
		if !ip.IsPrivate() {
			log.Info("IP is not private")

			log.Info(fmt.Sprintf("Checking in maxminddb: %s", e.Conn().RemoteAddr().String()))

			err = db.Lookup(ip, &record)
			if err != nil {
				log.Info("Issue lookup maxmind GeoLite2 Database")
			}

			log.Info(fmt.Sprintf("Country: %s", record.Country.ISOCode))
			if record.Country.ISOCode != "US" {
				e.Deny(&c.Text{Content: "You must be in the US to join this server."})
			}
		} else {
			log.Info("IP is private")
			log.Info(fmt.Sprintf("Private IP: %s", e.Conn().RemoteAddr().String()))

		}

	}
}

func allowListUserLogin(ctx context.Context) func(*proxy.LoginEvent) {
	return func(e *proxy.LoginEvent) {
		var arr []string

		log := logr.FromContextOrDiscard(ctx)
		log.Info(fmt.Sprintf("Hello %s (%s)! ", e.Player().Username(), e.Player().ID()))
		content, err := os.ReadFile("./allowlist.json")
		if err != nil {
			log.Info("Error during Unmarshal(): ", err)
			e.Player().Disconnect(&c.Text{Content: "You must be a subscriber to join this server."})
		} else {
			_ = json.Unmarshal([]byte(content), &arr)
			if !slices.Contains(arr, e.Player().Username()) {
				e.Player().Disconnect(&c.Text{Content: "You must be a subscriber to join this server."})
				log.Info("User is NOT a subscriber.")
			} else {
				log.Info("User is a subscriber.")
			}
		}
	}
}
