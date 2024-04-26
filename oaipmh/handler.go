// Copyright 2024 Martin Zimandl <martin.zimandl@gmail.com>
// Copyright 2024 Institute of the Czech National Corpus,
//                Faculty of Arts, Charles University
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package oaipmh

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/czcorpus/cnc-gokit/collections"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type ResultWrapper[T any] struct {
	Data     T
	Errors   OAIPMHErrors
	HTTPCode int
}

func (w *ResultWrapper[any]) NoError() bool {
	return !w.Errors.HasErrors() && w.HTTPCode < 400
}

func NewResultWrapper[T any](data T) ResultWrapper[T] {
	return ResultWrapper[T]{
		Data:     data,
		HTTPCode: http.StatusOK,
	}
}

type VLOHook interface {
	Identify() ResultWrapper[OAIPMHIdentify]
	GetRecord(req OAIPMHRequest) ResultWrapper[OAIPMHRecord]
	ListIdentifiers(req OAIPMHRequest) ResultWrapper[[]OAIPMHRecordHeader]
	ListMetadataFormats(req OAIPMHRequest) ResultWrapper[[]OAIPMHMetadataFormat]
	ListRecords(req OAIPMHRequest) ResultWrapper[[]OAIPMHRecord]
	ListSets(req OAIPMHRequest) ResultWrapper[[]OAIPMHSet]

	SupportsSets() bool
	SupportedMetadataPrefixes() []string
}

type VLOHandler struct {
	basePath string
	hook     VLOHook
}

func (a *VLOHandler) getReqResp(argSource url.Values) (*OAIPMHRequest, *OAIPMHResponse, error) {
	OAIURL, err := url.JoinPath(a.basePath, "oai")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to prepare OAIPMH request and response: %w", err)
	}
	req := &OAIPMHRequest{URL: OAIURL}
	resp := NewOAIPMHResponse(req)

	// get verb operation
	if !argSource.Has(ArgVerb) {
		resp.Errors.Add(ErrorCodeBadArgument, fmt.Sprintf("Missing required argument `%s`", ArgVerb))
		return req, resp, nil
	}
	req.Verb = getTypedArg[Verb](argSource, ArgVerb)
	if err := req.Verb.Validate(); err != nil {
		resp.Errors.Add(ErrorCodeBadVerb, fmt.Sprintf("Invalid verb `%s`", req.Verb))
		return req, resp, nil
	}

	// check required arguments
	if arg := req.Verb.ValidateRequiredArgs(argSource); arg != "" {
		resp.Errors.Add(ErrorCodeBadArgument, fmt.Sprintf("Missing required argument `%s` for verb `%s`", arg, req.Verb))
		return req, resp, nil
	}
	// check allowed arguments
	for k := range argSource {
		if !req.Verb.ValidateArg(k) {
			resp.Errors.Add(ErrorCodeBadArgument, fmt.Sprintf("Invalid argument `%s` for verb `%s`", k, req.Verb))
			return req, resp, nil
		}
	}

	req.Identifier = getTypedArg[string](argSource, ArgIdentifier)
	req.MetadataPrefix = getTypedArg[string](argSource, ArgMetadataPrefix)
	if from := getTypedArg[string](argSource, ArgFrom); from != "" {
		var parsed time.Time
		if strings.Contains(from, "T") {
			parsed, err = time.Parse(time.RFC3339, from)
		} else {
			parsed, err = time.Parse(time.DateOnly, from)
		}
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse `from`: %w", err)
		}
		parsed = parsed.In(time.UTC)
		req.From = &parsed
	}
	if until := getTypedArg[string](argSource, ArgUntil); until != "" {
		var parsed time.Time
		if strings.Contains(until, "T") {
			parsed, err = time.Parse(time.RFC3339, until)
		} else {
			parsed, err = time.Parse(time.DateOnly, until)
			parsed = parsed.Add(24 * time.Hour)
		}
		if err != nil {
			return nil, nil, fmt.Errorf("failed to until `from`: %w", err)
		}
		parsed = parsed.In(time.UTC)
		req.Until = &parsed
	}
	req.Set = getTypedArg[string](argSource, ArgSet)
	req.ResumptionToken = getTypedArg[string](argSource, ArgResumptionToken)
	return req, resp, nil
}

