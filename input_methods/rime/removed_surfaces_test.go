package rime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gaboolic/moqi-ime/imecore"
)

func TestTypeDuckV1MenuOmitsRemovedBackendSurfaces(t *testing.T) {
	ime := newIsolatedTestIME(t)
	items := flattenMenuItemsForRemovedSurfaceTest(ime.buildMenu())

	bannedText := []string{
		"下载方案集(&D)",
		"更新配置(&P)",
		"打开置顶短语",
		"自动插入成对符号",
		"打开成对符号设置",
		"WebDAV 配置...",
		"测试 WebDAV",
		"启用云剪贴板",
		"云剪贴板设置...",
		"外觀(&A) / Appearance",
		"切換主題 / Switch Theme",
		"行內預編輯 / Inline Preedit",
		"候選排列 / Candidate Layout",
		"每行候選數 / Candidates per Row",
		"候選間距 / Candidate Spacing",
		"候選總數 / Candidate Count",
		"字體大小 / Font Size",
		"候選文字字體 / Candidate Font",
		"註釋文字大小 / Comment Font Size",
		"註釋文字字體 / Comment Font",
		"候選框背景 / Candidate Background",
		"高亮顏色 / Highlight Color",
		"字體顏色 / Text Color",
		"高亮文字顏色 / Highlight Text Color",
		"匯入外觀 / Import Skin",
	}
	for _, text := range bannedText {
		if _, ok := items.byText[text]; ok {
			t.Fatalf("removed v1 menu surface %q is still visible", text)
		}
	}

	bannedIDs := []int{
		ID_DOWNLOAD_SCHEME_SET,
		ID_UPDATE_CONFIG,
		ID_OPEN_CUSTOM_PHRASE,
		ID_INPUT_AUTO_PAIR_QUOTES,
		ID_OPEN_AUTO_PAIR_SYMBOLS,
		ID_WEBDAV_SETTINGS,
		ID_CLOUD_CLIPBOARD_TEST,
		ID_CLOUD_CLIPBOARD_ENABLED,
		ID_CLOUD_CLIPBOARD_SETTINGS,
	}
	bannedIDs = append(bannedIDs, legacyAppearanceCommandIDsForRemovedSurfaceTest()...)
	for _, id := range bannedIDs {
		if _, ok := items.byID[id]; ok {
			t.Fatalf("removed v1 menu command id %d is still visible", id)
		}
	}
}

func TestTypeDuckV1RemovedBackendCommandsAreUnhandled(t *testing.T) {
	ime := newIsolatedTestIME(t)

	bannedIDs := []int{
		ID_DOWNLOAD_SCHEME_SET,
		ID_UPDATE_CONFIG,
		ID_OPEN_CUSTOM_PHRASE,
		ID_INPUT_AUTO_PAIR_QUOTES,
		ID_OPEN_AUTO_PAIR_SYMBOLS,
		ID_WEBDAV_SETTINGS,
		ID_CLOUD_CLIPBOARD_TEST,
		ID_CLOUD_CLIPBOARD_ENABLED,
		ID_CLOUD_CLIPBOARD_SETTINGS,
	}
	bannedIDs = append(bannedIDs, legacyAppearanceCommandIDsForRemovedSurfaceTest()...)
	for _, id := range bannedIDs {
		resp := ime.onCommand(&imecore.Request{
			SeqNum: 9000 + id,
			ID:     imecore.FlexibleID{Int: id, IsInt: true},
		}, imecore.NewResponse(9000+id, true))
		if resp.ReturnValue != 0 {
			t.Fatalf("removed v1 command id %d returned %d", id, resp.ReturnValue)
		}
	}
}

