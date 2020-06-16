package serialize

import (
	"context"

	"github.com/puppetlabs/nebula-tasks/pkg/expr/evaluate"
	"github.com/puppetlabs/nebula-tasks/pkg/expr/fn"
	"github.com/puppetlabs/nebula-tasks/pkg/expr/parse"
	"github.com/puppetlabs/nebula-tasks/pkg/expr/resolve"
)

type JSONTree struct {
	parse.Tree
}

func (jt *JSONTree) UnmarshalJSON(data []byte) error {
	tree, err := parse.ParseJSONString(string(data))
	if err != nil {
		return err
	}

	*jt = JSONTree{Tree: tree}
	return nil
}

func (jt JSONTree) MarshalJSON() ([]byte, error) {
	r, err := evaluate.NewEvaluator(
		evaluate.WithInvocationResolver(resolve.NewMemoryInvocationResolver(fn.NewMap(map[string]fn.Descriptor{}))),
		evaluate.WithResultMapper(evaluate.NewJSONResultMapper()),
	).EvaluateAll(context.Background(), jt.Tree)
	if err != nil {
		return nil, err
	}

	return r.Value.([]byte), nil
}