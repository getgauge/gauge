/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package lang

import (
	"context"

	gm "github.com/getgauge/gauge-proto/go/gauge_messages"
	"google.golang.org/grpc"
)

type mockClient struct {
	responses map[gm.Message_MessageType]interface{}
	err       error
}

func (r *mockClient) GetStepNames(ctx context.Context, in *gm.StepNamesRequest, opts ...grpc.CallOption) (*gm.StepNamesResponse, error) {
	return r.responses[gm.Message_StepNamesResponse].(*gm.StepNamesResponse), r.err
}
func (r *mockClient) CacheFile(ctx context.Context, in *gm.CacheFileRequest, opts ...grpc.CallOption) (*gm.Empty, error) {
	return &gm.Empty{}, r.err
}
func (r *mockClient) GetStepPositions(ctx context.Context, in *gm.StepPositionsRequest, opts ...grpc.CallOption) (*gm.StepPositionsResponse, error) {
	return r.responses[gm.Message_StepPositionsResponse].(*gm.StepPositionsResponse), r.err
}
func (r *mockClient) GetImplementationFiles(ctx context.Context, in *gm.Empty, opts ...grpc.CallOption) (*gm.ImplementationFileListResponse, error) {
	return r.responses[gm.Message_ImplementationFileListResponse].(*gm.ImplementationFileListResponse), r.err
}
func (r *mockClient) ImplementStub(ctx context.Context, in *gm.StubImplementationCodeRequest, opts ...grpc.CallOption) (*gm.FileDiff, error) {
	return r.responses[gm.Message_FileDiff].(*gm.FileDiff), r.err
}

func (r *mockClient) ValidateStep(ctx context.Context, in *gm.StepValidateRequest, opts ...grpc.CallOption) (*gm.StepValidateResponse, error) {
	return r.responses[gm.Message_StepValidateResponse].(*gm.StepValidateResponse), r.err
}
func (r *mockClient) Refactor(ctx context.Context, in *gm.RefactorRequest, opts ...grpc.CallOption) (*gm.RefactorResponse, error) {
	return r.responses[gm.Message_RefactorResponse].(*gm.RefactorResponse), r.err
}
func (r *mockClient) GetStepName(ctx context.Context, in *gm.StepNameRequest, opts ...grpc.CallOption) (*gm.StepNameResponse, error) {
	return r.responses[gm.Message_StepNameResponse].(*gm.StepNameResponse), r.err
}

func (r *mockClient) GetGlobPatterns(ctx context.Context, in *gm.Empty, opts ...grpc.CallOption) (*gm.ImplementationFileGlobPatternResponse, error) {
	return r.responses[gm.Message_ImplementationFileGlobPatternResponse].(*gm.ImplementationFileGlobPatternResponse), r.err
}

func (r *mockClient) KillProcess(ctx context.Context, in *gm.KillProcessRequest, opts ...grpc.CallOption) (*gm.Empty, error) {
	return nil, nil
}