func TestTypeDuckV1CustomizeUIIsStale(t *testing.T) {
	ime := newIsolatedTestIME(t)
	customizeUI := ime.customizeUIMap()

	if len(customizeUI) != 0 {
		t.Fatalf("CustomizeUI should not be sent in TypeDuck v1, got %#v", customizeUI)
	}
	if _, ok := customizeUI["autoPairQuotes"]; ok {
		t.Fatal("autoPairQuotes should not be sent in TypeDuck v1 CustomizeUI")
	}
	if _, ok := customizeUI["autoPairRules"]; ok {
		t.Fatal("autoPairRules should not be sent in TypeDuck v1 CustomizeUI")
	}
}

func TestTypeDuckV1ResponsesDoNotAttachCustomizeUI(t *testing.T) {
	oldCustomize := customizeTypeDuckSettingsFunc
	oldReloadTypeDuckSettingsFunc := reloadTypeDuckSettingsFunc
	customizeTypeDuckSettingsFunc = func(prefs typeDuckRimePreferences) bool {
		return true
	}
	reloadTypeDuckSettingsFunc = func(sharedDir, userDir, appname, appver string) bool {
		return true
	}
	defer func() {
		customizeTypeDuckSettingsFunc = oldCustomize
		reloadTypeDuckSettingsFunc = oldReloadTypeDuckSettingsFunc
	}()

	ime := newIsolatedTestIME(t)

	for _, req := range []*imecore.Request{
		{
			SeqNum: 9200,
			Method: "onActivate",
		},
		{
			SeqNum: 9201,
			Method: "typeduckSettingsUpdate",
			TypeDuckSettings: &imecore.TypeDuckSettingsUpdate{
				PageSize:         6,
				EnableCompletion: true,
				EnableCorrection: false,
				EnableSentence:   true,
				EnableLearning:   true,
				IsCangjie5:       true,
			},
		},
	} {
		resp := ime.HandleRequest(req)
		if len(resp.CustomizeUI) != 0 {
			t.Fatalf("%s attached stale CustomizeUI payload: %#v", req.Method, resp.CustomizeUI)
		}
	}
}

func TestTypeDuckV1BuildDoesNotPackageLegacyAppearanceThemes(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "scripts", "build.ps1"))
	if err != nil {
		t.Fatal(err)
	}
	script := string(data)
	if strings.Contains(script, `$runtimeFiles = @("appearance_themes.json", "rime.dll")`) {
		t.Fatal("build.ps1 still packages legacy appearance_themes.json as a required runtime file")
	}
	if !strings.Contains(script, `$runtimeFiles = @("rime.dll")`) {
		t.Fatal("build.ps1 should package rime.dll without requiring legacy appearance_themes.json")
	}
	if !strings.Contains(script, `Remove-IfExists -Path (Join-Path $PackageRimeDir "appearance_themes.json")`) {
		t.Fatal("build.ps1 should prune stale packaged appearance_themes.json from previous builds")
	}
}

func TestTypeDuckV1OnMenuDoesNotExposeLegacyBackendMenu(t *testing.T) {
	ime := newIsolatedTestIME(t)

	for _, id := range []string{"settings", "windows-mode-icon", "other"} {
		resp := ime.onMenu(&imecore.Request{
			SeqNum: 9100,
			ID:     imecore.FlexibleID{String: id},
		}, imecore.NewResponse(9100, true))
		if resp.ReturnValue != 0 {
			t.Fatalf("onMenu %q should be unreachable in TypeDuck v1, got ReturnValue=%d", id, resp.ReturnValue)
		}
		items, ok := resp.ReturnData.([]map[string]interface{})
		if !ok {
			t.Fatalf("onMenu %q returned unexpected payload type %#v", id, resp.ReturnData)
		}
		if len(items) != 0 {
			t.Fatalf("onMenu %q exposed legacy backend menu items: %#v", id, items)
		}
	}
}

