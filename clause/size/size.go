package size

import (
	"context"
	"errors"
	"fmt"

	"github.com/cyverse-de/querydsl/v2"
	"github.com/cyverse-de/querydsl/v2/clause"
	"github.com/cyverse-de/querydsl/v2/clauseutils"
	"github.com/mitchellh/mapstructure"
	"github.com/olivere/elastic/v7"
)

const (
	typeKey = "size"
)

var (
	documentation = clause.ClauseDocumentation{
		Summary: "Searches based on an object's file size. Searches matching this clause will only include files, as folders do not store a size.",
		Args: map[string]clause.ClauseArgumentDocumentation{
			"from": {Type: "string", Summary: "The lower end of the range (inclusive). Pass as a string, either a number of bytes or a number followed by optional whitespace and then one of 'KB', 'MB', 'GB', or 'TB', which refer to powers of 1024 bytes (commonly called kilo/mebi/gibi/tebibytes)."},
			"to":   {Type: "string", Summary: "The upper end of the range (inclusive). Pass as a string, as with 'from'."},
		},
	}
)

type SizeArgs struct {
	From string
	To   string
}

func SizeProcessor(_ context.Context, args map[string]interface{}) (elastic.Query, error) {
	var realArgs SizeArgs
	err := mapstructure.Decode(args, &realArgs)
	if err != nil {
		return nil, err
	}

	if realArgs.From == "" && realArgs.To == "" {
		return nil, errors.New("Neither from nor to was passed, cannot create clause.")
	}

	var from, to int64
	var rangetype clauseutils.RangeType

	if realArgs.From != "" {
		rangetype = clauseutils.LowerOnly
		from, err = clauseutils.StringToFilesize(realArgs.From)
		if err != nil {
			return nil, err
		}
	}

	if realArgs.To != "" {
		if rangetype == clauseutils.LowerOnly {
			rangetype = clauseutils.Both
		} else {
			rangetype = clauseutils.UpperOnly
		}
		to, err = clauseutils.StringToFilesize(realArgs.To)
		if err != nil {
			return nil, err
		}
	}

	return clauseutils.CreateRangeQuery("fileSize", rangetype, from, to), nil
}

func SizeSummary(_ context.Context, args map[string]interface{}) (string, error) {
	var realArgs SizeArgs
	err := mapstructure.Decode(args, &realArgs)
	if err != nil {
		return "", err
	}

	if realArgs.From == "" && realArgs.To == "" {
		return "", errors.New("Neither from nor to was passed, cannot create clause.")
	}

	return fmt.Sprintf("size=%s--%s", realArgs.From, realArgs.To), nil
}

func Register(qd *querydsl.QueryDSL) {
	qd.AddClauseTypeSummarized(typeKey, SizeProcessor, documentation, SizeSummary)
}
