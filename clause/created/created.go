package created

import (
	"context"
	"errors"
	"fmt"

	"github.com/cyverse-de/querydsl"
	"github.com/cyverse-de/querydsl/clause"
	"github.com/cyverse-de/querydsl/clauseutils"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/olivere/elastic.v5"
)

const (
	typeKey = "created"
)

var (
	documentation = clause.ClauseDocumentation{
		Summary: "Searches based on an object's creation date",
		Args: map[string]clause.ClauseArgumentDocumentation{
			"from": {Type: "string", Summary: "The start date for the range (inclusive). Pass as a string, milliseconds since epoch or in YYYY-MM-DDTHH:MM:SS.mss<TZ> format, where TZ can either be 'Z' or an offset in ±hh:mm format, or YYYY-MM-DD which assumes UTC and 0 values for all other fields."},
			"to":   {Type: "string", Summary: "The end date for the range (inclusive). Pass as a string, milliseconds since epoch or in YYYY-MM-DDTHH:MM:SS.mss<TZ> format, where TZ can either be 'Z' or an offset in ±hh:mm format, or YYYY-MM-DD which assumes UTC and 0 values for all other fields."},
		},
	}
)

type CreatedArgs struct {
	From string
	To   string
}

func CreatedProcessor(_ context.Context, args map[string]interface{}) (elastic.Query, error) {
	var realArgs CreatedArgs
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
		from, err = clauseutils.DateToEpochMs(realArgs.From)
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
		to, err = clauseutils.DateToEpochMs(realArgs.To)
		if err != nil {
			return nil, err
		}
	}

	return clauseutils.CreateRangeQuery("dateCreated", rangetype, from, to), nil
}

func CreatedSummary(_ context.Context, args map[string]interface{}) (string, error) {
	var realArgs CreatedArgs
	err := mapstructure.Decode(args, &realArgs)
	if err != nil {
		return "", err
	}

	if realArgs.From == "" && realArgs.To == "" {
		return "", errors.New("Neither from nor to was passed, cannot create clause.")
	}

	return fmt.Sprintf("created=%s--%s", realArgs.From, realArgs.To), nil
}

func Register(qd *querydsl.QueryDSL) {
	qd.AddClauseTypeSummarized(typeKey, CreatedProcessor, documentation, CreatedSummary)
}
