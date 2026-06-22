package memory

import (
	"context"
	"fmt"

	qdrant "github.com/qdrant/go-client/qdrant"
)

// InitCollections 初始化 Qdrant collections（npc_memory, world_knowledge）。
func InitCollections(ctx context.Context, client *qdrant.Client) error {
	for _, spec := range []struct{ name string }{
		{name: "npc_memory"},
		{name: "world_knowledge"},
	} {
		exists, err := client.CollectionExists(ctx, spec.name)
		if err != nil {
			return fmt.Errorf("check collection %s: %w", spec.name, err)
		}
		if exists {
			continue
		}

		err = client.CreateCollection(ctx, &qdrant.CreateCollection{
			CollectionName: spec.name,
			VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
				Size:     EmbedDim,
				Distance: qdrant.Distance_Cosine,
			}),
		})
		if err != nil {
			return fmt.Errorf("create collection %s: %w", spec.name, err)
		}
	}
	return nil
}
