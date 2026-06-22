package agent

import (
	"context"
	"strings"
	"testing"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"backend/internal/model"
)

type fakeChatModel struct {
	messages []*schema.Message
}

func (m *fakeChatModel) Generate(_ context.Context, input []*schema.Message, _ ...einomodel.Option) (*schema.Message, error) {
	m.messages = input
	return schema.AssistantMessage("ok", nil), nil
}

func (m *fakeChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...einomodel.Option) (*schema.StreamReader[*schema.Message], error) {
	msg, err := m.Generate(ctx, input, opts...)
	if err != nil {
		return nil, err
	}
	return schema.StreamReaderFromArray([]*schema.Message{msg}), nil
}

func TestEinoRunnerUsesTemplateHistoryAndChatModel(t *testing.T) {
	ctx := context.Background()
	modelStub := &fakeChatModel{}
	runner, err := NewEinoRunner(ctx, modelStub, nil)
	if err != nil {
		t.Fatalf("NewEinoRunner returned error: %v", err)
	}

	resp, err := runner.Run(ctx, PromptInput{
		NPC: &model.NPC{
			Name:        "Lina",
			Role:        "barista",
			Personality: "warm",
			Status:      "working",
			CurrentGoal: "serve coffee",
			Location:    model.Location{Name: "Cafe"},
		},
		UserInput:     "Any news?",
		MemoryContext: "The clock tower rang twice.",
		History: []model.ChatMessage{
			{Role: "user", Content: "Hello"},
			{Role: "npc", Content: "Good morning"},
		},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if resp.Content != "ok" {
		t.Fatalf("response = %q, want ok", resp.Content)
	}

	if len(modelStub.messages) != 4 {
		t.Fatalf("model received %d messages, want 4", len(modelStub.messages))
	}
	if modelStub.messages[0].Role != schema.System {
		t.Fatalf("first message role = %s, want system", modelStub.messages[0].Role)
	}
	system := modelStub.messages[0].Content
	for _, want := range []string{"Lina", "barista", "Cafe", "The clock tower rang twice."} {
		if !strings.Contains(system, want) {
			t.Fatalf("system prompt missing %q: %s", want, system)
		}
	}
	if modelStub.messages[1].Role != schema.User || modelStub.messages[1].Content != "Hello" {
		t.Fatalf("history user message = %+v", modelStub.messages[1])
	}
	if modelStub.messages[2].Role != schema.Assistant || modelStub.messages[2].Content != "Good morning" {
		t.Fatalf("history npc message = %+v", modelStub.messages[2])
	}
	if modelStub.messages[3].Role != schema.User || modelStub.messages[3].Content != "Any news?" {
		t.Fatalf("current user message = %+v", modelStub.messages[3])
	}
}
