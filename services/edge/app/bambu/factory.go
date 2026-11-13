// Package bambu selects and constructs a farm.ControlPlane for the edge agent.
package bambu

import (
	"fmt"
	"log/slog"

	"github.com/jjheffernan/BambuFarmer/packages/farm"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/bambu/cdp"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/bambu/direct"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/bambu/fixtureplane"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/bambu/native"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/bambu/none"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/config"
)

// Open builds the ControlPlane for this edge process from env-derived config.
//
// Selection:
//  1. FARM_MANAGER_TRANSPORT=direct_printer → in-process printer list (no LAN)
//  2. FARM_MANAGER_URL set + transport cdp → CDP stub
//  3. FARM_MANAGER_URL set → fm_native
//  4. EDGE_FIXTURE_DIR set (and no URL) → fixture
//  5. else → none
func Open(cfg config.Config, log *slog.Logger) (farm.ControlPlane, error) {
	if log == nil {
		log = slog.Default()
	}
	if cfg.FarmManagerTransport == string(farm.KindDirect) {
		log.Info("direct printer control plane")
		return direct.New(cfg.DirectPrinters), nil
	}
	if cfg.Configured() {
		switch cfg.FarmManagerTransport {
		case "cdp":
			log.Warn("CDP control plane is a stub; prefer fm_native until CDP lands")
			return cdp.New(cfg.CDPURL), nil
		case "fm_native", "":
			return native.New(native.Config{
				BaseURL:     cfg.FarmManagerURL,
				User:        cfg.FarmManagerUser,
				Password:    cfg.FarmManagerPassword,
				TLSInsecure: cfg.TLSInsecure,
				CertDir:     cfg.CertDir,
			})
		default:
			return nil, fmt.Errorf("%w: %s", farm.ErrUnsupportedKind, cfg.FarmManagerTransport)
		}
	}
	if cfg.UseFixtures() {
		return fixtureplane.New(cfg.FixtureDir, log)
	}
	return none.New(), nil
}
