// Code generated by gobind. DO NOT EDIT.

// JNI function headers for the Go <=> Java bridge.
//
//   autogenerated by gobind -lang=java command-line-arguments

#ifndef __Zcncore_H__
#define __Zcncore_H__

#include <jni.h>

extern jclass proxy_class_zcncore_AuthCallback;
extern jmethodID proxy_class_zcncore_AuthCallback_cons;

void cproxyzcncore_AuthCallback_OnSetupComplete(int32_t refnum, nint status, nstring err);

extern jclass proxy_class_zcncore_GetBalanceCallback;
extern jmethodID proxy_class_zcncore_GetBalanceCallback_cons;

void cproxyzcncore_GetBalanceCallback_OnBalanceAvailable(int32_t refnum, nint status, int64_t value, nstring info);

extern jclass proxy_class_zcncore_GetInfoCallback;
extern jmethodID proxy_class_zcncore_GetInfoCallback_cons;

void cproxyzcncore_GetInfoCallback_OnInfoAvailable(int32_t refnum, nint op, nint status, nstring info, nstring err);

extern jclass proxy_class_zcncore_GetMintNonceCallback;
extern jmethodID proxy_class_zcncore_GetMintNonceCallback_cons;

void cproxyzcncore_GetMintNonceCallback_OnBalanceAvailable(int32_t refnum, nint status, int64_t value, nstring info);

extern jclass proxy_class_zcncore_GetNonceCallback;
extern jmethodID proxy_class_zcncore_GetNonceCallback_cons;

void cproxyzcncore_GetNonceCallback_OnNonceAvailable(int32_t refnum, nint status, int64_t nonce, nstring info);

extern jclass proxy_class_zcncore_GetNotProcessedZCNBurnTicketsCallback;
extern jmethodID proxy_class_zcncore_GetNotProcessedZCNBurnTicketsCallback_cons;

// skipped method GetNotProcessedZCNBurnTicketsCallback.OnBalanceAvailable with unsupported parameter or return types

extern jclass proxy_class_zcncore_WalletCallback;
extern jmethodID proxy_class_zcncore_WalletCallback_cons;

void cproxyzcncore_WalletCallback_OnWalletCreateComplete(int32_t refnum, nint status, nstring wallet, nstring err);

extern jclass proxy_class_zcncore_BurnTicket;
extern jmethodID proxy_class_zcncore_BurnTicket_cons;
extern jclass proxy_class_zcncore_GetClientResponse;
extern jmethodID proxy_class_zcncore_GetClientResponse_cons;
extern jclass proxy_class_zcncore_GetMintNonceCallbackStub;
extern jmethodID proxy_class_zcncore_GetMintNonceCallbackStub_cons;
extern jclass proxy_class_zcncore_GetNonceCallbackStub;
extern jmethodID proxy_class_zcncore_GetNonceCallbackStub_cons;
extern jclass proxy_class_zcncore_GetNotProcessedZCNBurnTicketsCallbackStub;
extern jmethodID proxy_class_zcncore_GetNotProcessedZCNBurnTicketsCallbackStub_cons;
#endif
