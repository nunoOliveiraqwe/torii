package domain

const systemConfigID = 1 // there is only one system configuration

type SystemConfiguration struct {
	ID                        int
	IsFirstTimeSetupConcluded bool
}
