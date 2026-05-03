package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/fpigeonjr/music-for-coding-tui/internal/player"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

// runCmd executes a tea.Cmd synchronously and returns the resulting Msg.
// Safe to call on simple func()-returning cmds; do NOT call on tick-based cmds.
func runCmd(t *testing.T, cmd tea.Cmd) tea.Msg {
	t.Helper()
	if cmd == nil {
		return nil
	}
	return cmd()
}

// findResultMsg inspects a Cmd (possibly a tea.Batch) for a commandResultMsg
// without calling time-based sub-commands.
func findResultMsg(t *testing.T, cmd tea.Cmd) (commandResultMsg, bool) {
	t.Helper()
	if cmd == nil {
		return commandResultMsg{}, false
	}
	msg := cmd()
	if r, ok := msg.(commandResultMsg); ok {
		return r, true
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			if c == nil {
				continue
			}
			inner := c()
			if r, ok := inner.(commandResultMsg); ok {
				return r, true
			}
		}
	}
	return commandResultMsg{}, false
}

// ─── Command mode: entry / exit ──────────────────────────────────────────────

func TestUpdate_SlashEntersCommandMode(t *testing.T) {
	m := modelWithEpisodes()
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	fm := result.(model)
	if !fm.commandMode {
		t.Error("expected commandMode=true after /")
	}
}

func TestUpdate_EscExitsCommandMode(t *testing.T) {
	m := modelWithEpisodes()
	m.commandMode = true
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	fm := result.(model)
	if fm.commandMode {
		t.Error("expected commandMode=false after esc")
	}
}

func TestUpdate_EscClearsInputAndState(t *testing.T) {
	m := modelWithEpisodes()
	m.commandMode = true
	m.commandOutput = "old output"
	m.commandError = "old error"
	m.commandResult = true
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	fm := result.(model)
	if fm.commandOutput != "" || fm.commandError != "" || fm.commandResult {
		t.Errorf("esc should clear output=%q error=%q result=%v",
			fm.commandOutput, fm.commandError, fm.commandResult)
	}
}

func TestUpdate_CommandModeBlocksQuit(t *testing.T) {
	m := modelWithEpisodes()
	m.commandMode = true
	// 'q' in command mode should go to textinput, not quit
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd != nil {
		if _, isQuit := cmd().(tea.QuitMsg); isQuit {
			t.Error("q in commandMode should not quit")
		}
	}
}

func TestUpdate_EnterClosesCommandMode(t *testing.T) {
	m := modelWithEpisodes()
	m.commandMode = true
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	fm := result.(model)
	if fm.commandMode {
		t.Error("expected commandMode=false after enter")
	}
}

func TestUpdate_CommandResultMsg_SetsResultState(t *testing.T) {
	m := modelWithEpisodes()
	result, cmd := m.Update(commandResultMsg{output: "vol: 80%"})
	fm := result.(model)
	if !fm.commandResult {
		t.Error("expected commandResult=true after commandResultMsg")
	}
	if fm.commandOutput != "vol: 80%" {
		t.Errorf("commandOutput = %q, want %q", fm.commandOutput, "vol: 80%")
	}
	if cmd == nil {
		t.Error("expected clearCommandBarCmd to be scheduled")
	}
}

func TestUpdate_CommandErrMsg_SetsResultState(t *testing.T) {
	m := modelWithEpisodes()
	result, cmd := m.Update(commandErrMsg{err: errMpvNotFound})
	fm := result.(model)
	if !fm.commandResult {
		t.Error("expected commandResult=true after commandErrMsg")
	}
	if !strings.Contains(fm.commandError, "mpv") {
		t.Errorf("commandError = %q, want mpv error", fm.commandError)
	}
	if cmd == nil {
		t.Error("expected clearCommandBarCmd to be scheduled")
	}
}

func TestUpdate_ClearCommandBarMsg_ClosesBar(t *testing.T) {
	m := modelWithEpisodes()
	m.commandResult = true
	m.commandOutput = "some output"
	m.commandError = "some error"
	result, _ := m.Update(clearCommandBarMsg{})
	fm := result.(model)
	if fm.commandResult {
		t.Error("expected commandResult=false after clearCommandBarMsg")
	}
	if fm.commandOutput != "" || fm.commandError != "" {
		t.Errorf("expected cleared output/error, got output=%q error=%q",
			fm.commandOutput, fm.commandError)
	}
}