func (a *VLOHandler) handleRequest(ctx *gin.Context, req *OAIPMHRequest, resp *OAIPMHResponse) {
	var errors OAIPMHErrors
	httpCode := http.StatusOK
	switch req.Verb {
	case VerbIdentify:
		ans := a.hook.Identify()
		errors, httpCode = ans.Errors, ans.HTTPCode
		if ans.NoError() {
			resp.Identify = &ans.Data
			resp.Identify.BaseURL = req.URL
			resp.Identify.ProtocolVersion = resp.ProtocolVersion
		}

	case VerbGetRecord:
		if !collections.SliceContains(a.hook.SupportedMetadataPrefixes(), req.MetadataPrefix) {
			resp.Errors.Add(ErrorCodeCannotDisseminateFormat, "Unknown metadata format")
			writeXMLResponse(ctx.Writer, http.StatusBadRequest, resp)
			return
		}
		ans := a.hook.GetRecord(*req)
		errors, httpCode = ans.Errors, ans.HTTPCode
		if ans.NoError() {
			resp.GetRecord = &ans.Data
		}

	case VerbListIdentifiers:
		if !collections.SliceContains(a.hook.SupportedMetadataPrefixes(), req.MetadataPrefix) {
			resp.Errors.Add(ErrorCodeCannotDisseminateFormat, "Unknown metadata format")
			writeXMLResponse(ctx.Writer, http.StatusBadRequest, resp)
			return
		}
		if req.Set != "" && !a.hook.SupportsSets() {
			resp.Errors.Add(ErrorCodeNoSetHierarchy, "Sets functionality not implemented")
			writeXMLResponse(ctx.Writer, http.StatusNotImplemented, resp)
			return
		}
		ans := a.hook.ListIdentifiers(*req)
		errors, httpCode = ans.Errors, ans.HTTPCode
		if ans.NoError() {
			resp.ListIdentifiers = &ans.Data
		}

	case VerbListMetadataFormats:
		ans := a.hook.ListMetadataFormats(*req)
		errors, httpCode = ans.Errors, ans.HTTPCode
		if ans.NoError() {
			resp.ListMetadataFormats = &ans.Data
		}

	case VerbListRecords:
		if !collections.SliceContains(a.hook.SupportedMetadataPrefixes(), req.MetadataPrefix) {
			resp.Errors.Add(ErrorCodeCannotDisseminateFormat, "Unknown metadata format")
			writeXMLResponse(ctx.Writer, http.StatusBadRequest, resp)
			return
		}
		if req.Set != "" && !a.hook.SupportsSets() {
			resp.Errors.Add(ErrorCodeNoSetHierarchy, "Sets functionality not implemented")
			writeXMLResponse(ctx.Writer, http.StatusNotImplemented, resp)
			return
		}
		ans := a.hook.ListRecords(*req)
		errors, httpCode = ans.Errors, ans.HTTPCode
		if ans.NoError() {
			resp.ListRecords = &ans.Data
		}

	case VerbListSets:
		if !a.hook.SupportsSets() {
			resp.Errors.Add(ErrorCodeNoSetHierarchy, "Sets functionality not implemented")
			writeXMLResponse(ctx.Writer, http.StatusNotImplemented, resp)
			return
		}
		ans := a.hook.ListSets(*req)
		errors, httpCode = ans.Errors, ans.HTTPCode
		if ans.NoError() {
			resp.ListSets = &ans.Data
		}

	default:
		resp.Errors.Add(ErrorCodeBadArgument, fmt.Sprintf("Verb not implemented `%s`", req.Verb))
		httpCode = http.StatusNotImplemented
	}

	resp.Errors = append(resp.Errors, errors...)
	if httpCode >= 400 && !resp.Errors.HasErrors() {
		ctx.AbortWithStatus(httpCode)
		return
	}
	writeXMLResponse(ctx.Writer, httpCode, resp)
}

func (a *VLOHandler) HandleOAIGet(ctx *gin.Context) {
	req, resp, err := a.getReqResp(ctx.Request.URL.Query())
	if err != nil {
		log.Error().Err(err).Msg("Failed to handle OAIPMH Get request")
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if resp.Errors.HasErrors() {
		writeXMLResponse(ctx.Writer, http.StatusBadRequest, resp)
		return
	}
	a.handleRequest(ctx, req, resp)
}

func (a *VLOHandler) HandleOAIPost(ctx *gin.Context) {
	if err := ctx.Request.ParseForm(); err != nil {
		log.Error().Err(err).Msg("Failed to handle OAIPMH Post request")
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	req, resp, err := a.getReqResp(ctx.Request.PostForm)
	if err != nil {
		log.Error().Err(err).Msg("Failed to handle OAIPMH Post request")
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if resp.Errors.HasErrors() {
		writeXMLResponse(ctx.Writer, http.StatusBadRequest, resp)
		return
	}
	a.handleRequest(ctx, req, resp)
}

func (a *VLOHandler) HandleSelfLink(ctx *gin.Context) {
	req := OAIPMHRequest{
		URL:            ctx.Request.Host + ctx.Request.URL.Path,
		Identifier:     ctx.Param("recordId"),
		MetadataPrefix: ctx.DefaultQuery("format", "oai_dc"),
	}

	ans := a.hook.GetRecord(req)
	if ans.HTTPCode >= 400 {
		ctx.AbortWithStatus(ans.HTTPCode)
	} else {
		writeXMLResponse(ctx.Writer, ans.HTTPCode, ans.Data.Metadata.Value)
	}
}

func NewVLOHandler(basePath string, hook VLOHook) *VLOHandler {
	return &VLOHandler{
		basePath: basePath,
		hook:     hook,
	}
}
