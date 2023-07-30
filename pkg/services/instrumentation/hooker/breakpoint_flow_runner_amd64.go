package hooker

func (c *breakpointFlowRunner) ApplyBreakpointsState() (err error) {
	
	if !c.IsUnhookedState() {
		c.jumpDestination, err = c.nativeAPI.GetStateEntryAddr(c.function.Entry, c.function.End, c.stateID)
		if err != nil {
			return err
		}
	}
	return c.installHook()
}