func TestUpdate_EscClearsResultBar(t *testing.T) {
	m := modelWithEpisodes()
	m.commandMode = false
	m.commandResult = true
	m.commandOutput = "some result"
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	fm := result.(model)
	if fm.commandResult {
		t.Error("expected commandResult=false after esc on result bar")
	}
}

func TestUpdate_SlashClearsPreviousResult(t *testing.T) {
	m := modelWithEpisodes()
	m.commandResult = true
	m.commandOutput = "old output"
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	fm := result.(model)
	if fm.commandResult {
		t.Error("expected commandResult=false after re-opening command bar")
	}
	if fm.commandOutput != "" {
		t.Errorf("expected empty commandOutput after /, got %q", fm.commandOutput)
	}
}

// ─── executeCommand dispatch ──────────────────────────────────────────────────

func TestExecuteCommand_Empty(t *testing.T) {
	m := modelWithEpisodes()
	cmd := m.executeCommand("   ")
	if cmd != nil {
		t.Error("expected nil cmd for empty input")
	}
}

func TestExecuteCommand_Unknown(t *testing.T) {
	m := modelWithEpisodes()
	msg := runCmd(t, m.executeCommand("foobar"))
	errMsg, ok := msg.(commandErrMsg)
	if !ok {
		t.Fatalf("expected commandErrMsg, got %T", msg)
	}
	if errMsg.err == nil || !strings.Contains(errMsg.err.Error(), "foobar") {
		t.Errorf("expected error mentioning 'foobar', got %v", errMsg.err)
	}
}

func TestExecuteCommand_Quit(t *testing.T) {
	m := modelWithEpisodes()
	cmd := m.executeCommand("quit")
	if cmd == nil {
		t.Fatal("expected non-nil cmd for quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg from quit command")
	}
}

func TestExecuteCommand_ExitAlias(t *testing.T) {
	m := modelWithEpisodes()
	cmd := m.executeCommand("exit")
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg from exit alias")
	}
}

// ─── cmdTheme ────────────────────────────────────────────────────────────────

func TestCmdTheme_NoArgs(t *testing.T) {
	m := modelWithEpisodes()
	msg := runCmd(t, m.cmdTheme(nil))
	errMsg, ok := msg.(commandErrMsg)
	if !ok || errMsg.err == nil {
		t.Errorf("expected error for no args, got %T", msg)
	}
}

func TestCmdTheme_ExactMatch(t *testing.T) {
	m := modelWithEpisodes()
	m.theme = ThemeNord
	m.cmdTheme([]string{"dracula"})
	if m.theme.Name != "Dracula" {
		t.Errorf("theme = %q, want Dracula", m.theme.Name)
	}
	if m.themeMsg != "Dracula" {
		t.Errorf("themeMsg = %q, want Dracula", m.themeMsg)
	}
}

func TestCmdTheme_NormalisesHyphens(t *testing.T) {
	m := modelWithEpisodes()
	m.cmdTheme([]string{"gruvbox-dark"})
	if m.theme.Name != "Gruvbox Dark" {
		t.Errorf("theme = %q, want Gruvbox Dark", m.theme.Name)
	}
}

func TestCmdTheme_NormalisesUnderscores(t *testing.T) {
	m := modelWithEpisodes()
	m.cmdTheme([]string{"gruvbox_dark"})
	if m.theme.Name != "Gruvbox Dark" {
		t.Errorf("theme = %q, want Gruvbox Dark", m.theme.Name)
	}
}

func TestCmdTheme_ShortAlias_Gruvbox(t *testing.T) {
	m := modelWithEpisodes()
	m.cmdTheme([]string{"gruvbox"})
	if m.theme.Name != "Gruvbox Dark" {
		t.Errorf("theme = %q, want Gruvbox Dark", m.theme.Name)
	}
}

func TestCmdTheme_ShortAlias_Everforest(t *testing.T) {
	m := modelWithEpisodes()
	m.cmdTheme([]string{"everforest"})
	if m.theme.Name != "Everforest Dark" {
		t.Errorf("theme = %q, want Everforest Dark", m.theme.Name)
	}
}

