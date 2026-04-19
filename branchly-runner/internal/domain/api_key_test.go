package domain

import "testing"

func TestAgentType_IsValid_GPTCodex(t *testing.T) {
	if !AgentTypeGPTCodex.IsValid() {
		t.Error("AgentTypeGPTCodex.IsValid() = false, want true")
	}
}

func TestAgentType_IsValid_AllTypes(t *testing.T) {
	for _, tt := range []struct {
		agent AgentType
		want  bool
	}{
		{AgentTypeClaudeCode, true},
		{AgentTypeGemini, true},
		{AgentTypeGPTCodex, true},
		{"unknown", false},
		{"", false},
	} {
		if got := tt.agent.IsValid(); got != tt.want {
			t.Errorf("AgentType(%q).IsValid() = %v, want %v", tt.agent, got, tt.want)
		}
	}
}

func TestRequiredKeyProvider_GPTCodex_ReturnsOpenAI(t *testing.T) {
	got := RequiredKeyProvider(AgentTypeGPTCodex)
	if got != APIKeyProviderOpenAI {
		t.Errorf("RequiredKeyProvider(gpt-codex) = %q, want %q", got, APIKeyProviderOpenAI)
	}
}

func TestRequiredKeyProvider_AllAgents(t *testing.T) {
	cases := []struct {
		agent    AgentType
		provider APIKeyProvider
	}{
		{AgentTypeClaudeCode, APIKeyProviderAnthropic},
		{AgentTypeGemini, APIKeyProviderGoogle},
		{AgentTypeGPTCodex, APIKeyProviderOpenAI},
	}
	for _, tt := range cases {
		got := RequiredKeyProvider(tt.agent)
		if got != tt.provider {
			t.Errorf("RequiredKeyProvider(%q) = %q, want %q", tt.agent, got, tt.provider)
		}
	}
}
