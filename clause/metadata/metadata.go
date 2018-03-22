package metadata

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
	typeKey = "metadata"
)

var (
	documentation = clause.ClauseDocumentation{
		Summary: "Searches based on the metadata associated with an object. At least one of attribute, value, or unit should be non-blank.",
		Args: map[string]clause.ClauseArgumentDocumentation{
			"attribute":       {Type: "string", Summary: "The AVU's attribute field"},
			"value":           {Type: "string", Summary: "The AVU's value field"},
			"unit":            {Type: "string", Summary: "The AVU's unit field"},
			"metadata_types":  {Type: "[]string", Summary: "What types of metadata to search. Can include 'irods', 'cyverse', or blank for both types."},
			"attribute_exact": {Type: "bool", Summary: "Whether to search the attribute exactly, or add implicit wildcards"},
			"value_exact":     {Type: "bool", Summary: "Whether to search the value exactly, or add implicit wildcards"},
			"unit_exact":      {Type: "bool", Summary: "Whether to search the unit exactly, or add implicit wildcards"},
		},
	}
)

type MetadataArgs struct {
	Attribute      string
	Value          string
	Unit           string
	MetadataTypes  []string `mapstructure:"metadata_types"`
	AttributeExact bool     `mapstructure:"attribute_exact"`
	ValueExact     bool     `mapstructure:"value_exact"`
	UnitExact      bool     `mapstructure:"unit_exact"`
}

func makeNested(attr string, value string, unit string) elastic.Query {
	inner := elastic.NewBoolQuery()
	if attr != "" {
		inner.Must(elastic.NewQueryStringQuery(attr).Field("metadata.attribute"))
	}
	if value != "" {
		inner.Must(elastic.NewQueryStringQuery(value).Field("metadata.value"))
	}
	if unit != "" {
		inner.Must(elastic.NewQueryStringQuery(unit).Field("metadata.unit"))
	}
	return elastic.NewNestedQuery("metadata", inner)
}

func MetadataProcessor(_ context.Context, args map[string]interface{}) (elastic.Query, error) {
	var realArgs MetadataArgs
	err := mapstructure.Decode(args, &realArgs)
	if err != nil {
		return nil, err
	}

	if realArgs.Attribute == "" && realArgs.Value == "" && realArgs.Unit == "" {
		return nil, errors.New("Must provide at least one of attribute, value, or unit")
	}

	var includeIrods, includeCyverse bool
	if len(realArgs.MetadataTypes) == 0 {
		includeIrods = true
		includeCyverse = true
	} else {
		for _, t := range realArgs.MetadataTypes {
			if t == "irods" {
				includeIrods = true
			} else if t == "cyverse" {
				includeCyverse = true
			} else {
				return nil, fmt.Errorf("Got a metadata type of %q, but expected irods or cyverse", t)
			}
		}
	}

	finalq := elastic.NewBoolQuery()

	var attr, value, unit string
	if realArgs.AttributeExact {
		attr = realArgs.Attribute
	} else {
		attr = clauseutils.AddImplicitWildcard(realArgs.Attribute)
	}
	if realArgs.ValueExact {
		value = realArgs.Value
	} else {
		value = clauseutils.AddImplicitWildcard(realArgs.Value)
	}
	if realArgs.UnitExact {
		unit = realArgs.Unit
	} else {
		unit = clauseutils.AddImplicitWildcard(realArgs.Unit)
	}

	if includeIrods {
		finalq.Should(makeNested(attr, value, unit))
	}
	if includeCyverse {
		finalq.Should(elastic.NewHasChildQuery("file_metadata", makeNested(attr, value, unit)).ScoreMode("max").InnerHit(elastic.NewInnerHit()))
		finalq.Should(elastic.NewHasChildQuery("folder_metadata", makeNested(attr, value, unit)).ScoreMode("max").InnerHit(elastic.NewInnerHit()))
	}

	return finalq, nil
}

func Register(qd *querydsl.QueryDSL) {
	qd.AddClauseType(typeKey, MetadataProcessor, documentation)
}