func TestCmdTheme_ShortAlias_OneDark(t *testing.T) {
	m := modelWithEpisodes()
	m.cmdTheme([]string{"onedark"})
	if m.theme.Name != "One Dark" {
		t.Errorf("theme = %q, want One Dark", m.theme.Name)
	}
}

func TestCmdTheme_ShortAlias_One(t *testing.T) {
	m := modelWithEpisodes()
	m.cmdTheme([]string{"one"})
	if m.theme.Name != "One Dark" {
		t.Errorf("theme = %q, want One Dark (prefix match)", m.theme.Name)
	}
}

func TestCmdTheme_Unknown(t *testing.T) {
	m := modelWithEpisodes()
	msg := runCmd(t, m.cmdTheme([]string{"solarized"}))
	errMsg, ok := msg.(commandErrMsg)
	if !ok || errMsg.err == nil {
		t.Errorf("expected error for unknown theme, got %T", msg)
	}
	if !strings.Contains(errMsg.err.Error(), "available") {
		t.Errorf("error should list available themes, got: %v", errMsg.err)
	}
}

func TestCmdTheme_ReturnsResultMsg(t *testing.T) {
	m := modelWithEpisodes()
	cmd := m.cmdTheme([]string{"nord"})
	result, ok := findResultMsg(t, cmd)
	if !ok {
		t.Error("expected commandResultMsg in batch from cmdTheme")
	}
	if !strings.Contains(result.output, "Nord") {
		t.Errorf("result.output = %q, want to contain Nord", result.output)
	}
}

// ─── cmdVolume ────────────────────────────────────────────────────────────────

func TestCmdVolume_NoArgs(t *testing.T) {
	m := modelWithEpisodes()
	msg := runCmd(t, m.cmdVolume(nil))
	if _, ok := msg.(commandErrMsg); !ok {
		t.Errorf("expected commandErrMsg for no args, got %T", msg)
	}
}

func TestCmdVolume_Valid(t *testing.T) {
	m := modelWithEpisodes()
	cmd := m.cmdVolume([]string{"80"})
	if m.volume != 80 {
		t.Errorf("volume = %d, want 80", m.volume)
	}
	result, ok := runCmd(t, cmd).(commandResultMsg)
	if !ok {
		t.Fatalf("expected commandResultMsg")
	}
	if !strings.Contains(result.output, "80%") {
		t.Errorf("output = %q, want to contain 80%%", result.output)
	}
}

func TestCmdVolume_Zero(t *testing.T) {
	m := modelWithEpisodes()
	m.cmdVolume([]string{"0"})
	if m.volume != 0 {
		t.Errorf("volume = %d, want 0", m.volume)
	}
}

func TestCmdVolume_Max(t *testing.T) {
	m := modelWithEpisodes()
	m.cmdVolume([]string{"150"})
	if m.volume != 150 {
		t.Errorf("volume = %d, want 150", m.volume)
	}
}

func TestCmdVolume_TooHigh(t *testing.T) {
	m := modelWithEpisodes()
	orig := m.volume
	msg := runCmd(t, m.cmdVolume([]string{"200"}))
	if _, ok := msg.(commandErrMsg); !ok {
		t.Errorf("expected commandErrMsg for vol 200, got %T", msg)
	}
	if m.volume != orig {
		t.Errorf("volume should not change on error, got %d want %d", m.volume, orig)
	}
}

func TestCmdVolume_Negative(t *testing.T) {
	m := modelWithEpisodes()
	msg := runCmd(t, m.cmdVolume([]string{"-10"}))
	if _, ok := msg.(commandErrMsg); !ok {
		t.Errorf("expected commandErrMsg for negative volume, got %T", msg)
	}
}

func TestCmdVolume_NonNumeric(t *testing.T) {
	m := modelWithEpisodes()
	msg := runCmd(t, m.cmdVolume([]string{"loud"}))
	if _, ok := msg.(commandErrMsg); !ok {
		t.Errorf("expected commandErrMsg for non-numeric, got %T", msg)
	}
}

func TestCmdVolume_VolumeAlias(t *testing.T) {
	m := modelWithEpisodes()
	cmd := m.executeCommand("volume 50")
	if m.volume != 50 {
		t.Errorf("volume alias: volume = %d, want 50", m.volume)
	}
	if _, ok := runCmd(t, cmd).(commandResultMsg); !ok {
		t.Error("expected commandResultMsg from volume alias")
	}
}

