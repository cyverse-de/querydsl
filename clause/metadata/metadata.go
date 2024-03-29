package metadata

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cyverse-de/querydsl/v2"
	"github.com/cyverse-de/querydsl/v2/clause"
	"github.com/cyverse-de/querydsl/v2/clauseutils"
	"github.com/mitchellh/mapstructure"
	"github.com/olivere/elastic/v7"
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

func makeNested(suffix, attr, value, unit string) elastic.Query {
	inner := elastic.NewBoolQuery()
	if attr != "" {
		inner.Must(elastic.NewQueryStringQuery(attr).Field(fmt.Sprintf("metadata.%s.attribute", suffix)))
	}
	if value != "" {
		inner.Must(elastic.NewQueryStringQuery(value).Field(fmt.Sprintf("metadata.%s.value", suffix)))
	}
	if unit != "" {
		inner.Must(elastic.NewQueryStringQuery(unit).Field(fmt.Sprintf("metadata.%s.unit", suffix)))
	}
	return elastic.NewNestedQuery(fmt.Sprintf("metadata.%s", suffix), inner)
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
		finalq.Should(makeNested("irods", attr, value, unit))
	}
	if includeCyverse {
		finalq.Should(makeNested("cyverse", attr, value, unit))
	}

	return finalq, nil
}

func MetadataSummary(_ context.Context, args map[string]interface{}) (string, error) {
	var realArgs MetadataArgs
	err := mapstructure.Decode(args, &realArgs)
	if err != nil {
		return "", err
	}

	if realArgs.Attribute == "" && realArgs.Value == "" && realArgs.Unit == "" {
		return "", errors.New("Must provide at least one of attribute, value, or unit")
	}

	var a, v, u string
	if realArgs.Attribute != "" {
		if realArgs.AttributeExact {
			a = fmt.Sprintf("attr=\"%s\"", realArgs.Attribute)
		} else {
			a = fmt.Sprintf("attr~\"%s\"", realArgs.Attribute)
		}
	}
	if realArgs.Value != "" {
		if realArgs.ValueExact {
			v = fmt.Sprintf("value=\"%s\"", realArgs.Value)
		} else {
			v = fmt.Sprintf("value~\"%s\"", realArgs.Value)
		}
	}
	if realArgs.Unit != "" {
		if realArgs.UnitExact {
			u = fmt.Sprintf("unit=\"%s\"", realArgs.Unit)
		} else {
			u = fmt.Sprintf("unit~\"%s\"", realArgs.Unit)
		}
	}
	avu := strings.Join([]string{a, v, u}, ",")
	types := strings.Join(realArgs.MetadataTypes, ",")
	return fmt.Sprintf("metadata=(%s)(%s)", avu, types), nil
}

func Register(qd *querydsl.QueryDSL) {
	qd.AddClauseTypeSummarized(typeKey, MetadataProcessor, documentation, MetadataSummary)
}
