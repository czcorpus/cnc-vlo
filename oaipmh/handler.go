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
	"net/http"
	"net/url"

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
	hook VLOHook
}

func getReqResp(ctx *gin.Context, argSource url.Values) (*OAIPMHRequest, *OAIPMHResponse) {
	req := &OAIPMHRequest{URL: ctx.Request.Host + ctx.Request.URL.Path}
	resp := NewOAIPMHResponse(req)

	// get verb operation
	if !argSource.Has(ArgVerb) {
		resp.Errors.Add(ErrorCodeBadArgument, "Missing required argument `"+ArgVerb+"`")
		return req, resp
	}
	req.Verb = getTypedArg[Verb](argSource, ArgVerb)
	if err := req.Verb.Validate(); err != nil {
		resp.Errors.Add(ErrorCodeBadVerb, "Invalid verb `"+req.Verb.String()+"`")
		return req, resp
	}

	// check required arguments
	if arg := req.Verb.ValidateRequiredArgs(argSource); arg != "" {
		resp.Errors.Add(ErrorCodeBadArgument, "Missing required argument `"+arg+"` for verb "+req.Verb.String())
		return req, resp
	}
	// check allowed arguments
	for k := range argSource {
		if !req.Verb.ValidateArg(k) {
			resp.Errors.Add(ErrorCodeBadArgument, "Invalid argument `"+k+"` for verb "+req.Verb.String())
			return req, resp
		}
	}

	req.Identifier = getTypedArg[string](argSource, ArgIdentifier)
	req.MetadataPrefix = getTypedArg[string](argSource, ArgMetadataPrefix)
	req.From = getTypedArg[string](argSource, ArgFrom)
	req.Until = getTypedArg[string](argSource, ArgUntil)
	req.Set = getTypedArg[string](argSource, ArgSet)
	req.ResumptionToken = getTypedArg[string](argSource, ArgResumptionToken)
	return req, resp
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
		resp.Errors.Add(ErrorCodeBadArgument, "Verb not implemented `"+req.Verb.String()+"`")
		writeXMLResponse(ctx.Writer, http.StatusNotImplemented, resp)
		return
	}
	resp.Errors = append(resp.Errors, errors...)
	writeXMLResponse(ctx.Writer, httpCode, resp)
}

func (a *VLOHandler) HandleOAIGet(ctx *gin.Context) {
	req, resp := getReqResp(ctx, ctx.Request.URL.Query())
	if resp.Errors.HasErrors() {
		writeXMLResponse(ctx.Writer, http.StatusBadRequest, resp)
		return
	}
	a.handleRequest(ctx, req, resp)
}

func (a *VLOHandler) HandleOAIPost(ctx *gin.Context) {
	if err := ctx.Request.ParseForm(); err != nil {
		log.Error().Err(err).Send()
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	req, resp := getReqResp(ctx, ctx.Request.PostForm)
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

func NewVLOHandler(hook VLOHook) *VLOHandler {
	return &VLOHandler{hook: hook}
}
