#ifndef _HOOK_API_H
#define _HOOK_API_H

#include <stdint.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C"
{
#endif
    

    
    int Init(void* user_addr_hint);

    
    int Destroy();

    
    int RegisterFunctionBreakpointsState(void* function_entry_addr, void* function_end_addr, int num_breakpoints, void* breakpoints_addrs, void* breakpoint_execution_callback_addr, void* prologue_callback_addr, void* should_run_prologue, uint32_t function_stack_usage);


    
    void* GetInstructionMapping(void* function_entry_addr, void* function_end_addr, int state_id);

    
    void* GetUnpatchedInstructionMapping(void* function_entry_addr, void* function_end_addr);

    
    int GetFunctionType(void* function_entry_addr, void* function_end_addr);

    
    uint64_t GetDangerZoneStartAddress(void* function_entry_addr, void* function_end_addr);

    
    uint64_t GetDangerZoneEndAddress(void* function_entry_addr, void* function_end_addr);

    
    uint64_t GetHookAddress(void* function_entry_addr, void* function_end_addr, int state_id);

    
    int GetHookSizeBytes(void* function_entry_addr, void* function_end_addr, int state_id);

    
    void* GetHookBytesView(void* function_entry_addr, void* function_end_addr, int state_id);

    
    int GetStackUsageJSON(char* stack_usage_buffer, size_t stack_usage_buffer_size);

    
    int ApplyBreakpointsState(void* function_entry_addr, void* function_end_addr, int state_id);


    
    int ClearAllBreakpoints(void* function_entry_addr, void* function_end_addr);

    
    int TriggerWatchDog(unsigned long long timeout_ms);

    
    void DefuseWatchDog();

    
    const char* GetHookerLastError();



#ifdef __cplusplus
}
#endif
#endif 