// ─── cmdFavourite ────────────────────────────────────────────────────────────

func TestCmdFavourite_TogglesOn(t *testing.T) {
	m := modelWithEpisodes()
	m.favourites = make(map[int]bool) // isolate from persisted state
	epNum := m.currentEpisode().Number
	cmd := m.cmdFavourite()
	if !m.favourites[epNum] {
		t.Errorf("expected episode %d to be favourited", epNum)
	}
	result, ok := runCmd(t, cmd).(commandResultMsg)
	if !ok {
		t.Fatalf("expected commandResultMsg")
	}
	if !strings.Contains(result.output, "added to") {
		t.Errorf("output = %q, want 'added to'", result.output)
	}
}

func TestCmdFavourite_TogglesOff(t *testing.T) {
	m := modelWithEpisodes()
	m.favourites = make(map[int]bool) // isolate from persisted state
	epNum := m.currentEpisode().Number
	m.favourites[epNum] = true // pre-set
	cmd := m.cmdFavourite()
	if m.favourites[epNum] {
		t.Errorf("expected episode %d to be unfavourited", epNum)
	}
	result, ok := runCmd(t, cmd).(commandResultMsg)
	if !ok {
		t.Fatalf("expected commandResultMsg")
	}
	if !strings.Contains(result.output, "removed from") {
		t.Errorf("output = %q, want 'removed from'", result.output)
	}
}

func TestCmdFavourite_AllAliases(t *testing.T) {
	aliases := []string{"fav", "favourite", "favorite"}
	for _, alias := range aliases {
		m := modelWithEpisodes()
		cmd := m.executeCommand(alias)
		if cmd == nil {
			t.Errorf("%s: expected non-nil cmd", alias)
		}
		if _, ok := runCmd(t, cmd).(commandResultMsg); !ok {
			t.Errorf("%s: expected commandResultMsg", alias)
		}
	}
}

// ─── cmdRandom ────────────────────────────────────────────────────────────────

func TestCmdRandom_ChangesEpisode(t *testing.T) {
	m := modelWithEpisodes()
	m.currentIdx = 0
	m.cmdRandom()
	if m.currentIdx == 0 {
		t.Error("expected currentIdx to change after cmdRandom (3 episodes)")
	}
}

func TestCmdRandom_ReturnsResultMsg(t *testing.T) {
	m := modelWithEpisodes()
	cmd := m.cmdRandom()
	result, ok := findResultMsg(t, cmd)
	if !ok {
		t.Error("expected commandResultMsg from cmdRandom")
	}
	if result.output == "" {
		t.Error("expected non-empty output from cmdRandom")
	}
}

func TestCmdRandom_SingleEpisode_ReturnsMessage(t *testing.T) {
	m := modelWithEpisodes()
	m.episodes = m.episodes[:1]
	cmd := m.cmdRandom()
	result, ok := runCmd(t, cmd).(commandResultMsg)
	if !ok {
		t.Fatalf("expected commandResultMsg for single-episode case")
	}
	if !strings.Contains(result.output, "not enough") {
		t.Errorf("output = %q, want 'not enough'", result.output)
	}
}

// ─── cmdJump ──────────────────────────────────────────────────────────────────

func TestCmdJump_NoArgs(t *testing.T) {
	m := modelWithEpisodes()
	msg := runCmd(t, m.cmdJump(nil))
	if _, ok := msg.(commandErrMsg); !ok {
		t.Errorf("expected commandErrMsg for no args, got %T", msg)
	}
}

func TestCmdJump_ByNumber(t *testing.T) {
	m := modelWithEpisodes()
	m.currentIdx = 0 // ep 78
	cmd := m.cmdJump([]string{"77"})
	if m.currentIdx != 1 {
		t.Errorf("currentIdx = %d, want 1 (ep 77)", m.currentIdx)
	}
	result, ok := findResultMsg(t, cmd)
	if !ok {
		t.Error("expected commandResultMsg from jump by number")
	}
	if !strings.Contains(result.output, "77") {
		t.Errorf("output = %q, want ep number 77", result.output)
	}
}

func TestCmdJump_NumberNotFound(t *testing.T) {
	m := modelWithEpisodes()
	msg := runCmd(t, m.cmdJump([]string{"999"}))
	errMsg, ok := msg.(commandErrMsg)
	if !ok || errMsg.err == nil {
		t.Errorf("expected commandErrMsg for missing ep 999, got %T", msg)
	}
}

