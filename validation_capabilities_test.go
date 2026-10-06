package main

import (
	"context"
	"testing"

	basev0 "github.com/codefly-dev/core/generated/go/codefly/base/v0"
	agentv0 "github.com/codefly-dev/core/generated/go/codefly/services/agent/v0"
	builderv0 "github.com/codefly-dev/core/generated/go/codefly/services/builder/v0"
	"github.com/stretchr/testify/require"
)

func TestVaultAdvertisesNoWorkloadValidation(t *testing.T) {
	info, err := NewService().GetAgentInformation(context.Background(), &agentv0.AgentInformationRequest{})
	require.NoError(t, err)
	// Vault's own Go suites validate this agent. A consuming workspace has no
	// user source or test suite for the managed Vault service to execute.
	require.NotNil(t, info.Validation, "an absent contract triggers legacy runtime probing")
	require.False(t, info.Validation.GetLint().GetSupported())
	require.False(t, info.Validation.GetCompile().GetSupported())
	require.False(t, info.Validation.GetTest().GetSupported())
}

func TestBuilderValidationRemainsAvailable(t *testing.T) {
	info, err := NewService().GetAgentInformation(context.Background(), &agentv0.AgentInformationRequest{})
	if err != nil {
		t.Fatal(err)
	}
	validation := info.GetValidation()
	for name, operation := range map[string]*agentv0.ValidationOperationCapability{
		"audit": validation.GetAudit(), "build": validation.GetArtifactBuild(),
		"sbom": validation.GetSbom(), "sync": validation.GetSync(),
	} {
		if !operation.GetSupported() {
			t.Errorf("%s lost its supported operation", name)
		}
	}
	builder := NewBuilder()
	if err := builder.HeadlessLoad(context.Background(), &basev0.ServiceIdentity{Workspace: "workspace", Module: "module", Name: "service"}); err != nil {
		t.Fatal(err)
	}
	response, err := builder.Sync(context.Background(), &builderv0.SyncRequest{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if response.GetState().GetState() != builderv0.SyncStatus_SUCCESS || len(response.GetChangedFiles()) != 0 {
		t.Fatalf("unexpected sync result: %v", response)
	}
}
