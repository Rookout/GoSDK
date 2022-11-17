package types

type AugId = string

type AugConfiguration map[string]interface{}

type AugStatus = string

const (
	Pending AugStatus = "Pending"
	Active  AugStatus = "Active"
	Warning AugStatus = "Warning"
	Error   AugStatus = "Error"
	Deleted AugStatus = "Deleted"
)
