package agent

import (
	"context"

	"github.com/cloudwego/eino/schema"
)

// EinoRunner Eino 编排运行器。
type EinoRunner struct {
	chatModel *ChatModel
}

// NewEinoRunner 创建 Eino 编排运行器。
func NewEinoRunner(chatModel *ChatModel) *EinoRunner {
	return &EinoRunner{chatModel: chatModel}
}

// Run 执行对话生成，传入消息列表，返回模型回复。
func (r *EinoRunner) Run(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	return r.chatModel.Generate(ctx, messages)
}
