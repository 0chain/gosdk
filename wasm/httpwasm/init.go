package httpwasm

import (
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterMinerHandlers(s *mux.Router) {
	s.HandleFunc("/v1/transaction/put", smartContractTxnValue).Methods(http.MethodPost)
	s.HandleFunc("/v1/client/put", createWallet).Methods(http.MethodPost, http.MethodGet)
	s.HandleFunc("/v1/client/get", getClientDetail).Methods(http.MethodGet)
}

func RegisterSharderHandlers(s *mux.Router) {
	s.HandleFunc("/v1/transaction/get/confirmation", verifyTransaction).Methods(http.MethodPost, http.MethodGet)
	s.HandleFunc("/v1/screst/6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7/getblobbers", getBlobbers).Methods(http.MethodGet)
	s.HandleFunc("/v1/screst/6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7/getBlobber", getBlobber).Methods(http.MethodGet)
	s.HandleFunc("/v1/screst/6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7/getReadPoolStat", getPoolStat).Methods(http.MethodGet)
	s.HandleFunc("/v1/screst/6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7/getStakePoolStat", getStakePoolStat).Methods(http.MethodGet)
	s.HandleFunc("/v1/screst/6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7/getUserStakePoolStat", getUserStakePoolStat).Methods(http.MethodGet)
	s.HandleFunc("/v1/screst/6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7/getWritePoolStat", getPoolStat).Methods(http.MethodGet)
	s.HandleFunc("/v1/screst/6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7/getChallengePoolStat", getChallengePoolStat).Methods(http.MethodGet)
	s.HandleFunc("/v1/screst/6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7/getConfig", getConfig).Methods(http.MethodGet)
	s.HandleFunc("/v1/screst/6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7/allocation", allocation).Methods(http.MethodGet)
	s.HandleFunc("/v1/screst/6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7/allocations", allocations).Methods(http.MethodGet)
	s.HandleFunc("/v1/screst/6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7/allocation_min_lock", allocationMinLock).Methods(http.MethodGet)

}

func RegisterDefaultHandlers(s *mux.Router) {
	s.HandleFunc("/dns/network", getNetwork).Methods(http.MethodGet)
	s.HandleFunc("/commitfabric", commitToFabric).Methods(http.MethodPost)
}
