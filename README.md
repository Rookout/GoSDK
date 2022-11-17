# Guide to the Go Rook POC

Table of contents
- [What do we have](#what-do-we-have)
- [Current limitations](#current-limitations)
- [Some Go internals](#some-go-internals)
	- [The Go compiler, assembler and assembler language](#the-go-compiler-assembler-and-assembler-language)
	- [Go functions & stack splits](#go-functions--stack-splits)
	- [Garbage collection and functions](#garbage-collection-and-functions)
	- [Inlining](#inlining)
	- [Go calling convention](#go-calling-convention)
		- [Example](#example)
	- [Deferred functions](#deferred-functions)
	- [More reading](#more-reading-about-go-internals-through-the-lens-of-writing-go-assembly)
- [How Delve (the Go debugger) works & the Delve architecture](#how-delve-the-go-debugger-works--the-delve-architecture)
	- [Basics](#basics)
	- [Initialization and architecture](#initialization-and-architecture)
- [How the Go Rook POC works](#how-the-go-rook-poc-works)
	- [At a high level](#at-a-high-level)
		- [Placing a BP](#placing-a-bp)
		- [Getting a user msg](#getting-a-user-msg)
		- [Components of the Go Rook](#components-of-the-go-rook)
	- [Go Rook internals](#go-rook-internals)
		- [inprocess.Process](#inprocessprocess)
		- [inprocess.inmemoryThread](#inprocessinmemorythread)
		- [The actual hook](#the-actual-hook)
		- [Reading user data and converting to *rook.Variants](#reading-user-data-and-converting-to-rookvariants)
- [What's needed to make this work on every function, on any line and without crashing (TODO)](#whats-needed-to-make-this-work-on-every-function-on-any-line-and-without-crashing-todo)
	- [Replace the prologue hooking with a hooking method that can work on every line](#replace-the-prologue-hooking-with-a-hooking-method-that-can-work-on-every-line)
	- [Prevent crashing when placing a bp](#prevent-crashing-when-placing-a-bp)
	- [Distribution without exposing the sources?](#distribution-without-exposing-the-sources)

## What do we have
- Place breakpoints on some functions
- Read local variables and arguments
- Serialization of (hopefully) all Go types and user types into rook.Variant
- Works on Linux and Windows
## Current limitations
- Placing a breakpoint is not atomic. If the processor is executing the code at the point that's being modified,
  we will crash.
- BPs can only be set on the first line
- Functions which breakpoints can be set on must have a stack-split prologue
- Inlined functions are not supported (since their beginning is in the middle of another function) -
  pass `-gcflags '-N -l'` to `go build` to disable inlining.
- It doesn't compile for Mac (didn't implement it), and the same approach probably won't work for Mac -
  since macOS Catalina you can't `mprotect(PROT_READ|PROT_WRITE|PROT_EXEC)`, which is required for the in place approach.
## Some Go internals
I learned most of the following by just trying to place a hook, then seeing how it crashes. Because the Go compiler puts out one big executable, and you build all of the source code locally -- including the runtime,
it meant that running it under a debugger would make the debugger stop at the point of the crash, pointing to the part of the runtime I broke.

### The Go compiler, assembler and assembler language
- The Go compiler emits a platform-agnostic assembler language that's similar (or identical to?) Plan9. Plan9 is an old OS written by the authors of Go and I didn't read into it any further.
- It's platform-agnostic, but not really, because there are some platform specific stuff the implementation does differently. So the language is agnostic, but you can't just use the same code for very low level stuff (like we are...)
- The Go assembler translates this assembler language to native code.
- Finally the linker runs and puts it all together. The linker also creates some data structures used by the runtime (referring to moduledata, more on this later)
### Go functions & stack splits
- Goroutines start out with 2KB of stack (vs native threads, which on Windows get 1MB). Go functions have a stack split prologue at the beginning, meaning their stack pointer is checked
for if it's about to exceed the end of the stack. If so, a new stack with double the size is allocated and the old stack is copied.
- Copying sounds easy, but consider that there could be pointers to variables on the stack, so the stack copying code needs to adjust those pointers. When implementing an assembly function and we don't want to split, we can use the NOSPLIT annotation.
- This adjustment happens through use of the stack map, which tells the Go runtime which pointers point to what. When implementing an assembly function that doesn't have any pointers, we can use NO_LOCAL_POINTERS.
- There are multiple stack split prologues, depending on the size of the function. Delve can detect them, see the function `proc.GetPrologueLength` and how prologues are defined in Delve at `amd64_disasm.go` - prologuesAMD64.
- In the Go source code, the stacksplit compiler implementation (the bit that makes the compiler emit) can be found at `cmd\internal\obj\x86\obj6.go`, called `stacksplit`.
- You can see straight away that there are multiple prologues:
 ```golang
func stacksplit(ctxt *obj.Link, cursym *obj.LSym, p *obj.Prog, newprog obj.ProgAlloc, framesize int32, textarg int32) *obj.Prog {
...
	if framesize <= objabi.StackSmall {
...
```
### Garbage collection and functions
  - Unrelated but similar, the garbage collector runs through `gentraceback` (in `runtime/traceback.go`) - the `callback` being not nil means it's GC running.
    For the GC to work, your code must be in a static list generated by the linker. See the last section about what's needed ot make this work everywhere to learn more about this list (moduledata). 
### Inlining
- Short Go functions may be inlined, which means that their implementation is copied to the body of calling functions.
  This is done to save the cost of pushing args on stack or the stack split prologue.
- In case they are inlined, that means there are many copies of them in each call site.
- Fortunately, this inlining is described in the Go debug symbols, and Delve can find each of these sites.
- In fact, you don't even need to do anything special: Delve has functions that take filename and linenumber and give you a memory address (or a list of them).
  In case of an inlined function, it'll automatically give you the correct address for all of the callers. Yay! Easy!
### Go calling convention
- In Go, function arguments and return values are all passed on the stack, and all CPU registers are considered clobbered - it's the caller's responsibility to save them.
  This means that when implementing a hook, I could just do whatever I wanted as long as I did it before any hooked function code ran.
- There is a proposal for a register-based calling convention for Go: https://go.googlesource.com/proposal/+/refs/changes/78/248178/1/design/40724-register-calling.md
- That is also where I learned that the existing calling convention looks like this.
- When debugging, I pushed the value 0x1234567812345678 (which is 64bit) to the stack whenever I pushed a real value, and I would add an additional function argument for it, or an extra struct field.
  This helped me see that the stack was aligned correctly. If it wasn't, this value would show up in the wrong spot, and I'd know what mistake I made.
  The other values are real memory addresses, so it's hard by looking at them to tell if it's the right value.
#### Example
Say you have a function that looks like this:
`func hello(a str, b int) {}`
Then when calling it, you need to:
```golang
push b # notice the last argument is first
push a
call hello
```

Structs are also just laid out on the stack, so if you have a struct:
```golang
type myargs struct {
a str
b int
}
func hello2(c str, args myargs) {}
```

Then when calling it, first you push all the variables for the struct, in reverse order, then the `c str`:
```golang
push b
push a
push c
call hello2
```
### Deferred functions
- Defers do have special handling by the runtime in the stack, but we don't really care - 
  unlike other languages where try..except or try..catch blocks have special meaning in flow control, in Go, defers are **always** function calls, meaning
  to place a breakpoint (hook) on a defer you just hook the function the defer calls.
### More reading about Go internals (through the lens of writing Go assembly)
Recommended - https://github.com/teh-cmc/go-internals/blob/master/chapter1_assembly_primer/README.md
Explains a lot of the above
## How Delve (the Go debugger) works & the Delve architecture
### Learned through the original POC - using the Delve debugger as code to fetch variables
Original POC is at commit cf6f66eb1f0501430342c58666d4607bb6d9fdc0

It's also worth looking at the Delve documentation for writing a client at https://github.com/go-delve/delve/blob/master/Documentation/api/ClientHowto.md

#### Basics
- The Delve debugger is made out of a client and a server. The client is your IDE, or the CLI utility you use to debug.
- The server is the interesting part: it's the component responsible for doing the actual work - placing breakpoints, reading data, and sending it to the client.
- When you start `dlv`, the server is started (in `delve/service/rpc2/server.go`). Your client then instructs the server whether to attach to a running process or execute a process.
- Let's assume we picked attach (as in the original POC). The RPCServer creates a `*service.Debugger` and attaches it to the process.
- The RPCServer uses the `*service.Debugger` with its `CreateBreakpoint` function which takes an `*api.Breakpoint` and uses the `*proc.Target` function `SetBreakpoint` to set the breakpoint). The RPC server creates breakpoints requested by the client using these functions.
- At this point, we can create a breakpoint by adding it hardcoded to `rpccommon\server.go` `Server.Run()` and seeing that the debugger stops.
    ```golang
    _, err = s.debugger.CreateBreakpoint(&api.Breakpoint{File: "C:/work/repos/delve/cmd/rook/rookpkg/test.go", Line: 5, Cond: "", LoadArgs: &api.LoadConfig{FollowPointers: true, MaxVariableRecurse: 1, MaxStringLen: 512, MaxArrayValues: 10, MaxStructFields: -1}, LoadLocals: &api.LoadConfig{FollowPointers: true, MaxVariableRecurse: 1, MaxStringLen: 512, MaxArrayValues: 10, MaxStructFields: -1}})`
    ``` 
#### Initialization and architecture
- Depending on the debugging `backend` and OS, it picks a different implementation for attaching (a different backend for `*proc.Target`).
    ```golang
    // Attach will attach to the process specified by 'pid'.
    func (d *Debugger) Attach(pid int, path string) (*proc.Target, error) {
        switch d.config.Backend {
        case "native":
            return native.Attach(pid, d.config.DebugInfoDirectories)
        case "lldb":
            return betterGdbserialLaunchError(gdbserial.LLDBAttach(pid, path, d.config.DebugInfoDirectories))
        case "default":
            if runtime.GOOS == "darwin" {
                return betterGdbserialLaunchError(gdbserial.LLDBAttach(pid, path, d.config.DebugInfoDirectories))
            }
            return native.Attach(pid, d.config.DebugInfoDirectories)
        default:
            return nil, fmt.Errorf("unknown backend %q", d.config.Backend)
        }
    }
    ```
- Let's look at the implementation for the `native` backend, which uses the Windows debugging API (`DebugActiveProcess`) on Windows or `PTRACE_ATTACH` on Linux.
    ```golang
    // Attach to an existing process with the given PID.
    func Attach(pid int, _ []string) (*proc.Target, error) {
        dbp := newProcess(pid)
        var err error
        dbp.execPtraceFunc(func() {
            // TODO: Probably should have SeDebugPrivilege before starting here.
            err = _DebugActiveProcess(uint32(pid))
        })
        if err != nil {
            return nil, err
        }
        exepath, err := findExePath(pid)
        if err != nil {
            return nil, err
        }
        tgt, err := dbp.initialize(exepath, []string{})
        if err != nil {
            dbp.Detach(true)
            return nil, err
        }
        return tgt, nil
    }
    ```
- We can see here that a "process" object is created (`newProcess` as `dbp`) and the result of `dbp.initialize` is a `*proc.Target`.
- The definition for *proc.Target looks like this:
    ```golang
    // Target represents the process being debugged.
       type Target struct {
        Process
       
        proc ProcessInternal
       
    ...
       }
    ```
- So we see that *proc.Target is a proc.Process, but it also has a proc which is a ProcessInternal.
    ```golang
    type ProcessInternal interface {
        SetCurrentThread(Thread)
        // Restart restarts the recording from the specified position, or from the
        // last checkpoint if pos == "".
        // If pos starts with 'c' it's a checkpoint ID, otherwise it's an event
        // number.
        Restart(pos string) error
        Detach(bool) error
        ContinueOnce() (trapthread Thread, stopReason StopReason, err error)
    
        WriteBreakpoint(addr uint64) (file string, line int, fn *Function, originalData []byte, err error)
        EraseBreakpoint(*Breakpoint) error
    }
    ```
- ProcessInternal looks like it's responsible for the building blocks for the actual debugger's operation:
    - `ContinueOnce` seems to run the debugged process until it stops (we can guess by the name - continue, which acts this way when you use the command in debuggers, and by the return values - trapthread, stopReason)
    - `WriteBreakpoint` seems to take a memory address, writes a breakpoint and returns metadata plus the original data that was replaced. Since this implementation uses the native debugging API, we can assume the write places the `int 3`.
      The fact that the originalData is returned hints that the ProcessInternal implementation does not do any bookkeeping, but that shared Delve code does that.
    - `EraseBreakpoint` seems to remove breakpoints, and takes a *proc.Breakpoint as a parameter.
    - Let's look at the definition for *proc.Breakpoint:
        ```golang
        type Breakpoint struct {
            // File & line information for printing.
            FunctionName string
            File         string
            Line         int
        
            Addr         uint64 // Address breakpoint is set for.
            OriginalData []byte // If software breakpoint, the data we replace with breakpoint instruction.
            Name         string // User defined name of the breakpoint
            LogicalID    int    // ID of the logical breakpoint that owns this physical breakpoint
        
            // ...
            Kind BreakpointKind
        
            // Breakpoint information
            Tracepoint    bool // Tracepoint flag
            TraceReturn   bool
            Goroutine     bool     // Retrieve goroutine information
            Stacktrace    int      // Number of stack frames to retrieve
            Variables     []string // Variables to evaluate
            LoadArgs      *LoadConfig
            LoadLocals    *LoadConfig
            HitCount      map[int]uint64 // Number of times a breakpoint has been reached in a certain goroutine
            TotalHitCount uint64         // Number of times a breakpoint has been reached
       
            // ...
            DeferReturns []uint64
            // Cond: if not nil the breakpoint will be triggered only if evaluating Cond returns true
            Cond ast.Expr
        ```
    - This tells us a few interesting things:
        1. There's a breakpoint structure containing metadata that's maintained by Delve, so we don't need to keep track of breakpoints.
        2. Breakpoints can have a user defined name, which we can use for matching Rook Rule breakpoints with Delve breakpoints.
        3. Delve has configuration for how many stack frames to retrieve, which variables to evaluate (collected variables), and configuration for how to load args and locals (this maps to the Rook collection depth).
        4. It has native support for hit counting.
        5. It supports conditions that are Go AST (meaning conditions come in Go syntax) - not exactly right, but it means it already supports conditions. Cool! 
- Now let's go back to `native.Attach`, and take a look at the `newProcess` implementation. It returns a `*nativeProcess*, which GoLand tells us implements proc.ProcessInternal.
    - The interesting part is it has a *proc.BinaryInfo struct, which seems to be the parsed debug info of the binary. BinaryInfo has cool functions like `LineToPC` which maps filename and lineno to a list of memory addresses - exactly what we would need for a Rook.
    - Let's take a look at the implementation for `WriteBreakpoint`:
        ```golang
        func (dbp *nativeProcess) WriteBreakpoint(addr uint64) (string, int, *proc.Function, []byte, error) {
            f, l, fn := dbp.bi.PCToLine(uint64(addr))
        
            originalData := make([]byte, dbp.bi.Arch.BreakpointSize())
            _, err := dbp.currentThread.ReadMemory(originalData, addr)
            if err != nil {
                return "", 0, nil, nil, err
            }
            if err := dbp.writeSoftwareBreakpoint(dbp.currentThread, addr); err != nil {
                return "", 0, nil, nil, err
            }
        
            return f, l, fn, originalData, nil
        }
      ```
      And `writeSoftwareBreakpoint`:
      ```golang
      func (dbp *nativeProcess) writeSoftwareBreakpoint(thread *nativeThread, addr uint64) error {
      	_, err := thread.WriteMemory(addr, dbp.bi.Arch.BreakpointInstruction())
      	return err
      }
      ```
      We can see it uses a thread, which GoLand tells us satisfies the `proc.Thread` interface:
      ```golang
      type Thread interface {
      	MemoryReadWriter
      	Location() (*Location, error)
      	// Breakpoint will return the breakpoint that this thread is stopped at or
      	// nil if the thread is not stopped at any breakpoint.
      	Breakpoint() *BreakpointState
      	ThreadID() int
      
      	// Registers returns the CPU registers of this thread. The contents of the
      	// variable returned may or may not change to reflect the new CPU status
      	// when the thread is resumed or the registers are changed by calling
      	// SetPC/SetSP/etc.
      	// To insure that the the returned variable won't change call the Copy
      	// method of Registers.
      	Registers() (Registers, error)
      
      	// RestoreRegisters restores saved registers
      	RestoreRegisters(Registers) error
      	BinInfo() *BinaryInfo
      	StepInstruction() error
      	// Blocked returns true if the thread is blocked
      	Blocked() bool
      	// SetCurrentBreakpoint updates the current breakpoint of this thread, if adjustPC is true also checks for breakpoints that were just hit (this should only be passed true after a thread resume)
      	SetCurrentBreakpoint(adjustPC bool) error
      	// Common returns the CommonThread structure for this thread
      	Common() *CommonThread
      
      	SetPC(uint64) error
      	SetSP(uint64) error
      	SetDX(uint64) error
      }
      ```
    The Thread seems to both be responsible for reading and writing memory, and also provides processor registers.
    Presumably, Delve uses the values of the registers to determine the stack address and which values are in scope.
    So `Registers()` together with the Thread's ability to `ReadMemory()` would be all of the things needed to read variables.
- That's it. This is all the info we need to understand which parts we need to replace to switch Delve to using in-process hooking instead of native debugging: `ProcessInternal`, `Thread`, and we also need to replace the RPC Server (so that instead of being controlled through the RPC server, the Rook will just initialize the `*proc.Debugger`)
- Check the code of the DebuggerRulesManager at commit cf6f66eb1f0501430342c58666d4607bb6d9fdc0 to see an example for how breakpoints are placed and `continue` is used to read breakpoint data when a BP is hit.
## How the Go Rook POC works
### At a high level
#### Placing a BP:
   - Locate the function's entry point using Delve
   - If the BP is set on a function without a stack-split prologue or on a line that isn't the first line of the function, error.
   - Replace the beginning of the function (stack-split prologue) with a call to a manually-written assembly function, and pad the rest of the prologue with NOPs.
   - The assembly function extracts the needed info from registers and the stack
     and calls the callback function, written in Go, with this info as parameters.
   - The callback function uses that info to build the OnStackRegisters and InMemoryThread objects that conform to Delve interfaces (proc.Registers and proc.Thread),
     that can then be used by Delve to read local variables from the memory (using the debug symbols) into Delve objects representing the variables.
   - In another goroutine, DelveFrameNamespace (containing DelveObjectNamespace) and DelveStackNamespace are then built from the extracted locals and stack, and protobuf_utils.dumpDelveObject is used to convert
     these objects into *rook.Variants.
#### Getting a user msg:
   - Hook placed in prologue is hit, calls the assembly function
   - Assembly function extracts needed info and calls `callback` function
   - `callback` uses Delve facilities to read local variables, then sends the data to the `BreakpointHitHandler`
   - Returns to user calling code
   - In another goroutine, `BreakpointHitHandler` converts the Delve data to `DelveFrameNamespace` and `DelveStackNamespace` and finally converts those into a `*rook.Variant`.
   - `Output.SendUserMessage` creates a protobuf AugReportMessage and calls `AgentCom.SendNonBlocking`.
   - In another goroutine, `AgentCom` sends the user message to the controller.
#### Components of the Go Rook
- `Singleton` has a:
    - `Debugger` object (modified Delve code), has a:
        - `inprocess.Process` (custom implementation of the Delve proc.Process interface designed to run in-process with hooks) - does the actual placement and removal of breakpoints (hooks)
        - `inprocess.inmemoryThread` (custom implementation of the Delve proc.Thread interface, responsible for reading and writing to memory)
    - `RulesManager` object which implements the types.RulesManager interface (AddRule/RemoveRule), uses the Debugger object to place and remove breakpoints.
    - `AgentCom` based on C2CWsClient that connects to the controller, gets rule JSONs and converts them to type.RuleConfiguration and calls the `RulesManager` with `AddRule/RemoveRule`
    - `Output` (`SendUserMessage`, `SendRuleStatus`, `SendLogMessage`) that uses the `AgentCom` to send data to the controller
    - `BreakpointHitHandler` (sort of an `ActionRunProcessor`/`LocationFileLine` that has a static configuration) - takes the Delve objects and converts them to *rook.Variants in a goroutine separate from the hook, before calling `Output.SendUserMessage`
    
### Go Rook internals
#### inprocess.Process
- The important part is `WriteBreakpoint`:
    ```golang
    func (dbp *InprocessProcess) WriteBreakpoint(addr uint64) (string, int, *proc.Function, []byte, error) {
    	f, l, function := dbp.bi.PCToLine(uint64(addr))
    
    	_, firstLine, _ := dbp.bi.PCToLine(function.Entry)
    	if firstLine != l {
    		return "", 0, nil, nil, errors.Wrap(BpNotFirstLineErr)
    	}
    ```
    Since this POC relies on hooking the prologue, we want to return an error if we're not hooking the first line
    ```golang
    	jumpcode, err := dbp.makeCallqShellcode(function.Entry)
    	if err != nil {
    		return "", 0, nil, nil, errors.Wrap(err)
    	}
    ```
    We then generate the shellcode. This is a 5-byte relative `CALLQ` to `PrepareOnStackRegistersAndCallCallback`, which is a short function written in Go assembler (explained later).
    Since `PrepareOnStackRegistersAndCallCallback` is written in Go assembler and is actually linked into the program, we can assume it will be within 4GB of other code (which is what the 5-byte CALLQ requires).
    
    The hook itself must be short enough to fit within the prologue, so we use Delve's prologue parsing to detect whether there's a prologue and get the length.
    
    ```golang
    	length, hasPrologue := proc.GetPrologueLength(dbp, function)
    	if !hasPrologue {
    		return "", 0, nil, nil, errors.Wrap(BpNoPrologueErr)
    	}
    
    	if length < uint64(len(jumpcode)) {
    		return "", 0, nil, nil, errors.New("prologue is shorter than jumpcode")
    	}
    ```
    
    And replace whatever space is left in the prologue with NOPs (this allows us to return from the hook and continue execution normally).
    ```golang
    	jumpPadding := make([]byte, 0)
    	for i := uint64(0); i < length-uint64(len(jumpcode)); i++ {
    		jumpPadding = append(jumpPadding, 0x90) // nop
    	}
    
    	jumpcode = append(jumpcode, jumpPadding...)
    
    	originalData := make([]byte, length) // 1 extra byte for jmp offset
    	_, err = dbp.memReadWriter.ReadMemory(originalData, function.Entry)
    	if err != nil {
    		return "", 0, nil, nil, err
    	}
    
    ```
    This bit changes the write protection of the function's entry point to get it ready for writing. On Windows, it calls VirtualProtect, and on Linux it calls mprotect. For Mac it's not implemented.
    ```golang
    	err = MakeRwx(uintptr(function.Entry), len(jumpcode))
    	if err != nil {
    		return "", 0, nil, nil, err
    	}
    
    	_, err = dbp.memReadWriter.WriteMemory(function.Entry, jumpcode)
    	if err != nil {
    		return "", 0, nil, nil, err
    	}
    
    	return f, l, function, originalData, nil
    }
  ```
  Finally, it writes the jumpcode.
  
#### inprocess.inmemoryThread
Unlike native debugging, this is mostly still called a thread because it satisfies the thread interface, but it's actually only there to satisfy the ReadMemory, WriteMemory and Registers interface. We don't actually keep track of threads. The inprocess.Process uses a dummy thread to read/write memory in the current process' address space. When a breakpoint is hit, we create a fresh inmemoryThread for it, which represents the current state (registers).

#### The actual hook
The hook flow looks like this:
1. Hooked user code (raw bytes, written to memory)
2. CALLQ by address to `inprocess.PrepareOnStackRegistersAndCallCallback` (written in Go assembler)
3. CALLQ using linker symbol to `inprocess.callback` (written in plain Go)

First thing to keep in mind is that since our hook replaces the stack split prologue, we're relying on the stack split prologue added by the assembler to `inprocess.PrepareOnStackRegistersAndCallCallback` to replace it. But this is not bulletproof - as we know there are multiple kinds of prologues depending on function size, and the assembler function is definitely small.

Let's start from `callback` defined in `callback.go`. This function takes care of hiding any errors, extracting user data while we're in the user context, sending the data to another goroutine for async handling, and returning to user code.
```golang
func callback(stackRegs registers.OnStackRegisters) {
	prevPanicOnFault := debug.SetPanicOnFault(true)
	defer func() {
		if r := recover(); r != nil {
			logger.Logger().Errorf("recovered from panic: %v", r)
		}
		debug.SetPanicOnFault(prevPanicOnFault)
	}()
	th := NewInmemoryThread(dbp, stackRegs)
	// Sets current bp according to RIP value in registers. If BP not found, returns an error.
	err := th.SetCurrentBreakpoint(true)
	if err != nil {
		logger.Logger().Errorf("failed to set current breakpoint, err: %v", err)
		return
	}
	bp := api.ConvertBreakpoint(th.Breakpoint().Breakpoint)
	bpi, err := collectBreakpointInformation(bp, th)
	if err != nil {
		logger.Logger().Errorf("failed to collect breakpoint info, err: %v\n", err)
		return
	}

	err = bpHitHandler.ReportBreakpointHit(bp, bpi)
	if err != nil {
		logger.Logger().Errorf("failed to report breakpoint info, err: %v\n", err)
		return
	}
}
```

We can see that `callback` takes one parameter, a struct `OnStackRegisters`. This struct also "happens" to implement the interface for `proc.Registers`, which is used by Delve to read user data.

This is what `OnStackRegisters` looks like:
```golang
type OnStackRegisters struct {
	// Order is important - these are pushed onto the stack. In order for rsp to be unmodified when it's pushed,
	// it needs to be last (meaning it is pushed first).
	RIP uintptr
	TLSVal uintptr
	RBP    uintptr
	RSP    uintptr
}
```

The TLS value is used on x86/amd64 to get the pointer to the current goroutine -- for us that seems to be useful mostly for telling the goroutine ID. I thought it might be necessary for reading some data or the stack trace, but that doesn't seem to be the case. Delve seems to use it for stack traces and for evaluating code in the context of the user code. I was not able to actually read it, but simply replaced anything that used the goroutine with the thread - e.g. replaced calls to GoroutineScope with ThreadScope.

`PrepareOnStackRegistersAndCallCallback`'s job is to extract those values from the struct and call `callback`. Here's the code for it (minus comments - check the actual code if you have any questions):
```golang
TEXT ·PrepareOnStackRegistersAndCallCallback(SB), $0
NO_LOCAL_POINTERS

// The pushes here align with the struct registers.OnStackRegisters
PUSHQ SP
PUSHQ BP
MOVQ TLS, BX // Depending on architecture, translates into different code that puts the TLS pointer into BX. On x86/amd64, the pointer to the current *G (goroutine) is kept in TLS.
PUSHQ BX
MOVQ returnAddr-8(FP), BX // mov rbx, [rsp-8] - gets the return address, we need the RIP of the caller to find the breakpoint that called us
PUSHQ BX
CALL ·callback(SB)
POPQ BX
POPQ AX
POPQ BX
POPQ BX
RET
```

#### Reading user data and converting to *rook.Variants
`collectBreakpointInformation` was copied from Delve and modified to work without the TLS value, and read stack trace directly from the current stack since we're in-line and not in another process.
Once the callback is done collecting, it just passes the `*api.Breakpoint` (BP metadata) and `*inprocess.BreakpointInfo` (user data) (based on `*api.BreakpointInfo`) into BreakpointHitHandler via a channel,
which then wraps the stack trace in `DelveStackNamespace` and the frame in `DelveFrameNamespace`. `DelveFrameNamespace` wraps each local and argument in `DelveObjectNamespace`.
`DelveObjectNamespace` implements some namespace operations on the object (for example, reading from a slice or a map), and converts some `reflect` types to Rookout types.

Finally, `DelveObjectNamespace`'s `ToProtobuf` calls `protobuf_utils.dumpDelveObject` which takes a Delve `*proc.Variable` and returns a `*rook.Variant`. It should support *all* types, as I went one by one through the Delve code and `reflect` types enum and implemented all of them.

The code for `dumpDelveObject` is based on `*api.Variable`'s `writeTo`, which converts variables into a string representation, but basically tells you how to read the data for each type. I used that info + code from other Rooks to understand how to build the *rook.Variant.
![](rookdata.png)
### What's needed to make this work on every function, on any line and without crashing (TODO)
#### Replace the prologue hooking with a hooking method that can work on every line
- That essentially means resizing functions, so you need to disassemble the function, add the jump code, then reassemble it. Before reassembly, you also need to adjust all relative addresses in the moved code.
- Since the function size changed, you need to copy the function elsewhere, and place a permanent hook to the new location at the original function.
- This is complicated by the stack maps that are used on stack splits and the fact the garbage collector has a list of code sections -- it needs to consider the function being collected a valid function. I did not investigate deeper what the GC does with this info.
- Remember when I mentioned earlier that the linker creates some data structures in the binary? That's where this comes in.
- In runtime/symtab.go, you have this (this is a linked list of moduledata objects):
    ```golang
    var firstmoduledata moduledata  // linker symbol
    var lastmoduledatap *moduledata // linker symbol
    var modulesSlice *[]*moduledata // see activeModules
  ```
- Which among other things, has this: `minpc, maxpc uintptr`. Those are the minimum and maximum code addresses (text sections) - the PC stands for program counter.
- If garbage collection or a stack split runs while there's a return address in the stack that isn't in that range, the function will be considered invalid and the GC will throw a fatalthrow.
    - You'll see this error: `runtime: unexpected return pc for funcname called from 0xaddress`
    - This check happens at `runtime/traceback.go` at `gentraceback()`
      ```golang
			flr = findfunc(frame.lr)
			if !flr.valid() {
				// This happens if you get a profiling interrupt at just the wrong time.
				// In that context it is okay to stop early.
				// But if callback is set, we're doing a garbage collection and must
				// get everything, so crash loudly.
				doPrint := printing
				if doPrint && gp.m.incgo && f.funcID == funcID_sigpanic {
					// We can inject sigpanic
					// calls directly into C code,
					// in which case we'll see a C
					// return PC. Don't complain.
					doPrint = false
				}
				if callback != nil || doPrint {
					print("runtime: unexpected return pc for ", funcname(f), " called from ", hex(frame.lr), "\n")
					tracebackHexdump(gp.stack, &frame, lrPtr)
				}
				if callback != nil {
					throw("unknown caller pc")
				}
			}
      ```   
  - Studying the code in `gentraceback`, there seem to be ways to avoid this, and there's also the obvious way of adding another entry to the `moduledata` linked list, but it isn't obvious what else will break.
  
#### Prevent crashing when placing a bp
If you modify a function's code while that specific bit of code is being executed, because the change isn't atomic (the call is 5-bytes), the processor may start decoding the instruction and then see your hook as it continues, which will result in an Illegal Instruction processor exception.

A possible way to get around this, is on Rook start -- before any user code starts running (this technique in video rendering is called "double buffering"):
- Copy all functions twice: once unmodified, and once to be hooked.
- Modify the original function so that it looks like:
```
jmp short unmodified_label // 2 bytes
hooked_label:
jmp abs hooked
unmodified_label:
jmp abs unmodified
```
- Then, when you place a BP on a function, you first modify the hooked copy, and ONLY then change the original copy so that the short jmp goes to the hooked_label. Because the change is short (and the instruction is the same, just the argument is changed), it will be atomic.

Obvious problems with this are what I mentioned before - the stack maps and the `moduledata`. These copies of the function were not there at link time, so they're not in the module data. Is adding them enough? I dunno. Maybe there's another, easier solution. 
    
#### Distribution without exposing the sources?
Go packages are git repositories, which poses a challenge if we want to keep the GoRook closed source. We could potentially "release" by pushing to a git repo with precompiled binaries, and then load the .so or .dll in a small library. There are possible technical downsides (or upsides) though: by virtue of being a separate binary, it means we're running with a separate Go runtime. But that also means that our functions and stack maps are not known to the user's Go runtime, so if we hook the user's code with a callback that's in our binary, it'll be an unknown caller. So it's worth considering keeping it open source, or perhaps making it a private repository where an access key is given to paying customers.