func TestCmdJump_FuzzyTitle(t *testing.T) {
	m := modelWithEpisodes()
	m.currentIdx = 0 // ep 78 Datassette
	cmd := m.cmdJump([]string{"phonaut"})
	if m.currentIdx != 1 {
		t.Errorf("currentIdx = %d, want 1 (Phonaut fuzzy match)", m.currentIdx)
	}
	result, ok := findResultMsg(t, cmd)
	if !ok {
		t.Error("expected commandResultMsg from fuzzy jump")
	}
	if !strings.Contains(strings.ToLower(result.output), "phonaut") {
		t.Errorf("output = %q, want ep title containing 'Phonaut'", result.output)
	}
}

func TestCmdJump_FuzzyNoMatch(t *testing.T) {
	m := modelWithEpisodes()
	msg := runCmd(t, m.cmdJump([]string{"zzzzz"}))
	if _, ok := msg.(commandErrMsg); !ok {
		t.Errorf("expected commandErrMsg for no fuzzy match, got %T", msg)
	}
}

// ─── cmdHelp ─────────────────────────────────────────────────────────────────

func TestCmdHelp_SetsShowHelp(t *testing.T) {
	m := modelWithEpisodes()
	cmd := m.cmdHelp()
	if !m.showHelp {
		t.Error("expected showHelp=true after cmdHelp")
	}
	if cmd != nil {
		t.Error("expected nil cmd from cmdHelp")
	}
}

// ─── getAutocompleteHint ─────────────────────────────────────────────────────

func TestGetAutocompleteHint_EmptyShowsCommands(t *testing.T) {
	m := modelWithEpisodes()
	m.commandMode = true
	hint := m.getAutocompleteHint()
	for _, cmd := range []string{"theme", "jump", "vol", "fav", "random", "quit"} {
		if !strings.Contains(hint, cmd) {
			t.Errorf("empty hint should contain %q, got %q", cmd, hint)
		}
	}
}

func TestGetAutocompleteHint_ThemeNoArgs(t *testing.T) {
	m := modelWithEpisodes()
	m.commandMode = true
	m.commandInput.SetValue("theme")
	hint := m.getAutocompleteHint()
	for _, name := range []string{"dracula", "nord", "gruvbox", "onedark", "everforest"} {
		if !strings.Contains(hint, name) {
			t.Errorf("theme hint should contain %q, got %q", name, hint)
		}
	}
}

func TestGetAutocompleteHint_ThemeGhostText(t *testing.T) {
	m := modelWithEpisodes()
	m.commandMode = true
	m.commandInput.SetValue("theme gru")
	hint := m.getAutocompleteHint()
	if !strings.Contains(hint, "vbox dark") {
		t.Errorf("expected ghost text 'vbox dark' for 'gru', got %q", hint)
	}
}

func TestGetAutocompleteHint_ThemeComplete_NoHint(t *testing.T) {
	m := modelWithEpisodes()
	m.commandMode = true
	m.commandInput.SetValue("theme gruvbox dark")
	hint := m.getAutocompleteHint()
	if hint != "" {
		t.Errorf("expected empty hint for complete theme name, got %q", hint)
	}
}

func TestGetAutocompleteHint_VolAlias(t *testing.T) {
	m := modelWithEpisodes()
	m.commandMode = true
	m.commandInput.SetValue("vol")
	hint := m.getAutocompleteHint()
	if !strings.Contains(hint, "0-150") {
		t.Errorf("vol hint should contain '0-150', got %q", hint)
	}
}

func TestGetAutocompleteHint_VolumeAlias(t *testing.T) {
	m := modelWithEpisodes()
	m.commandMode = true
	m.commandInput.SetValue("volume")
	hint := m.getAutocompleteHint()
	if !strings.Contains(hint, "0-150") {
		t.Errorf("volume hint should contain '0-150', got %q", hint)
	}
}

func TestGetAutocompleteHint_Jump(t *testing.T) {
	m := modelWithEpisodes()
	m.commandMode = true
	m.commandInput.SetValue("jump")
	hint := m.getAutocompleteHint()
	if hint == "" {
		t.Error("expected non-empty hint for jump")
	}
}

