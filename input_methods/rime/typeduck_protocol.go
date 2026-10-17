package rime

import (
	"strings"

	"github.com/gaboolic/moqi-ime/imecore"
)

const typeDuckBackendName = "TypeDuck Rime"

func defaultTypeDuckCapabilities() []imecore.TypeDuckEngineCapability {
	names := []string{
		"pageSize",
		"enableCompletion",
		"enableCorrection",
		"enableSentence",
		"enableLearning",
		"isCangjie5",
		"showReverseCode",
		"displayLanguages",
		"mainLanguage",
		"isHeiTypeface",
		"showRomanization",
	}
	out := make([]imecore.TypeDuckEngineCapability, 0, len(names))
	for _, name := range names {
		out = append(out, imecore.TypeDuckEngineCapability{
			Name:      name,
			Supported: true,
		})
	}
	return out
}

func typeDuckSettingsSnapshotFromPreferences(message string) *imecore.TypeDuckSettingsSnapshot {
	prefs, _ := loadTypeDuckRimePreferences()
	if strings.TrimSpace(message) == "" {
		message = "ok"
	}
	return &imecore.TypeDuckSettingsSnapshot{
		Language:           "zh-HK",
		DisplayLanguage:    "eng",
		TraditionalHongKong: true,
		CandidatePageSize:  uint32(prefs.PageSize),
		MainLanguage:       "eng",
		PageSize:           uint32(prefs.PageSize),
		EnableCompletion:   prefs.EnableCompletion,
		EnableCorrection:   prefs.EnableCorrection,
		EnableSentence:     prefs.EnableSentence,
		EnableLearning:     prefs.EnableLearning,
		IsCangjie5:         prefs.IsCangjie5,
		Source:             typeDuckPreferencesFileName,
		StatusMessage:      message,
	}
}

func (ime *IME) typeDuckSettingsSnapshot(req *imecore.Request, resp *imecore.Response) *imecore.Response {
	snapshot := typeDuckSettingsSnapshotFromPreferences("ok")
	if ime != nil && ime.backend != nil {
		if schemaID := strings.TrimSpace(ime.backend.CurrentSchemaID()); schemaID != "" {
			snapshot.SchemaID = schemaID
		}
	}
	resp.TypeDuckSettingsSnapshot = snapshot
	resp.TypeDuckCapabilities = defaultTypeDuckCapabilities()
	resp.TypeDuckEngineHealth = &imecore.TypeDuckEngineHealth{
		Status:      int32(1),
		BackendName: typeDuckBackendName,
		Message:     "ok",
		Recoverable: true,
	}
	resp.ReturnValue = 1
	return resp
}

func (ime *IME) typeDuckHealth(req *imecore.Request, resp *imecore.Response) *imecore.Response {
	resp.TypeDuckEngineHealth = &imecore.TypeDuckEngineHealth{
		Status:      int32(1),
		BackendName: typeDuckBackendName,
		Message:     "ok",
		Recoverable: true,
	}
	resp.ReturnValue = 1
	return resp
}

func (ime *IME) typeDuckReconfigure(req *imecore.Request, resp *imecore.Response) *imecore.Response {
	reconfigure := req.TypeDuckReconfigure
	if reconfigure == nil {
		resp.Success = false
		resp.Error = "reconfigure payload missing"
		resp.TypeDuckError = &imecore.TypeDuckError{
			Code:        int32(3),
			Message:     resp.Error,
			Recoverable: true,
		}
		return resp
	}
	if schemaID := strings.TrimSpace(reconfigure.SchemaID); schemaID != "" && ime.backend != nil {
		if !ime.backend.SelectSchema(schemaID) {
			resp.Success = false
			resp.Error = "schema selection failed"
			return resp
		}
	}
	if reconfigure.ReloadUserDict || reconfigure.RestartBackend {
		if !ime.redeployTypeDuckSettings(req, resp) {
			resp.Success = false
			if resp.Error == "" {
				resp.Error = typeDuckSettingsApplyFailure
			}
			return resp
		}
	}
	resp.ReturnValue = 1
	resp.TypeDuckEngineHealth = &imecore.TypeDuckEngineHealth{
		Status:      int32(1),
		BackendName: typeDuckBackendName,
		Message:     "reconfigured",
		Recoverable: true,
	}
	return resp
}

func TypeDuckSettingsSnapshotFromLauncher(req *imecore.Request, clientID string) *imecore.Response {
	resp := imecore.NewResponse(req.SeqNum, true)
	ime := &IME{}
	resp = ime.typeDuckSettingsSnapshot(req, resp)
	return resp
}

func TypeDuckHealthFromLauncher(req *imecore.Request) *imecore.Response {
	resp := imecore.NewResponse(req.SeqNum, true)
	ime := &IME{}
	return ime.typeDuckHealth(req, resp)
}

func TypeDuckReconfigureFromLauncher(req *imecore.Request, clientID string) *imecore.Response {
	if clientID == "" {
		clientID = "launcher"
	}
	client := &imecore.Client{ID: clientID, GUID: "typeduck-reconfigure"}
	service := New(client)
	ime, ok := service.(*IME)
	if !ok {
		return &imecore.Response{SeqNum: req.SeqNum, Success: false, Error: typeDuckSettingsApplyFailure}
	}
	client.Service = ime
	if !ime.Init(&imecore.Request{
		Method:          "init",
		SeqNum:          req.SeqNum,
		IsWindows8Above: true,
	}) {
		return &imecore.Response{SeqNum: req.SeqNum, Success: false, Error: "TypeDuck backend could not start for reconfigure"}
	}
	defer ime.Close()
	return ime.HandleRequest(req)
}
