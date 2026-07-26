// internal/api/store.go
package api

import (
	"encoding/json"
	"net/http"
	"horizon-core-engine/internal/integration"
)

var integrationManager *integration.IntegrationManager

func InitIntegration() {
	integrationManager = integration.NewIntegrationManager()
	integrationManager.RegisterClient("Ethereum", integration.NewEtherscanClient())
	integrationManager.RegisterClient("BNB Smart Chain", integration.NewBscScanClient())
	integrationManager.RegisterClient("Avalanche", integration.NewSnowScanClient())
	integrationManager.RegisterClient("Arbitrum", integration.NewArbiscanClient())
	integrationManager.RegisterClient("Optimism", integration.NewOptimismClient())
}

func GetNetworkBalancesHandler(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "address parameter required", http.StatusBadRequest)
		return
	}

	balances := integrationManager.GetBalanceFromAllNetworks(address)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"address":  address,
		"balances": balances,
	})
}

func GetNetworksHandler(w http.ResponseWriter, r *http.Request) {
	networks := integrationManager.GetAllNetworks()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"networks": networks,
	})
}