func TestGetAutocompleteHint_FavAliases(t *testing.T) {
	m := modelWithEpisodes()
	m.commandMode = true
	for _, alias := range []string{"fav", "favourite", "favorite"} {
		m.commandInput.SetValue(alias)
		hint := m.getAutocompleteHint()
		if hint == "" {
			t.Errorf("%s: expected non-empty hint", alias)
		}
	}
}

func TestGetAutocompleteHint_Random(t *testing.T) {
	m := modelWithEpisodes()
	m.commandMode = true
	m.commandInput.SetValue("random")
	hint := m.getAutocompleteHint()
	if hint == "" {
		t.Error("expected non-empty hint for random")
	}
}

func TestGetAutocompleteHint_Help(t *testing.T) {
	m := modelWithEpisodes()
	m.commandMode = true
	m.commandInput.SetValue("help")
	hint := m.getAutocompleteHint()
	if hint == "" {
		t.Error("expected non-empty hint for help")
	}
}

func TestGetAutocompleteHint_Quit(t *testing.T) {
	m := modelWithEpisodes()
	m.commandMode = true
	m.commandInput.SetValue("quit")
	hint := m.getAutocompleteHint()
	if hint == "" {
		t.Error("expected non-empty hint for quit")
	}
}

// ─── renderCommandBar ────────────────────────────────────────────────────────

func TestRenderCommandBar_PromptVisibleInCommandMode(t *testing.T) {
	m := modelWithEpisodes()
	m.commandMode = true
	m.commandInput.Focus()
	bar := m.renderCommandBar()
	if !strings.Contains(bar, ":") {
		t.Error("expected ':' prompt in command bar")
	}
}

func TestRenderCommandBar_ShowsError(t *testing.T) {
	m := modelWithEpisodes()
	m.commandResult = true
	m.commandError = "unknown command: foo"
	bar := m.renderCommandBar()
	if !strings.Contains(bar, "unknown command: foo") {
		t.Errorf("expected error in bar, got %q", bar)
	}
}

func TestRenderCommandBar_ShowsOutput(t *testing.T) {
	m := modelWithEpisodes()
	m.commandResult = true
	m.commandOutput = "vol: 80%"
	bar := m.renderCommandBar()
	if !strings.Contains(bar, "vol: 80%") {
		t.Errorf("expected output in bar, got %q", bar)
	}
}

func TestView_CommandBarVisibleInCommandMode(t *testing.T) {
	m := modelWithEpisodes()
	m.commandMode = true
	m.commandInput.Focus()
	view := m.View()
	if !strings.Contains(view, ":") {
		t.Error("expected command bar in view when commandMode=true")
	}
}

func TestView_CommandBarVisibleInResultMode(t *testing.T) {
	m := modelWithEpisodes()
	m.commandResult = true
	m.commandOutput = "theme: Dracula"
	view := m.View()
	if !strings.Contains(view, "theme: Dracula") {
		t.Errorf("expected result in view when commandResult=true, got: %q", view)
	}
}

func TestView_CommandBarHiddenInNormalMode(t *testing.T) {
	m := modelWithEpisodes()
	m.commandMode = false
	m.commandResult = false
	m.commandOutput = "should-not-appear-xyz"
	view := m.View()
	if strings.Contains(view, "should-not-appear-xyz") {
		t.Error("command output should be hidden when commandMode=false and commandResult=false")
	}
}

// ─── Themes: CommandBarBg ─────────────────────────────────────────────────────

func TestAllThemes_HaveCommandBarBg(t *testing.T) {
	for _, theme := range Themes {
		if theme.CommandBarBg == "" {
			t.Errorf("theme %q is missing CommandBarBg", theme.Name)
		}
		if theme.CommandBarHint == "" {
			t.Errorf("theme %q is missing CommandBarHint", theme.Name)
		}
	}
}

func TestThemeNord_CommandBarBg(t *testing.T) {
	if ThemeNord.CommandBarBg == "" {
		t.Error("ThemeNord.CommandBarBg must not be empty")
	}
}

// ─── fuzzyMatch ──────────────────────────────────────────────────────────────

func TestFuzzyMatch_ExactMatch(t *testing.T) {
	if s := fuzzyMatch("hello", "hello"); s != 1.0 {
		t.Errorf("exact match score = %f, want 1.0", s)
	}
}

