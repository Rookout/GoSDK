#ifndef _HOOK_API_H
#define _HOOK_API_H

#include <stddef.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif



int RookoutInit(void *user_addr_hint);


int RookoutDestroy();


int RookoutRegisterFunctionBreakpointsState(void *function_entry_addr, void *function_end_addr, int num_breakpoints,
                                            void *breakpoints_addrs, void *breakpoint_failed_counters,
                                            void *breakpoint_execution_callback_addr, void *prologue_addr,
                                            int prologue_len, uint32_t function_stack_usage);


void *RookoutGetInstructionMapping(void *function_entry_addr, void *function_end_addr, int state_id);


void *RookoutGetUnpatchedInstructionMapping(void *function_entry_addr, void *function_end_addr);


int RookoutGetFunctionType(void *function_entry_addr, void *function_end_addr);


uint64_t RookoutGetDangerZoneStartAddress(void *function_entry_addr, void *function_end_addr);


uint64_t RookoutGetDangerZoneEndAddress(void *function_entry_addr, void *function_end_addr);


uint64_t RookoutGetHookAddress(void *function_entry_addr, void *function_end_addr, int state_id);


int RookoutGetHookSizeBytes(void *function_entry_addr, void *function_end_addr, int state_id);


void *RookoutGetHookBytesView(void *function_entry_addr, void *function_end_addr, int state_id);


int RookoutApplyBreakpointsState(void *function_entry_addr, void *function_end_addr, int state_id);


int RookoutClearAllBreakpoints(void *function_entry_addr, void *function_end_addr);


int RookoutTriggerWatchDog(unsigned long long timeout_ms);


void RookoutDefuseWatchDog();


const char *RookoutGetHookerLastError();

#ifdef __cplusplus
}
#endif
#endif 