func legacyAppearanceCommandIDsForRemovedSurfaceTest() []int {
	return []int{
		ID_APPEARANCE_INLINE_PREEDIT,
		ID_APPEARANCE_FONT_14,
		ID_APPEARANCE_FONT_16,
		ID_APPEARANCE_FONT_18,
		ID_APPEARANCE_FONT_20,
		ID_APPEARANCE_FONT_22,
		ID_APPEARANCE_FONT_24,
		ID_APPEARANCE_FONT_26,
		ID_APPEARANCE_FONT_28,
		ID_APPEARANCE_FONT_30,
		ID_APPEARANCE_COMMENT_FONT_14,
		ID_APPEARANCE_COMMENT_FONT_16,
		ID_APPEARANCE_COMMENT_FONT_18,
		ID_APPEARANCE_COMMENT_FONT_20,
		ID_APPEARANCE_COMMENT_FONT_22,
		ID_APPEARANCE_COMMENT_FONT_24,
		ID_APPEARANCE_COMMENT_FONT_26,
		ID_APPEARANCE_COMMENT_FONT_28,
		ID_APPEARANCE_COMMENT_FONT_30,
		ID_APPEARANCE_BG_WHITE,
		ID_APPEARANCE_BG_WARM,
		ID_APPEARANCE_BG_BLUE,
		ID_APPEARANCE_HL_BLUE,
		ID_APPEARANCE_HL_GRAY,
		ID_APPEARANCE_HL_GREEN,
		ID_APPEARANCE_TEXT_BLACK,
		ID_APPEARANCE_TEXT_DARKGRAY,
		ID_APPEARANCE_TEXT_BLUE,
		ID_APPEARANCE_HLTEXT_BLACK,
		ID_APPEARANCE_HLTEXT_WHITE,
		ID_APPEARANCE_HLTEXT_BLUE,
		ID_APPEARANCE_THEME_BASE,
		ID_APPEARANCE_FONT_FAMILY_SEGOE_UI,
		ID_APPEARANCE_FONT_FAMILY_YAHEI_UI,
		ID_APPEARANCE_FONT_FAMILY_DENGXIAN,
		ID_APPEARANCE_FONT_FAMILY_SIMSUN,
		ID_APPEARANCE_COMMENT_FONT_FAMILY_CONSOLAS,
		ID_APPEARANCE_COMMENT_FONT_FAMILY_YAHEI_UI,
		ID_APPEARANCE_COMMENT_FONT_FAMILY_DENGXIAN,
		ID_APPEARANCE_COMMENT_FONT_FAMILY_SIMSUN,
		ID_APPEARANCE_LAYOUT_VERTICAL,
		ID_APPEARANCE_LAYOUT_HORIZONTAL,
		ID_APPEARANCE_PER_ROW_3,
		ID_APPEARANCE_PER_ROW_5,
		ID_APPEARANCE_PER_ROW_7,
		ID_APPEARANCE_PER_ROW_9,
		ID_APPEARANCE_SPACING_0,
		ID_APPEARANCE_SPACING_10,
		ID_APPEARANCE_SPACING_20,
		ID_APPEARANCE_SPACING_30,
		ID_APPEARANCE_SPACING_40,
		ID_APPEARANCE_SPACING_50,
		ID_APPEARANCE_CAND_COUNT_3,
		ID_APPEARANCE_CAND_COUNT_5,
		ID_APPEARANCE_CAND_COUNT_7,
		ID_APPEARANCE_CAND_COUNT_9,
		ID_APPEARANCE_IMPORT_SKIN,
	}
}

type removedSurfaceMenuIndex struct {
	byText map[string]struct{}
	byID   map[int]struct{}
}

func flattenMenuItemsForRemovedSurfaceTest(items []map[string]interface{}) removedSurfaceMenuIndex {
	index := removedSurfaceMenuIndex{
		byText: map[string]struct{}{},
		byID:   map[int]struct{}{},
	}
	var walk func([]map[string]interface{})
	walk = func(entries []map[string]interface{}) {
		for _, entry := range entries {
			if text, ok := entry["text"].(string); ok {
				index.byText[text] = struct{}{}
			}
			if id, ok := entry["id"].(int); ok {
				index.byID[id] = struct{}{}
			}
			if submenu, ok := entry["submenu"].([]map[string]interface{}); ok {
				walk(submenu)
			}
		}
	}
	walk(items)
	return index
}