func TestFuzzyMatch_NoMatch(t *testing.T) {
	if s := fuzzyMatch("abc", "xyz"); s != 0.0 {
		t.Errorf("no-match score = %f, want 0.0", s)
	}
}

func TestFuzzyMatch_EmptyQuery(t *testing.T) {
	if s := fuzzyMatch("abc", ""); s != 0.0 {
		t.Errorf("empty query score = %f, want 0.0", s)
	}
}

func TestFuzzyMatch_SubsequenceMatch(t *testing.T) {
	// "dat" should match "datassette" well
	if s := fuzzyMatch("datassette", "dat"); s < 0.9 {
		t.Errorf("subsequence score = %f, want >= 0.9", s)
	}
}

// ─── reset ───────────────────────────────────────────────────────────────────

func TestExecuteCommand_Reset(t *testing.T) {
	m := modelWithEpisodes()
	cmd := m.executeCommand("reset")
	if cmd == nil {
		t.Fatal("expected non-nil cmd from reset command")
	}
	if !m.loading {
		t.Error("expected loading=true after reset")
	}
	if m.pl != nil {
		t.Error("expected pl=nil after reset")
	}
	if m.playerReady {
		t.Error("expected playerReady=false after reset")
	}
	if m.err != nil {
		t.Errorf("expected err=nil after reset, got %v", m.err)
	}
}

func TestResetPlayer_ClearsState(t *testing.T) {
	m := modelWithEpisodes()
	m.pl = nil // no real player in tests
	m.state = player.State{Loaded: true, Paused: false, Position: 120.0}
	m.err = errMpvNotFound
	m.pendingResume = 55.0

	m.resetPlayer()

	if m.pl != nil {
		t.Error("expected pl=nil after resetPlayer")
	}
	if m.playerReady {
		t.Error("expected playerReady=false after resetPlayer")
	}
	if m.state.Loaded {
		t.Error("expected state. Loaded=false after resetPlayer")
	}
	if m.state.Position != 0 {
		t.Errorf("expected state.Position=0, got %f", m.state.Position)
	}
	if m.err != nil {
		t.Errorf("expected err=nil after resetPlayer, got %v", m.err)
	}
	if m.pendingResume != 0 {
		t.Errorf("expected pendingResume=0, got %f", m.pendingResume)
	}
	if !m.loading {
		t.Error("expected loading=true after resetPlayer")
	}
}

func TestResetPlayer_ReturnsNonNilCmd(t *testing.T) {
	m := modelWithEpisodes()
	m.pl = nil
	cmd := m.resetPlayer()
	if cmd == nil {
		t.Error("expected non-nil cmd from resetPlayer")
	}
	// Don't execute it — would spawn real mpv
}

func TestResetPlayerCmd_NilPlayer(t *testing.T) {
	// Passing nil to resetPlayerCmd should not panic
	cmd := resetPlayerCmd(nil)
	if cmd == nil {
		t.Fatal("expected non-nil cmd from resetPlayerCmd")
	}
	// Execute it — will try to spawn mpv, but in test env mpv may not exist.
	// We check it doesn't panic; the result is a playerErrMsg if mpv is missing.
	msg := cmd()
	if _, ok := msg.(playerReadyMsg); ok {
		// mpv exists — that's fine
	} else if errMsg, ok := msg.(playerErrMsg); ok {
		// mpv not found — expected in test env
		if errMsg.err == nil {
			t.Error("expected non-nil error in playerErrMsg when mpv is missing")
		}
	} else {
		t.Errorf("expected playerReadyMsg or playerErrMsg, got %T", msg)
	}
}

func TestGetAutocompleteHint_Reset(t *testing.T) {
	m := modelWithEpisodes()
	m.commandMode = true
	m.commandInput.SetValue("reset")
	hint := m.getAutocompleteHint()
	if hint == "" {
		t.Error("expected non-empty hint for reset")
	}
	if !strings.Contains(hint, "kill") {
		t.Errorf("expected reset hint to mention killing, got %q", hint)
	}
}

func TestGetAutocompleteHint_EmptyContainsReset(t *testing.T) {
	m := modelWithEpisodes()
	m.commandMode = true
	m.commandInput.SetValue("")
	hint := m.getAutocompleteHint()
	if !strings.Contains(hint, "reset") {
		t.Errorf("empty hint should contain 'reset', got %q", hint)
	}
}
