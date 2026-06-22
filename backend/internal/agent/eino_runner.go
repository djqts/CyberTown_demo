package agent

import (
	"context"
	"fmt"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"backend/internal/memory"
)

type EinoRunner struct {
	runnable compose.Runnable[PromptInput, *schema.Message]
}

func NewEinoRunner(ctx context.Context, chatModel einomodel.BaseChatModel, memSvc *memory.Service) (*EinoRunner, error) {
	memoryRecall := compose.InvokableLambda(func(ctx context.Context, input PromptInput) (PromptInput, error) {
		if memSvc != nil && input.MemoryContext == "" && input.NPC != nil {
			input.MemoryContext = memSvc.Recall(ctx, input.NPC.ID, input.UserToken, input.UserInput)
		}
		return input, nil
	}, compose.WithLambdaType("NPCMemoryRecall"))

	contextMapper := compose.InvokableLambda(func(_ context.Context, input PromptInput) (map[string]any, error) {
		return BuildPromptValues(input), nil
	}, compose.WithLambdaType("NPCPromptContextMapper"))

	chatTemplate := prompt.FromMessages(schema.FString,
		schema.SystemMessage(SystemPromptTemplate),
		schema.MessagesPlaceholder("history", true),
		schema.UserMessage("{user_input}"),
	)

	runnable, err := compose.NewChain[PromptInput, *schema.Message]().
		AppendLambda(memoryRecall, compose.WithNodeName("npc_memory_recall")).
		AppendLambda(contextMapper, compose.WithNodeName("npc_context_mapper")).
		AppendChatTemplate(chatTemplate, compose.WithNodeName("npc_prompt")).
		AppendChatModel(chatModel, compose.WithNodeName("npc_chat_model")).
		Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("compile npc agent chain: %w", err)
	}

	return &EinoRunner{runnable: runnable}, nil
}

func (r *EinoRunner) Run(ctx context.Context, input PromptInput) (*schema.Message, error) {
	if r == nil || r.runnable == nil {
		return nil, fmt.Errorf("eino runner is not initialized")
	}

	resp, err := r.runnable.Invoke(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("invoke npc agent chain: %w", err)
	}
	return resp, nil
}
