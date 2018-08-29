package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"encoding/json"
	"github.com/graphql-go/graphql"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"crypto/sha256"
)

func main() {
	var rootCmd = &cobra.Command{
		Short: "json filter",
		Long:  `jf is a tool to pull out a subset of a json document in a standard, hashable way.`,
		Use:   "jf [flags] [file] [-]",
		DisableFlagsInUseLine: true,

		PreRunE: func(cmd *cobra.Command, args []string) error {
			if debug, _ := cmd.Flags().GetBool("debug"); debug {
				logrus.SetLevel(logrus.DebugLevel)
			}
			return nil
		},

		RunE: runCmdFunc,
	}

	rootCmd.Flags().Bool("debug", false, "enable debug logging")
	rootCmd.Flags().StringP("query", "q", "", "query string in graphql format")
	rootCmd.Flags().StringP("file", "f", "", "file containing query string in graphql format")
	rootCmd.Flags().BoolP("hash", "x", false, "output sha256 instead of filtered content")
	//rootCmd.Flags().BoolP("pretty-output", "p", true, "pretty-printed output")

	rootCmd.Flags().MarkHidden("debug")

	rootCmd.Execute()
}

func runCmdFunc(cmd *cobra.Command, args []string) error {
	query, _ := cmd.Flags().GetString("query")
	if query == "" {
		queryFile, _ := cmd.Flags().GetString("file")
		queryBytes, err := ioutil.ReadFile(queryFile)
		if err != nil {
			return fmt.Errorf("failed to read file at %s: `%s`", queryFile, err)
		}
		query = string(queryBytes)
	}

	path := os.ExpandEnv(args[0])
	if path == "-" {
		path="/dev/stdin"
	}
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file at %s: `%s`", path, err)
	}
	shouldHash, _ := cmd.Flags().GetBool("hash")
	if shouldHash {
		fmt.Printf("%x", Sha256Json(query, json.RawMessage(fileBytes)))
	} else {
		fmt.Print(string(FilterJson(query, json.RawMessage(fileBytes))))
	}
	return nil
}

func Sha256Json(query string, rawJson json.RawMessage) [sha256.Size]byte {
	return sha256.Sum256([]byte(FilterJson(query, rawJson)))
}

func FilterJson(query string, rawJson json.RawMessage) json.RawMessage {
	schema, err := QueryableSchemaFromJson(rawJson)
	if err != nil {
		logrus.Fatalf("failed to create new schema, error: %v", err)
	}

	params := graphql.Params{Schema: *schema, RequestString: query}
	r := graphql.Do(params)
	if len(r.Errors) > 0 {
		logrus.Fatalf("failed to execute graphql operation, errors: %+v", r.Errors)
	}

	rJSON, _ := json.Marshal(r.Data)
	return rJSON
}

func QueryableSchemaFromJson(rawJson json.RawMessage) (*graphql.Schema, error) {
	parsed := map[string]interface{}{}
	if err := json.Unmarshal(rawJson, &parsed); err != nil {
		return nil, err
	}
	schemaConfig := graphql.SchemaConfig{Query: ObjectFor("root", parsed)}
	scheme, err := graphql.NewSchema(schemaConfig)
	return &scheme, err
}

func ObjectFor(name string, parsed map[string]interface{}) *graphql.Object {
	fields := graphql.Fields{}
	for key, value := range parsed {
		logrus.Debugf("field %s, %#v, %T", key, value, value)
		fieldType := FieldTypeForField(key, value)
		fieldResolver := ResolverForField(key, value)
		if fieldType != nil {
			fields[key] = &graphql.Field{
				Type:    fieldType,
				Resolve: fieldResolver,
			}
		}
	}
	return graphql.NewObject(graphql.ObjectConfig{Name: name, Fields: fields})
}

func FieldTypeForField(name string, value interface{}) graphql.Output {
	switch value.(type) {
	case string:
		return graphql.String
	case float64:
		return graphql.Float
	case bool:
		return graphql.Boolean
	case []interface{}:
		list := value.([]interface{})
		if len(list) > 0 {
			return graphql.NewList(FieldTypeForField(name, list[0]))
		}
	case map[string]interface{}:
		return ObjectFor(name, value.(map[string]interface{}))
	}
	return nil
}

func ResolverForField(name string, value interface{}) graphql.FieldResolveFn {
	switch value.(type) {
	case []interface{}:
		list := value.([]interface{})
		return func(p graphql.ResolveParams) (interface{}, error) {
			vals := []interface{}{}
			for _, obj := range list {
				resolver := ResolverForField(name, obj)
				val, err := resolver(p)
				if err != nil {
					return nil, err
				}
				vals = append(vals, val)
			}
			return vals, nil
		}
	}
	return func(p graphql.ResolveParams) (interface{}, error) {
		return value, nil
	}
}
