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
	s.HandleFunc("/v1/client/get/balance", sharderGetBalance).Methods(http.MethodGet)
	s.HandleFunc("/v1/screst/cf8d0df9bd8cc637a4ff4e792ffe3686da6220c45f0e1103baa609f3f1751ef4/getLockConfig", sharderGetBalance).Methods(http.MethodGet)
	s.HandleFunc("/v1/screst/cf8d0df9bd8cc637a4ff4e792ffe3686da6220c45f0e1103baa609f3f1751ef4/getPoolsStats", sharderGetBalance).Methods(http.MethodGet)
	s.HandleFunc("/v1/screst/2bba5b05949ea59c80aed3ac3474d7379d3be737e8eb5a968c52295e48333ead/getPoolInfo", vestingPoolInfo).Methods(http.MethodGet)
	s.HandleFunc("/v1/screst/2bba5b05949ea59c80aed3ac3474d7379d3be737e8eb5a968c52295e48333ead/getClientPools", getClientPools).Methods(http.MethodGet)
	s.HandleFunc("/v1/screst/2bba5b05949ea59c80aed3ac3474d7379d3be737e8eb5a968c52295e48333ead/getConfig", getVestingSCConfig).Methods(http.MethodGet)
	s.HandleFunc("/v1/screst/6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d9/getMinerList", getMinerList).Methods(http.MethodGet)
	s.HandleFunc("/v1/screst/6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d9/getSharderList", getSharderList).Methods(http.MethodGet)
	s.HandleFunc("/v1/screst/6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d9/nodeStat", getNodeStat).Methods(http.MethodGet)

}

func RegisterDefaultHandlers(s *mux.Router) {
	s.HandleFunc("/dns/network", getNetwork).Methods(http.MethodGet)
	s.HandleFunc("/commitfabric", commitToFabric).Methods(http.MethodPost)
	s.HandleFunc("/setup", setupAuthHost).Methods(http.MethodPost)
	s.HandleFunc("/_nh/whoami", setupAuthHost).Methods(http.MethodGet)
}
